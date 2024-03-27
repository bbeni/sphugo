package main

import (
	"slices"
	"image"
	"image/draw"
	"image/color"
	"time"
	"sync"
	"fmt"
	"log"

	"github.com/faiface/gui"
	"github.com/faiface/gui/win"
	"github.com/faiface/mainthread"

	"github.com/golang/freetype/truetype"

	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"

	"github.com/bbeni/sphugo/sim"
)

// TODO: remove later
var _ = fmt.Println

const (
	PANEL_W    = 330
	RENDERER_W = 1280
	RENDERER_H = 720

	TEXT_SIZE  = 24
	BTN_H 	   = 70
	MARGIN_BOT = 4

	SEEKER_H   = 24
	SEEKER_W   = 36
	SEEKER_MIN_W = 8
	SEEKER_FRAME_MARGIN = 4

	BOT_PANEL_H = 220
)

type Theme struct {
	Background	   color.RGBA

	ButtonText	   color.RGBA
	ButtonUp	   color.RGBA
	ButtonHover	   color.RGBA
	ButtonBlink    color.RGBA

	SeekerBackground color.RGBA
	SeekerFrame		 color.RGBA
	SeekerCursor	 color.RGBA

	ProfilerForeground color.RGBA
	ProfilerBackground color.RGBA
	ProfilerCursor     color.RGBA

	FontFace	   font.Face
}

func run() {

	W 			:= RENDERER_W + PANEL_W
	H 	  		:= RENDERER_H + SEEKER_H + BOT_PANEL_H


	var fontMu sync.Mutex

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

	colorTheme := &Theme{
		Background:  	 	color.RGBA{24,   24,  24, 255},

		ButtonText: 	 	color.RGBA{255, 250, 240, 255}, // Floral White
		ButtonUp:   	 	color.RGBA{36,   33,  36, 255}, // Raisin Black
		ButtonHover:	 	color.RGBA{45,   45,  45, 255},
		ButtonBlink:	 	color.RGBA{70,   70,  70, 255},

		SeekerFrame:	 	color.RGBA{140,   0,   0, 255},
		SeekerBackground:	color.RGBA{120,   0,  30, 255},
		SeekerCursor:	 	color.RGBA{220,  20,  50, 255},

		ProfilerForeground: color.RGBA{10,  180,  20, 255},
		ProfilerBackground: color.RGBA{24,   24,  24, 255},
		ProfilerCursor:     color.RGBA{48,   48,  48, 255},

		FontFace:		  fontFace,
	}

	w, err := win.New(win.Title("SFUGO - Simulation Renderer"), win.Size(W, H))
	if err != nil {
		panic(err)
	}

	w.Draw() <- func(drw draw.Image) image.Rectangle {
		r := image.Rect(0, 0, W, H)
		backgroundImg := image.NewUniform(colorTheme.Background)
		draw.Draw(drw, r, backgroundImg, image.ZP, draw.Src)
		return r
	}


	// create example config file if not existent
	exampleConfigFilePath := "./example.sph-config"
	sim.GenerateDefaultConfigFile(exampleConfigFilePath)

	simulation := sim.MakeSimulationFromConfig(exampleConfigFilePath)
	animator   := sim.MakeAnimator(&simulation)

	// TODO: cleanup this mess, too many channels and/or missleading names!
	simulationToggle := make(chan bool)
	animationToggle  := make(chan bool)
	framesChanged    := make(chan int)
	cursorChanged    := make(chan int)
	seekerChanged1   := make(chan int)
	seekerChanged2   := make(chan int)
	energyProfiler   := make(chan float64)

	mux, env := gui.NewMux(w)

	{
		xa := W - PANEL_W
		xb := W
		go Button(mux.MakeEnv(), "Run/Stop Simulation", colorTheme,
			image.Rect(xa, 0, xb, BTN_H),
			&fontMu, func() {
				simulationToggle <- true
		})
		go Button(mux.MakeEnv(), "Play/Pause Animation", colorTheme,
			image.Rect(xa, 1*(BTN_H+MARGIN_BOT), xb, 1*(BTN_H+MARGIN_BOT)+BTN_H),
			&fontMu, func() {
				animationToggle <- true
		})
		go Button(mux.MakeEnv(), "Load Configuration", colorTheme,
			image.Rect(xa, 2*(BTN_H+MARGIN_BOT), xb, 2*(BTN_H+MARGIN_BOT)+BTN_H),
			&fontMu, func() {

		})
		go Button(mux.MakeEnv(), "Render to .mp4", colorTheme,
			image.Rect(xa, 3*(BTN_H+MARGIN_BOT), xb, 3*(BTN_H+MARGIN_BOT)+BTN_H),
			&fontMu, func() {

		})
		go Button(mux.MakeEnv(), "Current Frame to .png", colorTheme,
			image.Rect(xa, 4*(BTN_H+MARGIN_BOT), xb, 4*(BTN_H+MARGIN_BOT)+BTN_H),
			&fontMu, func() {
				i := animator.ActiveFrame
				file_path := fmt.Sprintf("frame%v.png", i)
				animator.FrameToPNG(file_path)
				log.Printf("Created PNG %s.", file_path)
		})
	}

	{
		go Simulator(mux.MakeEnv(),
			simulationToggle,
			framesChanged, energyProfiler,
			&simulation, &animator)

		go Renderer(mux.MakeEnv(),
			animationToggle, seekerChanged1,
			cursorChanged, seekerChanged2,
			image.Rect(0, 0, RENDERER_W, RENDERER_H), &animator)

		go Seeker(mux.MakeEnv(),
			framesChanged, cursorChanged,
			seekerChanged1, seekerChanged2,
			image.Rect(0, RENDERER_H, RENDERER_W, RENDERER_H+SEEKER_H), colorTheme)

		go DataViewer(mux.MakeEnv(),
			framesChanged, seekerChanged2, energyProfiler,
			image.Rect(0, RENDERER_H + SEEKER_H, RENDERER_W, RENDERER_H+SEEKER_H+BOT_PANEL_H), colorTheme, &simulation)

	}


	//simulationToggle <- true
	//animationToggle <- true


	// we use the master env now, w is used by the mux
	for event := range env.Events() {
		switch ev := event.(type) {
		case win.WiClose:
			close(env.Draw())
		case win.KbDown:
			switch ev.Key {
			case win.KeyEscape:
				close(env.Draw())
			case win.KeySpace:
				animationToggle <- true
			}
		case win.KbType:
			//fmt.Println(ev.String())
		}
	}
}


