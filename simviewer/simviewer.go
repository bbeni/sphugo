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

	ButtonText  color.RGBA
	ButtonUp    color.RGBA
	ButtonHover color.RGBA
	ButtonBlink color.RGBA

	OptionText  color.RGBA
	OptionUp    color.RGBA
	OptionHover color.RGBA

	TermBackground color.RGBA
	ErrorRed       color.RGBA
	SuccessGreen   color.RGBA

	SeekerBackground color.RGBA
	SeekerFrame      color.RGBA
	SeekerCursor     color.RGBA

	ProfilerForeground color.RGBA
	ProfilerBackground color.RGBA
	ProfilerCursor     color.RGBA

	FontFace     font.Face
	FontFaceTerm font.Face
}


// slowly migrating to this
type SimviewerState struct {
	CursorPos int
	AnimationRunning bool
	CurrentFrame image.Image
	TermMsg string
}

var svState SimviewerState
func run() {

	W := RENDERER_W + PANEL_W
	H := RENDERER_H + SEEKER_H + BOT_PANEL_H

	//var fontMu sync.Mutex
	var fontFace font.Face
	{
		font, err := truetype.Parse(gomono.TTF)
		if err != nil {
			panic(err)
		}

		fontFace = truetype.NewFace(font, &truetype.Options{
			Size: TEXT_SIZE,
		})
	}

	var fontFaceTerm font.Face
	{
		font, err := truetype.Parse(gomono.TTF)
		if err != nil {
			panic(err)
		}

		fontFaceTerm = truetype.NewFace(font, &truetype.Options{
			Size: TERM_TEXT_SIZE,
		})
	}

	colorTheme := &Theme{
		Background: color.RGBA{24, 24, 24, 255},

		ButtonText:  color.RGBA{255, 250, 240, 255}, // Floral White
		ButtonUp:    color.RGBA{36, 33, 36, 255},    // Raisin Black
		ButtonHover: color.RGBA{45, 45, 45, 255},
		ButtonBlink: color.RGBA{70, 70, 70, 255},

		TermBackground: color.RGBA{36, 23, 36, 255},
		ErrorRed:       color.RGBA{220, 60, 50, 255},
		SuccessGreen:   color.RGBA{60, 220, 23, 255},

		OptionText:  color.RGBA{36, 33, 36, 255},    // Raisin Black
		OptionUp:    color.RGBA{255, 250, 240, 255}, // Floral White
		OptionHover: color.RGBA{240, 220, 200, 255},

		SeekerFrame:      color.RGBA{140, 0, 0, 255},
		SeekerBackground: color.RGBA{120, 0, 30, 255},
		SeekerCursor:     color.RGBA{220, 20, 50, 255},

		ProfilerForeground: color.RGBA{10, 180, 20, 255},
		ProfilerBackground: color.RGBA{24, 24, 24, 255},
		ProfilerCursor:     color.RGBA{48, 48, 48, 255},

		FontFace:     fontFace,
		FontFaceTerm: fontFaceTerm,
	}

	if err := tomato.Create(W, H, "SFUGO - Simulation Renderer"); err != nil {
		panic(err)
	}

	/*w.Draw() <- func(drw draw.Image) image.Rectangle {
		r := image.Rect(0, 0, W, H)
		backgroundImg := image.NewUniform(colorTheme.Background)
		draw.Draw(drw, r, backgroundImg, image.ZP, draw.Src)
		return r
	}*/

	simulationToggle := make(chan bool) // if false is sent it turns it off
	energyProfiler := make(chan float64)


	//energyProfilerReset := make(chan bool)
	//drawOnce := make(chan bool)

	// Event Fowarding Channels
	// @Todo make this behind some abstraction maybe?
	//forwarding := make([]chan tomato.Ev, 7)
	//forwarding_boxes := make([]image.Rectangle, len(forwarding))
	//for i := range forwarding {
	//	forwarding[i] = make(chan tomato.Ev)
	//}
	

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
	//animator   := sim.MakeAnimatorGL(&simulation)

	go Simulator(simulationToggle, energyProfiler, &simulation, &animator)

	/*
	xa := W - PANEL_W
	xb := W

	go Button("Load Configuration", colorTheme,
		forwarding[4],
		image.Rect(xa, 4*(BTN_H+MARGIN_BOT), xb, 4*(BTN_H+MARGIN_BOT)+BTN_H),
		&fontMu, func() {
			go ConfigChoser(forwarding[5],
				"./",
				image.Rect(xa, 5*(BTN_H+MARGIN_BOT), xb, RENDERER_H),
				colorTheme,
				&fontMu,
				func(configPath string) {
					//simulationToggle <- false
					//animationToggle <- false
					err, simulation = sim.MakeSimulationFromConfig(configPath)
					if err != nil {
						msgStream <- fmt.Sprintf("%v", err)
					} else {
						msgStream <- fmt.Sprintf("!loaded `%v` sucessfully!", configPath)
					}

					//seekerChanged1 <- 0
					//animator   = sim.MakeAnimator(&simulation)
					//animator = sim.MakeAnimatorGL(&simulation)
					//env.GL() <- func() { animator.Init(W, H) }

					//drawOnce <- true
					//energyProfilerReset <- true
					//framesChanged <- 1
				})
		})*/

	tick := 0
	eventsThisTick := make([]tomato.Ev, 0)
	previousTime := time.Now()

	svState.CurrentFrame = animator.Frames[svState.CursorPos]
	dataViewer := MakeDv()

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
			case energyChanged := <-energyProfiler:
				dataViewer.AddValue(energyChanged)
			default:
				break events_loop
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

		dvWhere := image.Rect(0, RENDERER_H+SEEKER_H, RENDERER_W, RENDERER_H+SEEKER_H+BOT_PANEL_H)
		dataViewer.DrawDataViewer(svState.CursorPos, colorTheme, dvWhere)

		if svState.AnimationRunning {
			svState.CursorPos = (svState.CursorPos + 1) % len(animator.Frames)
			svState.CurrentFrame = animator.Frames[svState.CursorPos]
		}
		tomato.ToDraw(image.Rect(0, 0, RENDERER_W, RENDERER_H), svState.CurrentFrame)

		{
			tomato.Layout(0, tomato.Vertical, image.Rect(W - PANEL_W, 0, W, RENDERER_H))

			if tomato.TextButton(0, "Run/Stop Simulation") {
				// race occurs here, for now we just async call the toggle to not block here
				// it feels kinda slower now ... but not so sure
				go func() {simulationToggle <- true}()
				//simulationToggle <- true
			}

			if tomato.TextButton(1, "Play/Pause Animation") {
				svState.AnimationRunning = !svState.AnimationRunning
			}
			tomato.TextButton(2, "Render to .mp4")
			tomato.TextButton(3, "Current Frame to .png")
			tomato.TextButton(4, "Load Configuration")
		}

		tomato.DrawUi()
		tomato.Win.SwapBuffers()
		tomato.Clear()

		dt := time.Now().Sub(previousTime)
		previousTime = time.Now()
		_ = dt
		fmt.Printf("%.3v FPS\n", 1/dt.Seconds())
		time.Sleep(time.Second / 120)
		tick++
	}
}

