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


const (
	PANEL_W    = 330
	RENDERER_W = 1280
	RENDERER_H = 720
	TEXT_SIZE  = 32
	SEEKER_H   = 40
	SEEKER_W   = 20
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
	BTN_H 		:= 100
	MARGIN_BOT	:= 4

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
		Background:  color.RGBA{18,   18,  18, 255},

		ButtonText:  color.RGBA{255, 250, 240, 255}, // Floral White
		ButtonUp:    color.RGBA{36,   33,  36, 255}, // Raisin Black
		ButtonHover: color.RGBA{45,   45,  45, 255},
		ButtonBlink: color.RGBA{100, 100, 100, 255},

		SeekerBackground:		color.RGBA{100,  0,  0, 255},
		SeekerFrame:	 		color.RGBA{150,  0,  0, 255},
		SeekerCursor:	 		color.RGBA{180, 30, 40, 255},

		FontFace:    fontFace,
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

	cmd  := make(chan string)
	play := make(chan string)
	mux, env := gui.NewMux(w)

	{
		xa := W - PANEL_W
		xb := W
		go Button(mux.MakeEnv(), "Run Simulation", colorTheme, image.Rect(xa, 0, xb, BTN_H), &fontMu, func() {
			cmd <- "."
		})
		go Button(mux.MakeEnv(), "Pause Simulation", colorTheme, image.Rect(xa, 1*(BTN_H+MARGIN_BOT), xb, 1*(BTN_H+MARGIN_BOT)+BTN_H), &fontMu, func() {
			cmd <- "."
		})
		go Button(mux.MakeEnv(), "Play/Pause", colorTheme, image.Rect(xa, 2*(BTN_H+MARGIN_BOT), xb, 2*(BTN_H+MARGIN_BOT)+BTN_H), &fontMu, func() {
			play <- "."
		})
		go Button(mux.MakeEnv(), "Render", colorTheme, image.Rect(xa, 3*(BTN_H+MARGIN_BOT), xb, 3*(BTN_H+MARGIN_BOT)+BTN_H), &fontMu, func() {

		})

	}


	simulation := tg.MakeSimulation()

	{
		go Simulator(mux.MakeEnv(), cmd, &simulation)
		go Renderer(mux.MakeEnv(), image.Rect(0, 0, RENDERER_W, RENDERER_H), play, &simulation)
		go Seeker(mux.MakeEnv(), image.Rect(0, RENDERER_H, RENDERER_W, RENDERER_H+SEEKER_H), colorTheme)

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

func Simulator(env gui.Env, cmd <-chan string, simulation *tg.Simulation) {

	simulation.Init()

	running := false
	step := 0

	for {
		select {
		case <-cmd:
			running = !running
		default:
			if running {
				simulation.Step(step)
				step += 1
			}
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

func Seeker(env gui.Env, r image.Rectangle, colorTheme *Theme) {

	env.Draw() <- func(drw draw.Image) image.Rectangle {
		imgUni := image.NewUniform(colorTheme.SeekerBackground)
		draw.Draw(drw, r, imgUni, image.ZP, draw.Src)

		imgUni = image.NewUniform(colorTheme.SeekerCursor)

		cursorRect := r
		cursorRect.Max.X = SEEKER_W
		draw.Draw(drw, cursorRect, imgUni, image.ZP, draw.Src)

		return r
	}

	for event := range env.Events() {
		switch event := event.(type) {
		default:
			fmt.Println(event)
		}
	}
}


func Renderer(env gui.Env, r image.Rectangle, play <-chan string, simulation *tg.Simulation) {

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
		case <-play:
			running = !running
		default:
			if running && simulation != nil {
				if step >= len(simulation.Frames) {
					step = 0
				}
				env.Draw() <- drawFrame(step)
				step += 1
				time.Sleep(time.Second / 60)
			}
		}
	}
}

func main() {
	mainthread.Run(run)
}