// Background Process for starting/stopping simulation
func Simulator(env gui.Env,
	simToggle <-chan bool, 										// input
	framesChanged chan<- int, energyProfiler chan<- float64,    // output
	simulation *sim.Simulation, animator *sim.Animator) {

	running := false
	for {
		select {
		case <-simToggle:
			running = !running
		default:
			if running {
				simulation.Step()
				energy := simulation.TotalEnergy()
				//energy := simulation.TotalDensity()

				animator.Frame()
				framesChanged  <- len(animator.Frames)
				energyProfiler <- energy
			}
		}
	}
}

func Seeker(env gui.Env,
	framesCh <-chan int, cursorCh <-chan int,   // input
	seekerChanged1, seekerChanged2 chan<- int,  // output
	r image.Rectangle, colorTheme *Theme) {

	frameCount := 0
	cursorPos  := 0

	// draws the seeker at a given position
	drawSeeker := func(frameCount, cursorPos int) func(draw.Image) image.Rectangle {

		if cursorPos >= frameCount && frameCount != 0 {
			panic("cursor pos higher that frame count!")
		}

		if frameCount == 0 {
			frameCount = 1
		}

		return func(drw draw.Image) image.Rectangle {

			//
			// draw background
			//

			imgUni := image.NewUniform(colorTheme.SeekerBackground)
			draw.Draw(drw, r, imgUni, image.ZP, draw.Src)

			//
			// draw frame hints
			//

			imgUni     = image.NewUniform(colorTheme.SeekerFrame)
			frameRect := r

			frameDx := float32(r.Dx()) / float32(frameCount)

			if frameDx <= SEEKER_MIN_W {
				draw.Draw(drw, r, imgUni, image.ZP, draw.Src)
				frameDx = float32(SEEKER_MIN_W)
			} else {
				for i := range frameCount {
					frameRect.Min.X = int(frameDx * float32(i))   + SEEKER_FRAME_MARGIN / 2
					frameRect.Max.X = int(frameDx * float32(i+1)) - SEEKER_FRAME_MARGIN / 2
					draw.Draw(drw, frameRect, imgUni, image.ZP, draw.Src)
				}
			}

			//
			// draw cursor
			//

			imgUni = image.NewUniform(colorTheme.SeekerCursor)

			cursorW := Min(int(frameDx), int(SEEKER_W))
			cursorOffset := image.Point{int(frameDx/2 - float32(cursorW)/2), 0}
			cursorRect := r
			cursorRect.Min.X = int((float32(r.Dx()) / float32(frameCount)) * float32(cursorPos) )
			cursorRect.Max.X = cursorRect.Min.X + cursorW
			cursorRect = cursorRect.Add(cursorOffset)

			draw.Draw(drw, cursorRect.Intersect(r), imgUni, image.ZP, draw.Src)
			return r
		}
	}

	needRedraw := true
	pressed := false
	for {
		select {
		case newCount := <-framesCh:
			frameCount = newCount
			needRedraw = true
		case newPos := <-cursorCh:
			cursorPos = newPos
			needRedraw = true
		case event := <-env.Events():
			switch event := event.(type) {
				case win.MoDown:
					if event.Point.In(r) && frameCount != 0 {
						cursorPos = int(float32(event.Point.X) / float32(RENDERER_W) * float32(frameCount))
						seekerChanged1 <- cursorPos
						seekerChanged2 <- cursorPos
						pressed = true
						needRedraw = true
					}
				case win.MoUp:
					pressed = false
				case win.MoMove:
					if pressed && frameCount != 0{
						oldCursorPos := cursorPos
						cursorPos = int(float32(event.Point.X) / float32(RENDERER_W) * float32(frameCount))
						if cursorPos >= frameCount {
							cursorPos = frameCount - 1
						}
						if cursorPos < 0 {
							cursorPos = 0
						}

						if oldCursorPos != cursorPos {
							needRedraw = true
							seekerChanged1 <- cursorPos
							seekerChanged2 <- cursorPos
						}
					}
			}
		}

		if needRedraw {
			nerfedCursorPos := cursorPos
			if cursorPos >= frameCount && frameCount != 0 {
				nerfedCursorPos = frameCount - 1
			}
			env.Draw() <- drawSeeker(frameCount, nerfedCursorPos)
			needRedraw = false
			time.Sleep(time.Second / 300)
		}
	}

}


