package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/bbeni/sphugo/sim"
	"github.com/bbeni/tomato"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/math/fixed"
)

// TODO: remove later
var _ = fmt.Println
var _ = log.Print
var _ = os.Exit
var _ = strings.Split

const (
	PANEL_W    = 445
	RENDERER_W = 1280
	RENDERER_H = 720

	TEXT_SIZE  = 24
	BTN_H      = 56
	MARGIN_BOT = 4

	TERM_TEXT_SIZE = 18

	OPTION_H   = 64
	OPTION_PAD = 4

	SEEKER_H     = 24
	SEEKER_W     = 36
	SEEKER_MIN_W = 8
	SEEKER_PAD   = 4

	BOT_PANEL_H = 220
)

type Theme struct {
	Background color.RGBA

	ButtonTheme        tomato.ButtonColorTheme
	ButtonChooserTheme tomato.ButtonColorTheme

	TermBackground color.RGBA
	ErrorRed       color.RGBA
	SuccessGreen   color.RGBA

	SeekerBackground color.RGBA
	SeekerFrame      color.RGBA
	SeekerCursor     color.RGBA

	ProfilerForeground color.RGBA
	ProfilerBackground color.RGBA
	ProfilerCursor     color.RGBA

	MainFont font.Face
	TermFont font.Face
}

// slowly migrating to this
type SimviewerState struct {
	CursorPos           int
	AnimationRunning    bool
	ConfigChooserOpened bool
	CurrentFrame        image.Image
	TermMsg             string
}

var svState SimviewerState
var dataViewer DataViewInfo

