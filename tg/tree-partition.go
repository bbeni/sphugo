package tg

import (
	"os"
	"fmt"
	"math/rand"
	"log"
	"image"
	"image/png"
	"image/color"
	"time"
)

// Configuration
const (
	N_PARTICLES = 2200
	MAX_PARTICLES_PER_CELL = 8
	SPLIT_FRACTION = 0.5       // Fraction of left to total space for Treebuild(), usually 0.5.
	USE_RANDOM_SEED = false    // for generating randomly distributed particles in init_uniformly()
)

// Image generation config
const (
	IMAGE_W = 512*2
	IMAGE_H = 512*2
	RECT_OFFSET = 1  // Pixel offset for upper right corner in negative x and y direction
	TREE_PNG_FNAME = "tree.png"
)

// Math functions for int
func Abs(x int) int {
	if x < 0 { return -x }
	return x
}

func Max(x, y int) int {
	if x >= y {return x}
	return y
}

func Min(x, y int) int {
	if x <= y {return x}
	return y
}

// Linear algebra
type Vec2i struct {
	X, Y int
}

type Vec2 struct {
	X, Y float64
}


type Particle struct {
	Pos, Vel Vec2
	Rho float64
}

type Cell struct {
	LowerLeft Vec2
	UpperRight Vec2
	Particles []Particle
	lower *Cell
	upper *Cell
}

type Orientation int
const (
	Vertical Orientation = iota
	Horizontal
)

func (orientation Orientation) other() (Orientation) {
	if orientation == Vertical { return Horizontal }
	return Vertical
}

func InitUniformly(particles []Particle) {

	if USE_RANDOM_SEED {
		rand.Seed(time.Now().UnixNano())
	} else {
	    rand.Seed(12345678)
	}

	for i, _ := range particles {
		particles[i].Pos = Vec2{rand.Float64(), rand.Float64()}
		particles[i].Pos = Vec2{rand.Float64(), rand.Float64()}
		particles[i].Rho = 1
	}
}


/* The function Partition() partitions an array of type Particle based
 on their 2d position. They are compared to a pivot value called middle
 in a "bubble sort like" manner in a specified axis that can either be
 "Vertical" or "Horizontal". The tests should cover most edge cases.
 Returns two partitioned slices a, b (just indices of array in Go).*/

func Partition (ps []Particle, orientation Orientation, middle float64) (a, b []Particle) {
	i := 0
	j := len(ps) - 1

	if orientation == Vertical {
		for i < j {
			for i < j && ps[i].Pos.Y <= middle { i++ }
			for i < j && ps[j].Pos.Y > middle { j-- }

			if ps[i].Pos.Y > ps[j].Pos.Y {
				ps[i], ps[j] = ps[j], ps[i]
			}
			if i == j && middle > ps[i].Pos.Y {i++}
		}
	} else {
		for i < j {
			for i < j && ps[i].Pos.X <= middle { i++ }
			for i < j && ps[j].Pos.X > middle { j-- }

			if ps[i].Pos.X > ps[j].Pos.X {
				ps[i], ps[j] = ps[j], ps[i]
			}
		}
		if i == j && middle > ps[i].Pos.X {i++}
	}
	return ps[:i], ps[i:]
}


/*The function Treebuild() recurses and partitions an array of
 N_PARTICLES length int Cells that have maximally
 MAX_PARTICLES_PER_CELL particles.
 The SPLIT_FRACTION determines the fraction of space in
 the specific direction for left/total or top/total. */

func (root *Cell) Treebuild (orientation Orientation) {
	
	var mid float64
	if orientation == Vertical {
		mid = SPLIT_FRACTION * root.LowerLeft.Y + (1 - SPLIT_FRACTION) * root.UpperRight.Y
	} else {
		mid = SPLIT_FRACTION * root.LowerLeft.X + (1 - SPLIT_FRACTION) * root.UpperRight.X
	}
	
	a, b := Partition(root.Particles, orientation, mid)
	
	if len(a) > 0 {
		root.lower = &Cell{
			Particles: a,
			LowerLeft: root.LowerLeft,
			UpperRight: root.UpperRight,
		}

		if orientation == Vertical {
			root.lower.UpperRight.Y = mid
		} else {
			root.lower.UpperRight.X = mid
		}

		if len(a) > MAX_PARTICLES_PER_CELL {
			root.lower.Treebuild(orientation.other())
		}
	}

	if len(b) > 0 {
		root.upper = &Cell{
			Particles: b,
			LowerLeft: root.LowerLeft,
			UpperRight: root.UpperRight,
		}

		if orientation == Vertical {
			root.upper.LowerLeft.Y = mid
		} else {
			root.upper.LowerLeft.X = mid
		}

		if len(b) > MAX_PARTICLES_PER_CELL {
			root.upper.Treebuild(orientation.other())
		}
	}
}

func (root *Cell) Dumptree(level int) {
	for i := 0; i < level; i++ { fmt.Print("  ") }
	fmt.Println(root.LowerLeft, root.UpperRight)
	if root.upper != nil {
		root.upper.Dumptree(level + 1)
	}
	if root.lower != nil {
		root.lower.Dumptree(level + 1)
	}
}

func (root *Cell) Countlevel() int {
	a, b := 0, 0
	if root.upper != nil { a = root.upper.Countlevel() }
	if root.lower != nil { b = root.lower.Countlevel() }
	return Max(a, b) + 1
}


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

	if root.lower != nil {
		draw_cells(img, w, h, root.lower, level + 1, max_level)
	}

	if root.upper != nil {
		draw_cells(img, w, h, root.upper, level + 1, max_level)
	}

	color_index := uint8(level*256/max_level)
	draw_quad(img, x1, y1, x2 - 1, y2 - 1, rainbow_ramp(color_index))
}

func MakeTreePng(particles []Particle, root* Cell) {

	const w, h = IMAGE_W, IMAGE_H
	black := color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	white := color.NRGBA{R: 255, G: 255, B: 255, A: 255}

	img := image.NewNRGBA(image.Rect(0, 0, w, h))

	// Black Background
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, black)
		}
	}

	// Draw the compartiment cells
	draw_cells(img, w, h, root, 0, root.Countlevel())

	// Draw all particles in white
	for _, particle := range particles {
		x := int(particle.Pos.X * w)
		y := int(particle.Pos.Y * w)
		img.Set(x, y, white)
	}

	// draw a rainbow on lower left corner
	for i := range 256 {
		draw_line(img, Vec2i{i, IMAGE_H}, Vec2i{i, IMAGE_H-10}, rainbow_ramp(uint8(i)))
	}

	file, err := os.Create(TREE_PNG_FNAME)
	if err != nil {
		log.Fatalf("Error: couldn't create file %q : %q", TREE_PNG_FNAME, err)
		os.Exit(1)
	}
	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		log.Fatalf("Error: couldn't encode PNG %q : %q", TREE_PNG_FNAME, err)
		os.Exit(1)
	}
}
