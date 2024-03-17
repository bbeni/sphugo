/* Graphics related stuff

Example code:

	canvas := tg.NewCanvas(300, 300)
	canvas.Clear(tg.BLACK)
	canvas.DrawCircle(100,100,50,10,tg.WHITE)
	canvas.DrawDisk(150,150,30,tg.RED)
	canvas.AsPNG("test.png")

*/
package tg

import (
	"os"
	"log"
	"image"
	"image/png"
	"image/color"
)

type Color = color.NRGBA

var (
	BLACK = color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	WHITE = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	RED = color.NRGBA{R: 255, G: 0, B: 0, A: 255}
	GREEN = color.NRGBA{R: 0, G: 255, B: 0, A: 255}
	BLUE = color.NRGBA{R: 0, G: 0, B: 255, A: 255}
)

type Canvas struct {
	Img *image.NRGBA
	W int
	H int
}

func NewCanvas(width, height int) (Canvas) {
	var c Canvas
	c.Img = image.NewNRGBA(image.Rect(0, 0, width, height))
	c.W = width
	c.H = height
	return c
}

func (c Canvas) Clear(color Color) {
	for y := 0; y < c.H; y++ {
		for x := 0; x < c.W; x++ {
			c.Img.Set(x, y, color)
		}
	}
}

func (c Canvas) DrawPoint(x, y int, color Color) {
	c.Img.Set(x, y, color)
}

func (c Canvas) DrawCircle(cx, cy, radius, border float32, color Color) {
	// stupid approach...
	for y := range c.H {
		for x:= range c.W {
			dx := float32(x) - cx
			dy := float32(y) - cy
			rSquared := dx*dx + dy*dy
			if rSquared >= radius*radius && rSquared <= (radius+border)*(radius+border) {
				c.DrawPoint(x, y, color)
			}
		}
	}
}

func (c Canvas) DrawDisk(cx, cy, radius float32, color Color) {
	// stupid approach...
	for y := range c.H {
		for x:= range c.W {
			dx := float32(x) - cx
			dy := float32(y) - cy
			rSquared := dx*dx + dy*dy
			if rSquared <= radius*radius {
				c.DrawPoint(x, y, color)
			}
		}
	}
}

func (c Canvas) AsPNG(file_path string) {

	file, err := os.Create(file_path)
	if err != nil {
		log.Fatalf("Error: couldn't create file %q : %q", file_path, err)
		os.Exit(1)
	}
	defer file.Close()

	err = png.Encode(file, c.Img)
	if err != nil {
		log.Fatalf("Error: couldn't encode PNG %q : %q", file_path, err)
		os.Exit(1)
	}
}

func MakeTreePng(particles []Particle, root* Cell) {

	var canvas = NewCanvas(IMAGE_W, IMAGE_H)
	canvas.Clear(BLACK)

	img := canvas.Img
	w, h := IMAGE_W, IMAGE_H

	// Draw the compartiment cells
	draw_cells(img, w, h, root, 0, root.Countlevel())

	// Draw all particles in white
	for _, particle := range particles {
		x := int(particle.Pos.X * float64(w))
		y := int(particle.Pos.Y * float64(w))
		img.Set(x, y, WHITE)
	}

	// draw a rainbow on lower left corner
	for i := range 256 {
		draw_line(img, Vec2i{i, IMAGE_H}, Vec2i{i, IMAGE_H-10}, rainbow_ramp(uint8(i)))
	}

	canvas.AsPNG(TREE_PNG_FNAME)
}



//
// TODO: factor out stuff to have pure graphics part here
//

// index 0 to 255 gives a rainbow
func rainbow_ramp(index uint8) (color.NRGBA){
	x := int(index)
	r := Max(Max(Min(255, 620 - 4 * x), 0), 2 * x - 400)
	g := Max(Min(Min(255, 3 * x), 820 - 4 * x), 0)
	b := Min(Max(0, 4 * x - 620), 255)
	return color.NRGBA{
		R: uint8(r), G: uint8(g), B: uint8(b), A: 255,
	}
}

func draw_line(img *image.NRGBA, a, b Vec2i, color color.NRGBA) {

	delta_x := b.X - a.X
	delta_y := b.Y - a.Y

	step_x := 1
	if delta_x < 0 { step_x = -1 }

	step_y := 1
	if delta_y < 0 { step_y = -1 }

	if delta_x == 0 && delta_y == 0 {
		img.Set(a.X, a.Y, color)
		return
	}

	if Abs(delta_x) >= Abs(delta_y) {
		for i := range (Abs(delta_x) + 1) {
			x := b.X - step_x * i
			y := b.Y - int(float64(delta_y * i * step_x) / float64(delta_x))
			img.Set(x, y, color)
		}
	} else {
		for i := range (Abs(delta_y) + 1) {
			y := b.Y - step_y * i
			x := b.X - int(float64(delta_x * i * step_y) / float64(delta_y))
			img.Set(x, y, color)
		}
	}
}

func draw_quad(img *image.NRGBA, x1, y1, x2, y2 int, color color.NRGBA) {
	draw_line(img, Vec2i{x1, y1}, Vec2i{x2, y1}, color)
	draw_line(img, Vec2i{x2, y1}, Vec2i{x2, y2}, color)
	draw_line(img, Vec2i{x2, y2}, Vec2i{x1, y2}, color)
	draw_line(img, Vec2i{x1, y2}, Vec2i{x1, y1}, color)
}

func draw_cells(img *image.NRGBA, w, h int, root *Cell, level, max_level int) {
	x1 := int(root.LowerLeft.X * float64(w))
	y1 := int(root.LowerLeft.Y * float64(h))
	x2 := int(root.UpperRight.X * float64(w))
	y2 := int(root.UpperRight.Y * float64(h))

	if root.Lower != nil {
		draw_cells(img, w, h, root.Lower, level + 1, max_level)
	}

	if root.Upper != nil {
		draw_cells(img, w, h, root.Upper, level + 1, max_level)
	}

	color_index := uint8(level*256/max_level)
	draw_quad(img, x1, y1, x2 - 1, y2 - 1, rainbow_ramp(color_index))
}