func run() {

	W := RENDERER_W + PANEL_W
	H := RENDERER_H + SEEKER_H + BOT_PANEL_H

	//var fontMu sync.Mutex
	var fontFace font.Face
	var fontFaceTerm font.Face
	{
		font, err := truetype.Parse(gomono.TTF)
		if err != nil {
			panic(err)
		}

		fontFace = truetype.NewFace(font, &truetype.Options{Size: TEXT_SIZE})
		fontFaceTerm = truetype.NewFace(font, &truetype.Options{Size: TERM_TEXT_SIZE})
	}

	colorTheme := &Theme{
		Background: color.RGBA{24, 24, 24, 255},

		ButtonTheme: tomato.ButtonColorTheme{
			Text:     color.RGBA{255, 250, 240, 255}, // Floral White
			BgUp:     color.RGBA{36, 33, 36, 255},    // Raisin Black
			BgHover:  color.RGBA{45, 45, 45, 255},
			FontFace: fontFace,
		},
		ButtonChooserTheme: tomato.ButtonColorTheme{
			Text:     color.RGBA{36, 33, 36, 255},    // Raisin Black
			BgUp:     color.RGBA{255, 250, 240, 255}, // Floral White
			BgHover:  color.RGBA{240, 220, 200, 255},
			FontFace: fontFace,
		},

		//ButtonBlink: color.RGBA{70, 70, 70, 255},

		TermBackground: color.RGBA{36, 23, 36, 255},
		ErrorRed:       color.RGBA{220, 60, 50, 255},
		SuccessGreen:   color.RGBA{60, 220, 23, 255},

		SeekerFrame:      color.RGBA{140, 0, 0, 255},
		SeekerBackground: color.RGBA{120, 0, 30, 255},
		SeekerCursor:     color.RGBA{220, 20, 50, 255},

		ProfilerForeground: color.RGBA{10, 180, 20, 255},
		ProfilerBackground: color.RGBA{24, 24, 24, 255},
		ProfilerCursor:     color.RGBA{48, 48, 48, 255},

		MainFont: fontFace,
		TermFont: fontFaceTerm,
	}

	if err := tomato.Create(W, H, "SFUGO - Simulation Renderer"); err != nil {
		panic(err)
	}

	simulationToggle := make(chan bool) // if false is sent it turns it off

	// create example config file if not existent
	exampleConfigFilePaths := [2]string{"example.sph-config", "tube.sph-config"}
	sim.GenerateDefaultConfigFiles(exampleConfigFilePaths)

	err, simulation := sim.MakeSimulationFromConfig(exampleConfigFilePaths[0])
	if err != nil {
		svState.TermMsg = fmt.Sprintf("%v", err)
	} else {
		svState.TermMsg = fmt.Sprintf("!loaded `%v` sucessfully  (Hint: config files are in the same directory as this program!) ", exampleConfigFilePaths[0])
	}

	animator := sim.MakeAnimator(&simulation)

	go Simulator(simulationToggle, &simulation, &animator)

	tick := 0
	eventsThisTick := make([]tomato.Ev, 0)
	previousTime := time.Now()
	previousTimeActually := time.Now()

	svState.CurrentFrame = animator.Frames[svState.CursorPos]
	configFiles := ListAvailableConfigFiles(".")

	tomato.SetupUi()
	for tomato.Alive() {

		//fmt.Println(tick)
		eventsThisTick = eventsThisTick[:0]

	events_loop:
		for {
			select {
			case event := <-tomato.Events():
				switch event.Kind {
				case tomato.KeyDown:
					if event.Key == tomato.Escape {
						tomato.Die()
					}
					if event.Key == tomato.Space {
						svState.AnimationRunning = !svState.AnimationRunning
					}
				case tomato.MouDown,
					tomato.MouUp,
					tomato.MouMove:
					eventsThisTick = append(eventsThisTick, event)
				}
			default:
				break events_loop
			}
		}

		{
			tomato.Layout(0, tomato.Vertical, image.Rect(W-PANEL_W, 0, W, RENDERER_H))

			if tomato.TextButton(0, "Run/Stop Simulation", &colorTheme.ButtonTheme) {
				// race occurs here, for now we just async call the toggle to not block here
				// it feels kinda slower now ... but not so sure
				go func() { simulationToggle <- true }()
				//simulationToggle <- true
			}

			if tomato.TextButton(1, "Play/Pause Animation", &colorTheme.ButtonTheme) {
				svState.AnimationRunning = !svState.AnimationRunning
			}

			tomato.TextButton(2, "Render to .mp4", &colorTheme.ButtonTheme)
			tomato.TextButton(3, "Current Frame to .png", &colorTheme.ButtonTheme)
			if tomato.TextButton(4, "Load Configuration", &colorTheme.ButtonTheme) {
				svState.ConfigChooserOpened = !svState.ConfigChooserOpened
				if svState.ConfigChooserOpened {
					configFiles = ListAvailableConfigFiles(".")
					tomato.InvalidateElements() // Preamtively delete all buttons in this layout for later redraw
				}
			}

			if svState.ConfigChooserOpened {
				for i, configPath := range configFiles {
					if tomato.TextButton(5+i, configPath, &colorTheme.ButtonChooserTheme) {
						svState.ConfigChooserOpened = false
						simulationToggle <- false

						err, simulation = sim.MakeSimulationFromConfig(configPath)
						if err != nil {
							svState.TermMsg = fmt.Sprintf("%v", err)
						} else {
							svState.TermMsg = fmt.Sprintf("!loaded `%v` sucessfully!", configPath)
						}
						animator = sim.MakeAnimator(&simulation)
						svState.CurrentFrame = animator.Frames[0]
						svState.CursorPos = 0
						svState.AnimationRunning = false
						dataViewer.Mutex.Lock()
						dataViewer.Values = dataViewer.Values[:0]
						dataViewer.Mutex.Unlock()
					}
				}
			}
		}

		// Terminal
		Terminal(image.Rect(W-PANEL_W, RENDERER_H, W, H), colorTheme, svState.TermMsg)

		// Seeker
		where := image.Rect(0, RENDERER_H, RENDERER_W, RENDERER_H+SEEKER_H)
		cursorPosBefore := svState.CursorPos
		svState.CursorPos = Seeker(len(animator.Frames),
			svState.CursorPos,
			where,
			colorTheme,
			eventsThisTick)
		if cursorPosBefore != svState.CursorPos {
			svState.CurrentFrame = animator.Frames[svState.CursorPos]
		}

		// DataViewer
		dvWhere := image.Rect(0, RENDERER_H+SEEKER_H, RENDERER_W, RENDERER_H+SEEKER_H+BOT_PANEL_H)
		dataViewer.DrawDataViewer(svState.CursorPos, colorTheme, dvWhere)

		// Frame
		if svState.AnimationRunning {
			svState.CursorPos = (svState.CursorPos + 1) % len(animator.Frames)
			svState.CurrentFrame = animator.Frames[svState.CursorPos]
		}
		tomato.ToDraw(image.Rect(0, 0, RENDERER_W, RENDERER_H), svState.CurrentFrame)

		tomato.DrawUi()
		tomato.Win.SwapBuffers()
		tomato.Clear()

		// FPS target does actually not work correctly... the real fps are lower thean requeseted..
		dt := time.Since(previousTime)
		const desiredFrameTime time.Duration = time.Second / 120
		if dt < desiredFrameTime {
			time.Sleep(desiredFrameTime - dt)
		}

		// is measureing correct?
		fmt.Printf("FPS %.4v\n", 1.0/time.Since(previousTimeActually).Seconds())
		previousTimeActually = time.Now()
		previousTime = time.Now()
		tick++
	}
}

