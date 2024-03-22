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

	SEEKER_H   = 24
	SEEKER_W   = 16
	SEEKER_MIN_W = 8
	SEEKER_FRAME_MARGIN = 4

	BOT_PANEL_H = 240
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
		Background:  	  color.RGBA{18,   18,  18, 255},

		ButtonText: 	  color.RGBA{255, 250, 240, 255}, // Floral White
		ButtonUp:   	  color.RGBA{36,   33,  36, 255}, // Raisin Black
		ButtonHover:	  color.RGBA{45,   45,  45, 255},
		ButtonBlink:	  color.RGBA{70,   70,  70, 255},

		SeekerFrame:	  color.RGBA{140,   0,   0, 255},
		SeekerBackground: color.RGBA{120,   0,  30, 255},
		SeekerCursor:	  color.RGBA{220,  20,  50, 255},

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
	seekerChanged    := make(chan int)


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
		go Button(mux.MakeEnv(), "Load Configuration", colorTheme, image.Rect(xa, 2*(BTN_H+MARGIN_BOT), xb, 2*(BTN_H+MARGIN_BOT)+BTN_H), &fontMu, func() {

		})
		go Button(mux.MakeEnv(), "Render to mp4", colorTheme, image.Rect(xa, 3*(BTN_H+MARGIN_BOT), xb, 3*(BTN_H+MARGIN_BOT)+BTN_H), &fontMu, func() {

		})

	}


	simulation := tg.MakeSimulation()

	{
		go Simulator(mux.MakeEnv(), simulationToggle, framesChanged, &simulation)

		go Renderer(mux.MakeEnv(), animationToggle, framesChanged, cursorChanged, seekerChanged,
			image.Rect(0, 0, RENDERER_W, RENDERER_H), &simulation)

		go Seeker(mux.MakeEnv(), framesChanged, cursorChanged, seekerChanged,
			image.Rect(0, RENDERER_H, RENDERER_W, RENDERER_H+SEEKER_H), colorTheme)

		go DataViewer(mux.MakeEnv(), framesChanged, cursorChanged,
			image.Rect(0, RENDERER_H + SEEKER_H, RENDERER_W, RENDERER_H+SEEKER_H+BOT_PANEL_H), colorTheme, &simulation)

	}


	simulationToggle <- true
	animationToggle <- true


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
				animationToggle<- true
			}
		case win.KbType:
			//fmt.Println(ev.String())
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


func Seeker(env gui.Env, framesCh chan int, cursorCh chan int, seekerChanged chan<- int,
		r image.Rectangle, colorTheme *Theme) {

	frameCount := 0
	cursorPos  := 0

	// draws the seeker at a given position
	drawSeeker := func(frameCount, cursorPos int) func(draw.Image) image.Rectangle {

		if cursorPos >= frameCount && frameCount != 0{
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

			frameSx := float32(r.Dx()) / float32(frameCount)
			frameDx := float32(r.Dx()) / float32(frameCount) - float32(SEEKER_FRAME_MARGIN)

			if frameDx <= SEEKER_MIN_W {
				draw.Draw(drw, r, imgUni, image.ZP, draw.Src)
				frameDx = float32(SEEKER_MIN_W)
			} else {
				for i := range frameCount {
					frameRect.Min.X = int(frameSx * float32(i))
					frameRect.Max.X = int(frameSx * float32(i) + frameDx)
					draw.Draw(drw, frameRect, imgUni, image.ZP, draw.Src)
				}
			}

			//
			// draw cursor
			//

			imgUni = image.NewUniform(colorTheme.SeekerCursor)

			cursorW := Min(int(frameDx), int(SEEKER_W))
			cursorOffset := image.Point{int(frameDx/2) - cursorW/2, 0}
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
						fmt.Println(cursorPos)
						seekerChanged <- cursorPos
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
							fmt.Println(cursorPos)
							needRedraw = true
							seekerChanged <- cursorPos
							time.Sleep(time.Second / 60)
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
		}
	}

}


func Renderer(env gui.Env, aniToggle <-chan bool, framesCh chan<- int, cursorCh chan int, seekerChanged <-chan int,
		r image.Rectangle, simulation *tg.Simulation) {

	drawFrame := func(i int) func(draw.Image) image.Rectangle {
		if i > 0 {
			return func(drw draw.Image) image.Rectangle {
				draw.Draw(drw, r, simulation.Frames[i], image.ZP, draw.Src)
				return r
			}
		}

		return func(drw draw.Image) image.Rectangle {
			img := image.NewUniform(color.RGBA{0,0,0,255})
			draw.Draw(drw, r, img, image.ZP, draw.Src)
			return r
		}
	}

	running := false
	step := 0

	env.Draw() <- drawFrame(step)

	for {
		select{
		case <- aniToggle:
			running = !running
		case frameNumber := <- seekerChanged:
			step = frameNumber
			env.Draw() <- drawFrame(step)
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
					r image.Rectangle, colorTheme *Theme, simulation *tg.Simulation) {


	env.Draw() <- func(drw draw.Image) image.Rectangle {
		img := image.NewUniform(colorTheme.Background)
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