func Renderer(env gui.Env,
	aniToggle <-chan bool, seekerChanged1 <-chan int, // input
	cursorCh chan<- int, seekerChanged2	chan<- int,	 // output
	r image.Rectangle, animator *sim.Animator) {

	drawFrame := func(i int) func(draw.Image) image.Rectangle {

		if i == 0 && len(animator.Frames) == 0 {
			return func(drw draw.Image) image.Rectangle {
				img := image.NewUniform(color.RGBA{0,0,0,255})
				draw.Draw(drw, r, img, image.ZP, draw.Src)
				return r
			}
		}

		animator.ActiveFrame = i

		return func(drw draw.Image) image.Rectangle {
			draw.Draw(drw, r, animator.Frames[i], image.ZP, draw.Src)
			return r
		}

	}

	needsRedraw := false
	running := false
	step := 0

	env.Draw() <- drawFrame(step)


	// TODO: decouple animation and event loop
	// maybe not here the lag happens! probably in other part of code
	for {
		select{
		case <- aniToggle:
			running = !running
		case frameNumber := <-seekerChanged1:
			step = frameNumber

			out:
			for _ = range 10 {
				time.Sleep(time.Second / 300)
				select {
				case fn := <-seekerChanged1:
					step = fn
				default:
					break out
				}
			}

			needsRedraw = true
		default:

			if animator != nil && !running && needsRedraw {
				env.Draw() <- drawFrame(step)
				needsRedraw = false
			}

			if running && animator != nil {
				if step >= len(animator.Frames) {
					step = 0
				}
				cursorCh       <- step
				seekerChanged2 <- step
				env.Draw() <- drawFrame(step)
				step += 1

				time.Sleep(time.Second / 60)
			}
		}
	}
}