// Background Process for starting/stopping simulation
func Simulator(simToggle <-chan bool, simulation *sim.Simulation, animator *sim.Animator) {
	running := false
	for {
		select {
		case x := <-simToggle:
			// in case really wanna turn it off
			if !x {
				running = false
			} else {
				running = !running
			}
		default:
			if running {
				simulation.Step()
				energy := simulation.TotalEnergy()
				//energy := simulation.TotalDensity()
				animator.Frame()

				dataViewer.Mutex.Lock()
				dataViewer.Values = append(dataViewer.Values, energy)
				dataViewer.Mutex.Unlock()
			}
		}
		time.Sleep(time.Second / 960)
	}
}

// @Todo use an id per element and an active id to tracke which element is clicked last
// but for now just use a global variable
var seekerPressed bool

// try immediate ui style, still the event list should not be needed to detect drag and clicked
// return the new position of the cursor
func Seeker(frameCount, cursorPos int, where image.Rectangle, colorTheme *Theme, events []tomato.Ev) int {
	if cursorPos >= frameCount && frameCount != 0 {
		panic("cursor pos higher that frame count!")
	}

	if frameCount == 0 {
		frameCount = 1
	}

	rect := where.Sub(where.Min)
	drw := image.NewRGBA(rect)

	//
	// draw background
	//

	imgUni := image.NewUniform(colorTheme.SeekerBackground)
	draw.Draw(drw, rect, imgUni, image.ZP, draw.Src)

	//
	// draw frame hints
	//

	imgUni = image.NewUniform(colorTheme.SeekerFrame)
	frameRect := rect

	frameDx := float32(rect.Dx()) / float32(frameCount)

	if frameDx <= SEEKER_MIN_W {
		draw.Draw(drw, rect, imgUni, image.ZP, draw.Src)
		frameDx = float32(SEEKER_MIN_W)
	} else {
		for i := range frameCount {
			frameRect.Min.X = int(frameDx*float32(i)) + SEEKER_PAD/2
			frameRect.Max.X = int(frameDx*float32(i+1)) - SEEKER_PAD/2
			draw.Draw(drw, frameRect, imgUni, image.ZP, draw.Src)
		}
	}

	//
	// draw cursor
	//

	imgUni = image.NewUniform(colorTheme.SeekerCursor)

	cursorW := Min(int(frameDx), int(SEEKER_W))
	cursorOffset := image.Point{int(frameDx/2 - float32(cursorW)/2), 0}
	cursorRect := rect
	cursorRect.Min.X = int((float32(rect.Dx()) / float32(frameCount)) * float32(cursorPos))
	cursorRect.Max.X = cursorRect.Min.X + cursorW
	cursorRect = cursorRect.Add(cursorOffset)

	draw.Draw(drw, cursorRect.Intersect(rect), imgUni, image.ZP, draw.Src)
	tomato.ToDraw(where, drw)

	// check for drag, click and return the updated cursor pos
	for _, event := range events {
		switch event.Kind {
		case tomato.MouDown:
			if event.Point.In(where) && frameCount != 0 {
				cursorPos = int(float32(event.Point.X) / float32(RENDERER_W) * float32(frameCount))
				seekerPressed = true
			}
		// @Todo might be problematic beacuse the MouUp event might never arrive!
		case tomato.MouUp:
			seekerPressed = false
		case tomato.MouMove:
			if seekerPressed && frameCount != 0 {
				cursorPos = int(float32(event.Point.X) / float32(RENDERER_W) * float32(frameCount))

				if cursorPos >= frameCount {
					cursorPos = frameCount - 1
				}

				if cursorPos < 0 {
					cursorPos = 0
				}
			}
		}
	}

	return cursorPos
}