// Background Process for starting/stopping simulation
func Simulator(
	simToggle <-chan bool, // input
	energyProfiler chan<- float64, // output
	simulation *sim.Simulation, animator *sim.Animator) {
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
				//animator.AddFrame()
				energyProfiler <- energy
			}
			time.Sleep(time.Second / 100)
		}
	}
}


//@Todo use an id per element and an active id to tracke which element is clicked last
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


func (dv *Dv) DrawDataViewer(cursorPos int, colorTheme *Theme, where image.Rectangle) {
	rect := where.Sub(where.Min)
	drw := image.NewRGBA(rect)

	col := image.NewUniform(colorTheme.ProfilerForeground)
	curCol := image.NewUniform(colorTheme.ProfilerCursor)
	bgCol := image.NewUniform(colorTheme.ProfilerBackground)

	draw.Draw(drw, rect, bgCol, image.ZP, draw.Src)

	frameCount := len(dv.Energies)

	if frameCount == 0 {
		return
	}

	minE := slices.Max(dv.Energies)
	maxE := slices.Min(dv.Energies)
	dE := maxE - minE
	dx := float32(rect.Dx()) / float32(frameCount+1)

	// draw cursor
	// TODO: make cursor here and in seeker behave the same
	{
		rectC := rect
		rectC.Min.X = int(dx*float32(cursorPos))
		rectC.Max.X = int(dx*float32(cursorPos)) + Max(int(dx), SEEKER_MIN_W)
		draw.Draw(drw, rectC.Intersect(rect), curCol, image.ZP, draw.Src)
	}

	// draw energy scaled to height
	rectC := rect
	for i := range frameCount {
		h := int((dv.Energies[i]-minE)/dE*float64(rect.Dy())) + rect.Min.Y
		rectC.Min.X = int(dx * float32(i+1))
		rectC.Max.X = int(dx * float32(i+2))
		rectC.Min.Y = h
		rectC.Max.Y = h + 2
		draw.Draw(drw, rectC, col, image.ZP, draw.Src)
	}

	tomato.ToDraw(where, drw)
}

