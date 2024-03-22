package main

import (
	"image"
	"image/draw"
	"image/color"
	"time"
	"sync"
	"fmt"

	"github.com/faiface/gui"
	"github.com/faiface/gui/win"
	"github.com/faiface/mainthread"

	"github.com/golang/freetype/truetype"

	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"

	"github.com/bbeni/sphugo/tg"
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

	SEEKER_H   = 30
	SEEKER_W   = 10
	SEEKER_FRAME_MARGIN = 2

	BOT_PANEL_H = 200
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

	FontFace	   font.Face
}

func run() {

	W 			:= RENDERER_W + PANEL_W
	H 	  		:= RENDERER_H + SEEKER_H


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
		Background:  	  color.RGBA{18,   18,  18, 255},

		ButtonText: 	  color.RGBA{255, 250, 240, 255}, // Floral White
		ButtonUp:   	  color.RGBA{36,   33,  36, 255}, // Raisin Black
		ButtonHover:	  color.RGBA{45,   45,  45, 255},
		ButtonBlink:	  color.RGBA{70,   70,  70, 255},

		SeekerBackground: color.RGBA{140,   0,   0, 255},
		SeekerFrame:	  color.RGBA{120,   0,  30, 255},
		SeekerCursor:	  color.RGBA{220,   20,  50, 255},

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

	simulationToggle := make(chan bool)
	animationToggle  := make(chan bool)
	framesChanged    := make(chan int)
	cursorChanged    := make(chan int)

	mux, env := gui.NewMux(w)

	{
		xa := W - PANEL_W
		xb := W
		go Button(mux.MakeEnv(), "Run/Stop Simulation", colorTheme, image.Rect(xa, 0, xb, BTN_H), &fontMu, func() {
			simulationToggle <- true
		})
		go Button(mux.MakeEnv(), "Play/Pause Animation", colorTheme, image.Rect(xa, 1*(BTN_H+MARGIN_BOT), xb, 1*(BTN_H+MARGIN_BOT)+BTN_H), &fontMu, func() {
			animationToggle <- true
		})
		go Button(mux.MakeEnv(), "Load SPHUGO File", colorTheme, image.Rect(xa, 2*(BTN_H+MARGIN_BOT), xb, 2*(BTN_H+MARGIN_BOT)+BTN_H), &fontMu, func() {

		})
		go Button(mux.MakeEnv(), "Render mp4", colorTheme, image.Rect(xa, 3*(BTN_H+MARGIN_BOT), xb, 3*(BTN_H+MARGIN_BOT)+BTN_H), &fontMu, func() {

		})

	}


	simulation := tg.MakeSimulation()

	{
		go Simulator(mux.MakeEnv(), simulationToggle, framesChanged, &simulation)

		go Renderer(mux.MakeEnv(), animationToggle, framesChanged, cursorChanged,
			image.Rect(0, 0, RENDERER_W, RENDERER_H), &simulation)

		go Seeker(mux.MakeEnv(), framesChanged, cursorChanged,
			image.Rect(0, RENDERER_H, RENDERER_W, RENDERER_H+SEEKER_H), colorTheme)

		//go DataViewer(mux.MakeEnv(), framesChanged, cursorChanged, r image.Rectangle, simulation *tg.Simulation)

	}

	// we use the master env now, w is used by the mux
	for event := range env.Events() {
		switch event.(type) {
		case win.WiClose:
			close(env.Draw())
		case win.KbDown:
			close(env.Draw())
		}
	}
}

func Simulator(env gui.Env, simToggle <-chan bool, framesChanged chan<- int, simulation *tg.Simulation) {

	simulation.Init()

	running := false
	step := 0

	for {
		select {
		case _ = <-simToggle:
			running = !running
		default:
			if running {
				simulation.Step(step)
				step += 1
				framesChanged <- len(simulation.Frames)
			}
		}
	}
}


func Seeker(env gui.Env, framesCh chan int, cursorCh chan int,
		r image.Rectangle, colorTheme *Theme) {

	frameCount := 0
	cursorPos  := 0

	// draws the seeker at a given position
	drawSeeker := func(frameCount, cursorPos int) func(draw.Image) image.Rectangle {

		if cursorPos > frameCount{
			panic("cursor pos higher that frame count!")
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
			// width W, framewidth f, and padding 2, number of frames N
			// pixel and margin:     f 2 f 2 f
			//                   i = 0   1   2
			//
			// x values left: i*W/N
			// x values rigt: (i+1)*(w/N) - 2
			imgUni = image.NewUniform(colorTheme.SeekerFrame)
			frameRect := r


			// upper bound for number of frame displayed
			cap := r.Dx()/SEEKER_FRAME_MARGIN/2
			cappedFrameCount := frameCount
			if cappedFrameCount > cap {
				cappedFrameCount = cap
			}

			for i := range cappedFrameCount {
				frameRect.Min.X = int((float32(r.Dx()) / float32(cappedFrameCount)) * float32(i) )
				frameRect.Max.X = int((float32(r.Dx()) / float32(cappedFrameCount)) * float32(i+1) - SEEKER_FRAME_MARGIN)
				draw.Draw(drw, frameRect, imgUni, image.ZP, draw.Src)
			}

			//
			// draw cursor
			//

			imgUni = image.NewUniform(colorTheme.SeekerCursor)
			cursorRect := r
			if frameCount == 0 {
				frameCount = 1
			}
			cursorRect.Min.X = int((float32(r.Dx()) / float32(frameCount)) * float32(cursorPos) )
			cursorRect.Max.X = cursorRect.Min.X + SEEKER_W
			draw.Draw(drw, cursorRect, imgUni, image.ZP, draw.Src)
			return r
		}
	}

	env.Draw() <- drawSeeker(0, 0)

	for {
		select {
		case newCount := <-framesCh:
			frameCount = newCount
			env.Draw() <- drawSeeker(frameCount, cursorPos)
		case newPos := <-cursorCh:
			cursorPos = newPos
			env.Draw() <- drawSeeker(frameCount, cursorPos)
		}
	}

}


func Renderer(env gui.Env, aniToggle <-chan bool, framesCh chan<- int, cursorCh chan int,
		r image.Rectangle, simulation *tg.Simulation) {

	env.Draw() <- func(drw draw.Image) image.Rectangle {
		img := image.NewUniform(color.RGBA{0,0,0,255})
		draw.Draw(drw, r, img, image.ZP, draw.Src)
		return r
	}

	drawFrame := func(i int) func(draw.Image) image.Rectangle {
		return func(drw draw.Image) image.Rectangle {
			draw.Draw(drw, r, simulation.Frames[i], image.ZP, draw.Src)
			return r
		}
	}

	running := false
	step := 0

	for {
		select{
		case <- aniToggle:
			running = !running
		default:
			if running && simulation != nil {
				if step >= len(simulation.Frames) {
					step = 0
				}
				env.Draw() <- drawFrame(step)
				step += 1
				cursorCh <- step
				time.Sleep(time.Second / 60)
			}
		}
	}
}

func DataViewer(env gui.Env, framesCh chan<- int, cursorCh chan int,
					r image.Rectangle, simulation *tg.Simulation) {


	env.Draw() <- func(drw draw.Image) image.Rectangle {
		img := image.NewUniform(color.RGBA{0,100,0,255})
		draw.Draw(drw, r, img, image.ZP, draw.Src)
		return r
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


func main() {
	mainthread.Run(run)
}