func (dv *DataViewInfo) DrawDataViewer(cursorPos int, colorTheme *Theme, where image.Rectangle) {
	rect := where.Sub(where.Min)
	drw := image.NewRGBA(rect)

	col := image.NewUniform(colorTheme.ProfilerForeground)
	curCol := image.NewUniform(colorTheme.ProfilerCursor)
	bgCol := image.NewUniform(colorTheme.ProfilerBackground)

	draw.Draw(drw, rect, bgCol, image.ZP, draw.Src)

	if len(dv.Values) == 0 {
		return
	}

	dv.Mutex.Lock()
	minV := slices.Max(dv.Values)
	maxV := slices.Min(dv.Values)
	dV := maxV - minV
	dx := float32(rect.Dx()) / float32(len(dv.Values)+1)

	// draw cursor
	// TODO: make cursor here and in seeker behave the same
	{
		rectC := rect
		rectC.Min.X = int(dx * float32(cursorPos))
		rectC.Max.X = int(dx*float32(cursorPos)) + Max(int(dx), SEEKER_MIN_W)
		draw.Draw(drw, rectC.Intersect(rect), curCol, image.ZP, draw.Src)
	}

	// draw energy scaled to height
	rectC := rect
	for i := range len(dv.Values) {
		h := int((dv.Values[i]-minV)/dV*float64(rect.Dy())) + rect.Min.Y
		rectC.Min.X = int(dx * float32(i+1))
		rectC.Max.X = int(dx * float32(i+2))
		rectC.Min.Y = h
		rectC.Max.Y = h + 2
		draw.Draw(drw, rectC, col, image.ZP, draw.Src)
	}
	dv.Mutex.Unlock()

	tomato.ToDraw(where, drw)
}

type DataViewInfo struct {
	Label  string
	Values []float64
	Mutex  sync.Mutex
}

func ListAvailableConfigFiles(configFolder string) []string {
	entries, err := os.ReadDir(configFolder)
	if err != nil {
		panic(err)
	}

	configFilePaths := make([]string, 0, 20)
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".sph-config") {
			configFilePaths = append(configFilePaths, e.Name())
		}
	}
	return configFilePaths
}

type Terminal_Info struct {
	Surface image.Image
	LastMsg string
}

var ti Terminal_Info

func Terminal(r image.Rectangle, colorTheme *Theme, message string) {

	if message != ti.LastMsg || ti.Surface == nil {
		redraw := func(msg string, sucess bool) draw.Image {
			rect := r.Sub(r.Min)
			drw := image.NewRGBA(rect)
			bg := image.NewUniform(colorTheme.TermBackground)
			draw.Draw(drw, rect, bg, image.ZP, draw.Src)

			textColor := colorTheme.ErrorRed
			if sucess {
				textColor = colorTheme.SuccessGreen
			}

			textImage := tomato.RenderTextMulti(msg, textColor, colorTheme.TermBackground, colorTheme.TermFont, rect.Dx())
			textRect := rect
			draw.Draw(drw, textRect, textImage, textImage.Bounds().Min, draw.Src)
			return drw
		}

		ti.LastMsg = message
		if rune(message[0]) == '!' {
			ti.Surface = redraw(message[1:], true)
		} else {
			ti.Surface = redraw(message, false)
		}
	}

	tomato.ToDraw(r, ti.Surface)
}

func RenderText(text string, textColor, btnColor color.RGBA, fontFace font.Face) draw.Image {

	drawer := &font.Drawer{
		Src:  &image.Uniform{textColor},
		Face: fontFace,
		Dot:  fixed.P(0, 0),
	}

	b26_6, _ := drawer.BoundString(text)
	bounds := image.Rect(
		b26_6.Min.X.Floor(),
		b26_6.Min.Y.Floor(),
		b26_6.Max.X.Ceil(),
		b26_6.Max.Y.Ceil(),
	)

	drawer.Dst = image.NewRGBA(bounds)
	btnUpUniform := image.NewUniform(btnColor)
	draw.Draw(drawer.Dst, bounds, btnUpUniform, image.ZP, draw.Src)
	drawer.DrawString(text)
	return drawer.Dst
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	run()
}