func DataViewer(env gui.Env,
	framesCh <-chan int, seekerChanged <-chan int, energyProfiler <-chan float64, // input
	r image.Rectangle, colorTheme *Theme, simulation *sim.Simulation) {

	redraw := func(energies []float64, cursorPos int) func(drw draw.Image) image.Rectangle {
		return func(drw draw.Image) image.Rectangle {

			col    := image.NewUniform(colorTheme.ProfilerForeground)
			curCol := image.NewUniform(colorTheme.ProfilerCursor)
			bgCol  := image.NewUniform(colorTheme.ProfilerBackground)

			draw.Draw(drw, r, bgCol, image.ZP, draw.Src)

			frameCount := len(energies)

			if frameCount == 0 {
				return r
			}

			minE := slices.Max(energies)
			maxE := slices.Min(energies)
			dE   := maxE - minE
			dx   := float32(r.Dx()) / float32(frameCount)

			// draw cursor
			// TODO: make cursor here and in seeker behave the same
			{
				rect := r
				rect.Min.X = int(dx * float32(cursorPos))
				rect.Max.X = int(dx * float32(cursorPos)) + Max(int(dx), SEEKER_MIN_W)
				draw.Draw(drw, rect.Intersect(r), curCol, image.ZP, draw.Src)
			}

			// draw energy scaled to height
			rect := r
			for i := range frameCount {
				h := int((energies[i] - minE) / dE * float64(r.Dy())) + r.Min.Y
				rect.Min.X = int(dx * float32(i))
				rect.Max.X = int(dx * float32(i+1))
				rect.Min.Y = h
				rect.Max.Y = h+2
				draw.Draw(drw, rect, col, image.ZP, draw.Src)
			}

			return r
		}
	}

	cursorPos := 0
	energies  := make([]float64, 0)
	env.Draw() <- redraw(energies, cursorPos)

	for {
		select {
		case cursorPos = <-seekerChanged:
			env.Draw() <- redraw(energies, cursorPos)
		case energy := <-energyProfiler:
			energies = append(energies, energy)
			env.Draw() <- redraw(energies, cursorPos)
		}
	}

}

func RenderText(text string, textColor, btnColor color.RGBA, fontFace font.Face) (draw.Image) {

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


func Button(env gui.Env, text string, colorTheme *Theme,
 	r image.Rectangle, mu *sync.Mutex, clicked func()) {

	var textImageUp    image.Image
	var textImageHover image.Image
	{
		mu.Lock()
		textImageUp    = RenderText(text, colorTheme.ButtonText, colorTheme.ButtonUp, colorTheme.FontFace)
		textImageHover = RenderText(text, colorTheme.ButtonText, colorTheme.ButtonHover, colorTheme.FontFace)
		mu.Unlock()
	}

	redraw := func(visible bool, hover bool) func(draw.Image) image.Rectangle {
		return func(drw draw.Image) image.Rectangle {

			var textImage image.Image
			var buttonBg  image.Image
			if hover {
				buttonBg  = image.NewUniform(colorTheme.ButtonHover)
				textImage = textImageHover
			} else {
				buttonBg  = image.NewUniform(colorTheme.ButtonUp)
				textImage = textImageUp
			}

			if visible {
				draw.Draw(drw, r, buttonBg, image.ZP, draw.Src)
				textRect := r
				textRect.Min.Y += textRect.Dy()/2 - textImage.Bounds().Dy()/2
				textRect.Min.X += textRect.Dx()/2 - textImage.Bounds().Dx()/2

				draw.Draw(drw, textRect, textImage, textImage.Bounds().Min, draw.Src)
			} else {
				draw.Draw(drw, r, image.NewUniform(colorTheme.ButtonBlink), image.ZP, draw.Src)
			}
			return r
		}
	}

	env.Draw() <- redraw(true, false)

	for event := range env.Events() {
		switch event := event.(type) {
		case win.MoDown:
			if event.Point.In(r) {
				clicked()
				for i := 0; i < 3; i++ {
					env.Draw() <- redraw(false, false)
					time.Sleep(time.Second / 10)
					env.Draw() <- redraw(true, false)
					time.Sleep(time.Second / 10)
				}
			}
		case win.MoMove:
			if event.Point.In(r) {
				env.Draw() <- redraw(true, true)
			} else {
				env.Draw() <- redraw(true, false)
			}
		}

	}

	close(env.Draw())
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
	mainthread.Run(run)
}