func (dv *Dv) Reset() {
	dv.Energies = dv.Energies[:0]
}

func (dv *Dv) AddValue(v float64) {
	dv.Energies = append(dv.Energies, v)
}

func MakeDv() Dv {
	return Dv{
		Label: "Energy",
	}
}

// @Todo rename
type Dv struct {
	Label string
	Energies []float64
}



func ConfigChoser(events chan tomato.Ev, configFolder string, r image.Rectangle, colorTheme *Theme, mu *sync.Mutex, callback func(string)) {

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

	// make textures for options
	textImagesUp    := make([]image.Image, 0, len(configFilePaths))
	textImagesHover := make([]image.Image, 0, len(configFilePaths))
	for _, path := range configFilePaths {
		text := path
		var textImageUp    image.Image
		var textImageHover image.Image
		{
			mu.Lock()
			textImageUp    = RenderText(text, colorTheme.OptionText, colorTheme.OptionUp, colorTheme.FontFace)
			textImageHover = RenderText(text, colorTheme.OptionText, colorTheme.OptionHover, colorTheme.FontFace)
			mu.Unlock()
		}
		textImagesUp    = append(textImagesUp, textImageUp)
		textImagesHover = append(textImagesHover, textImageHover)
	}


	drawOption := func(index int) (draw.Image, draw.Image, image.Rectangle) {

		rect := r
		rect.Min.Y = r.Min.Y + index * OPTION_H
		rect.Max.Y = r.Min.Y + (index + 1)  * OPTION_H - OPTION_PAD

		textBounds := textImagesUp[index].Bounds()
		var buttonBgHover  = image.NewUniform(colorTheme.OptionHover)
		var buttonBgUp  = image.NewUniform(colorTheme.OptionUp)

		drwUp    := image.NewRGBA(r)
		drwHover := image.NewRGBA(r)
		
		draw.Draw(drwUp,    rect, buttonBgUp, image.ZP, draw.Src)
		draw.Draw(drwHover, rect, buttonBgHover, image.ZP, draw.Src)

		textRect := rect
		textRect.Min.Y += textRect.Dy()/2 - textBounds.Dy()/2
		textRect.Min.X += textRect.Dx()/2 - textBounds.Dx()/2

		draw.Draw(drwUp, textRect, textImagesUp[index], textBounds.Min, draw.Src)
		
		draw.Draw(drwHover, textRect, textImagesHover[index], textBounds.Min, draw.Src)
		
		return drwUp, drwHover, rect
	}

	var optionsHover []draw.Image
	var optionsUp    []draw.Image
	var optionsRects []image.Rectangle

	for i := range configFilePaths {
		u, h, r := drawOption(i) 
		optionsHover = append(optionsHover, h)
		optionsUp    = append(optionsUp, u)
		optionsRects = append(optionsRects, r)
		tomato.ToDraw(r, u)
	}

	over := -1

	exit:
	for event := range events {
		switch event.Kind {
		case tomato.MouDown:
			i := (event.Point.Y - r.Min.Y) / OPTION_H
			if event.Point.In(r) && i >= 0 && i < len(configFilePaths) {
				callback(configFilePaths[i])
				break exit
			}
		case tomato.MouMove:
			if event.Point.In(r) {
				supposedIndex := (event.Point.Y - r.Min.Y) / OPTION_H
				if supposedIndex != over && supposedIndex >= 0 && supposedIndex < len(configFilePaths){
					tomato.ToDraw(optionsRects[supposedIndex], optionsHover[supposedIndex])
					if over >= 0 {
						tomato.ToDraw(optionsRects[over], optionsUp[over])
					}
					over = supposedIndex
				}
			}
		}
	}

	fmt.Println(len(optionsHover))

	tomato.ToDraw(r, image.NewUniform(colorTheme.Background))
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

			textImage := RenderTextMulti(msg, textColor, colorTheme.TermBackground, colorTheme.FontFaceTerm, rect.Dx())
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

func RenderTextMulti(text string, textColor, bgColor color.RGBA, fontFace font.Face, maxWidth int) draw.Image {

	drawer := &font.Drawer{
		Src:  &image.Uniform{textColor},
		Face: fontFace,
		Dot:  fixed.P(0, 0),
	}

	lines := make([]string, 0)

	j := 0
	i := 0
	for i = 0; i < len(text)-1; i++ {
		b26_6, _ := drawer.BoundString(text[j : i+1])
		if b26_6.Max.X.Ceil()-b26_6.Min.X.Floor() > maxWidth {
			lines = append(lines, text[j:i])
			j = i
		}
	}

	if i != j {
		lines = append(lines, text[j:i])
	}

	maxW := 0
	lineH := 0
	for _, line := range lines {
		b26_6, _ := drawer.BoundString(line)
		bounds := image.Rect(
			b26_6.Min.X.Floor(),
			b26_6.Min.Y.Floor(),
			b26_6.Max.X.Ceil(),
			b26_6.Max.Y.Ceil(),
		)
		if bounds.Dx() > maxW {
			maxW = bounds.Dx()
		}
		if bounds.Dy() > lineH {
			lineH = bounds.Dy()
		}
	}

	bounds := image.Rect(0, 0, maxW, (len(lines)+1)*lineH)
	result := image.NewRGBA(bounds)
	bgUniform := image.NewUniform(bgColor)
	draw.Draw(result, bounds, bgUniform, image.ZP, draw.Src)

	for i, line := range lines {
		tImage := RenderText(line, textColor, bgColor, fontFace)
		draw.Draw(result, bounds, tImage, bounds.Min.Sub(image.Pt(0, lineH*(i+1))), draw.Src)
	}
	return result
}

func Button(text string, colorTheme *Theme,
	events chan tomato.Ev,
	r image.Rectangle, mu *sync.Mutex, clicked func()) {

	var textImageUp image.Image
	var textImageHover image.Image
	{
		mu.Lock()
		textImageUp = RenderText(text, colorTheme.ButtonText, colorTheme.ButtonUp, colorTheme.FontFace)
		textImageHover = RenderText(text, colorTheme.ButtonText, colorTheme.ButtonHover, colorTheme.FontFace)
		mu.Unlock()
	}

	redraw := func(visible bool, hover bool) draw.Image {
		img := image.NewRGBA(r)
		var textImage image.Image
		var buttonBg image.Image
		if hover {
			buttonBg = image.NewUniform(colorTheme.ButtonHover)
			textImage = textImageHover
		} else {
			buttonBg = image.NewUniform(colorTheme.ButtonUp)
			textImage = textImageUp
		}
		if visible {
			draw.Draw(img, r, buttonBg, image.ZP, draw.Src)
			textRect := r
			textRect.Min.Y += textRect.Dy()/2 - textImage.Bounds().Dy()/2
			textRect.Min.X += textRect.Dx()/2 - textImage.Bounds().Dx()/2

			draw.Draw(img, textRect, textImage, textImage.Bounds().Min, draw.Src)
		} else {
			draw.Draw(img, r, image.NewUniform(colorTheme.ButtonBlink), image.ZP, draw.Src)
		}
		return img
	}

	normalImg := redraw(true, false)
	hoveredImg := redraw(true, true)
	blinkImg := redraw(false, false)

	tomato.ToDraw(r, normalImg)

	for ev := range events {
		if ev.Kind == tomato.MouMove {
			if ev.Point.In(r) {
				tomato.ToDraw(r, hoveredImg)
			} else {
				tomato.ToDraw(r, normalImg)
			}
		}
		if ev.Kind == tomato.MouDown {
			if ev.Point.In(r) {
				clicked()
				for i := 0; i < 3; i++ {
					tomato.ToDraw(r, blinkImg)
					time.Sleep(time.Second / 10)
					tomato.ToDraw(r, hoveredImg)
					time.Sleep(time.Second / 10)
				}
			}
		}
	}
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
