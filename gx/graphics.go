/*
	Graphics related stuff

Implements badly some image drawing methods.
Mainly used for debugging for now.

Example code:

	canvas := gx.NewCanvas(300, 300)
	canvas.Clear(gx.BLACK)
	canvas.DrawCircle(100,100,50,10,gx.WHITE)
	canvas.DrawDisk(150,150,30,gx.RED)
	canvas.ToPNG("test.png")
*/
package gx

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
)

type Color = color.NRGBA

var (
	BLACK  = color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	WHITE  = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	RED    = color.NRGBA{R: 255, G: 0, B: 0, A: 255}
	GREEN  = color.NRGBA{R: 0, G: 255, B: 0, A: 255}
	BLUE   = color.NRGBA{R: 0, G: 0, B: 255, A: 255}
	ORANGE = color.NRGBA{R: 255, G: 165, B: 0, A: 255}

	SKYBLUE_PALE   = color.NRGBA{204, 255, 255, 255}
	SKYBLUE_OPAQUE = color.NRGBA{128, 204, 204, 255}
	SKYBLUE        = color.NRGBA{178, 255, 255, 255}
)

type Vec2i struct {
	X, Y int
}

type Canvas struct {
	Img *image.NRGBA
	W   int
	H   int
}

func NewCanvas(width, height int) Canvas {
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

	xa := int(math.Floor(float64(cx - radius)))
	xb := int(math.Ceil(float64(cx + radius)))
	ya := int(math.Floor(float64(cy - radius)))
	yb := int(math.Ceil(float64(cy + radius)))

	for x := xa; x <= xb; x++ {
		for y := ya; y <= yb; y++ {
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

	xa := int(math.Floor(float64(cx - radius)))
	xb := int(math.Ceil(float64(cx + radius)))
	ya := int(math.Floor(float64(cy - radius)))
	yb := int(math.Ceil(float64(cy + radius)))

	for x := xa; x <= xb; x++ {
		for y := ya; y <= yb; y++ {
			dx := float32(x) - cx
			dy := float32(y) - cy
			rSquared := dx*dx + dy*dy
			if rSquared <= radius*radius {
				c.DrawPoint(x, y, color)
			}
		}
	}
}

func (c Canvas) ToPNG(file_path string) {

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

	log.Printf("Created PNG %s.", file_path)
}

func Max(x, y int) int {
	if x >= y {
		return x
	}
	return y
}

func Min(x, y int) int {
	if x <= y {
		return x
	}
	return y
}

// index 0 to 255 gives a rainbow
func RainbowRamp(index uint8) Color {
	x := int(index)
	r := Max(Max(Min(255, 620-4*x), 0), 2*x-400)
	g := Max(Min(Min(255, 3*x), 820-4*x), 0)
	b := Min(Max(0, 4*x-620), 255)
	return color.NRGBA{
		R: uint8(r), G: uint8(g), B: uint8(b), A: 255,
	}
}

// index 0 to 255 gives a
// whatever map cheap remake of
// https://www.kennethmoreland.com/color-advice/
// ordered on wish
func ParaRamp(index uint8) Color {
	x := float64(index)
	r := math.Min(215, math.Max(0, -math.Abs((x-60)*31/17-232)+245))
	g := math.Min(190, math.Max(0, -math.Abs(x*31/17-232)+245))
	b := math.Min(215, math.Max(0, -math.Abs((x+70)*31/17-232)+245))
	return color.NRGBA{
		R: uint8(r), G: uint8(g), B: uint8(b), A: 255,
	}
}

// https://www.kennethmoreland.com/color-advice/
var blackbodyRGB = [17 * 3]uint8{
	0, 0, 0,
	36, 15, 9,
	62, 22, 17,
	90, 27, 22,
	119, 30, 26,
	150, 33, 30,
	180, 38, 34,
	197, 65, 28,
	214, 88, 19,
	228, 112, 7,
	231, 141, 18,
	233, 169, 29,
	233, 195, 39,
	231, 222, 50,
	246, 240, 144,
	255, 255, 255,
	255, 255, 255,
}

func HeatRamp(index uint8) Color {

	t := float32(index%16) / 16
	i := int(index) / 16

	r := float32(blackbodyRGB[i*3+3])*t + float32(blackbodyRGB[i*3])*(1-t)
	g := float32(blackbodyRGB[i*3+4])*t + float32(blackbodyRGB[i*3+1])*(1-t)
	b := float32(blackbodyRGB[i*3+5])*t + float32(blackbodyRGB[i*3+2])*(1-t)

	return color.NRGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}
}

// Kindlmann
// https://www.kennethmoreland.com/color-advice/
var kindlmannRGB = [17 * 3]uint8{
	0, 0, 0,
	37, 3, 57,
	37, 5, 109,
	24, 8, 163,
	8, 51, 160,
	6, 83, 127,
	5, 105, 105,
	6, 127, 83,
	7, 148, 47,
	15, 168, 8,
	63, 186, 9,
	133, 199, 10,
	205, 205, 10,
	251, 210, 163,
	253, 232, 223,
	255, 255, 255,
	255, 255, 255,
}

func ToxicRamp(index uint8) Color {

	t := float32(index%16) / 16
	i := int(index) / 16

	r := float32(kindlmannRGB[i*3+3])*t + float32(kindlmannRGB[i*3])*(1-t)
	g := float32(kindlmannRGB[i*3+4])*t + float32(kindlmannRGB[i*3+1])*(1-t)
	b := float32(kindlmannRGB[i*3+5])*t + float32(kindlmannRGB[i*3+2])*(1-t)

	return color.NRGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}
}

func (c Canvas) DrawLine(start, end Vec2i, color Color) {
	draw_line(c.Img, start, end, color)
}

func (c Canvas) DrawRect(lowerLeft, upperRight Vec2i, color Color) {
	draw_rect(c.Img, lowerLeft.X, lowerLeft.Y, upperRight.X, upperRight.Y, color)
}

func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func draw_line(img *image.NRGBA, a, b Vec2i, color color.NRGBA) {

	delta_x := b.X - a.X
	delta_y := b.Y - a.Y

	step_x := 1
	if delta_x < 0 {
		step_x = -1
	}

	step_y := 1
	if delta_y < 0 {
		step_y = -1
	}

	if delta_x == 0 && delta_y == 0 {
		img.Set(a.X, a.Y, color)
		return
	}

	if Abs(delta_x) >= Abs(delta_y) {
		for i := range Abs(delta_x) + 1 {
			x := b.X - step_x*i
			y := b.Y - int(float64(delta_y*i*step_x)/float64(delta_x))
			img.Set(x, y, color)
		}
	} else {
		for i := range Abs(delta_y) + 1 {
			y := b.Y - step_y*i
			x := b.X - int(float64(delta_x*i*step_y)/float64(delta_y))
			img.Set(x, y, color)
		}
	}
}

func draw_rect(img *image.NRGBA, x1, y1, x2, y2 int, color color.NRGBA) {
	draw_line(img, Vec2i{x1, y1}, Vec2i{x2, y1}, color)
	draw_line(img, Vec2i{x2, y1}, Vec2i{x2, y2}, color)
	draw_line(img, Vec2i{x2, y2}, Vec2i{x1, y2}, color)
	draw_line(img, Vec2i{x1, y2}, Vec2i{x1, y1}, color)
}
