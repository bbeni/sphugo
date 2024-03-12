package main

import (
	"os"
	"fmt"
	"math/rand"
	"log"
	"image"
	"image/png"
	"image/color"
)

// Configuration
const (
	N_PARTICLES = 2200
	MAX_PARTICLES_PER_CELL = 8
	SPLIT_FRACTION = 0.5
)

// Image generation config
const (
	IMAGE_W = 512*2
	IMAGE_H = 512*2
	RECT_OFFSET = 1
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
	x, y int
}

type Vec2 struct {
	x, y float64
}


type Particle struct {
	pos, vel Vec2
	rho float64
}

type Cell struct {
	lower_left Vec2
	upper_right Vec2
	particles []Particle
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

func init_uniformly(particles []Particle) {

    //rand.Seed(time.Now().UnixNano())
    rand.Seed(12345678)

	for i, _ := range particles {
		particles[i].pos = Vec2{rand.Float64(), rand.Float64()}
		particles[i].pos = Vec2{rand.Float64(), rand.Float64()}
		particles[i].rho = 1
	}
}

func Partition (ps []Particle, orientation Orientation, middle float64) (a, b []Particle) {
	i := 0
	j := len(ps) - 1

	if orientation == Vertical {
		for i < j {
			for i < j && ps[i].pos.y <= middle { i++ }
			for i < j && ps[j].pos.y > middle { j-- }

			if ps[i].pos.y > ps[j].pos.y {
				ps[i], ps[j] = ps[j], ps[i]
			}
			if i == j && middle > ps[i].pos.y {i++}
		}
	} else {
		for i < j {
			for i < j && ps[i].pos.x <= middle { i++ }
			for i < j && ps[j].pos.x > middle { j-- }

			if ps[i].pos.x > ps[j].pos.x {
				ps[i], ps[j] = ps[j], ps[i]
			}
		}
		if i == j && middle > ps[i].pos.x {i++}
	}
	return ps[:i], ps[i:]
}

func (root *Cell) Treebuild (orientation Orientation) {
	
	var mid float64
	if orientation == Vertical {
		mid = SPLIT_FRACTION * root.lower_left.y + (1 - SPLIT_FRACTION) * root.upper_right.y
	} else {
		mid = SPLIT_FRACTION * root.lower_left.x + (1 - SPLIT_FRACTION) * root.upper_right.x
	}
	
	a, b := Partition(root.particles, orientation, mid)
	
	if len(a) > 0 {
		root.lower = &Cell{
			particles: a,
			lower_left: root.lower_left,
			upper_right: root.upper_right,
		}

		if orientation == Vertical {
			root.lower.upper_right.y = mid
		} else {
			root.lower.upper_right.x = mid
		}

		if len(a) > MAX_PARTICLES_PER_CELL {
			root.lower.Treebuild(orientation.other())
		}
	}

	if len(b) > 0 {
		root.upper = &Cell{
			particles: b,
			lower_left: root.lower_left,
			upper_right: root.upper_right,
		}

		if orientation == Vertical {
			root.upper.lower_left.y = mid
		} else {
			root.upper.lower_left.x = mid
		}

		if len(b) > MAX_PARTICLES_PER_CELL {
			root.upper.Treebuild(orientation.other())
		}
	}
}

func (root *Cell) Dumptree(level int) {
	for i := 0; i < level; i++ { fmt.Print("  ") }
	fmt.Println(root.lower_left, root.upper_right)
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

	delta_x := b.x - a.x
	delta_y := b.y - a.y

	step_x := 1
	if delta_x < 0 { step_x = -1 }

	step_y := 1
	if delta_y < 0 { step_y = -1 }

	if delta_x == 0 && delta_y == 0 {
		img.Set(a.x, a.y, color)
		return
	}

	if Abs(delta_x) >= Abs(delta_y) {
		for i := range (Abs(delta_x) + 1) {
			x := b.x - step_x * i
			y := b.y - int(float64(delta_y * i * step_x) / float64(delta_x))
			img.Set(x, y, color)
		}
	} else {
		for i := range (Abs(delta_y) + 1) {
			y := b.y - step_y * i
			x := b.x - int(float64(delta_x * i * step_y) / float64(delta_y))
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
	x1 := int(root.lower_left.x * float64(w))
	y1 := int(root.lower_left.y * float64(h))
	x2 := int(root.upper_right.x * float64(w))
	y2 := int(root.upper_right.y * float64(h))

	if root.lower != nil {
		draw_cells(img, w, h, root.lower, level + 1, max_level)
	}

	if root.upper != nil {
		draw_cells(img, w, h, root.upper, level + 1, max_level)
	}

	color_index := uint8(level*256/max_level)
	draw_quad(img, x1, y1, x2 - 1, y2 - 1, rainbow_ramp(color_index))
}

func make_tree_png(particles []Particle, root* Cell) {

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
		x := int(particle.pos.x * w)
		y := int(particle.pos.y * w)
		img.Set(x, y, white)
	}

	// draw a rainbow on lower left corner
	for i := range 256 {
		draw_line(img, Vec2i{i, IMAGE_H}, Vec2i{i, IMAGE_H-10}, rainbow_ramp(uint8(i)))
	}

	file, err := os.Create(TREE_PNG_FNAME)
	if err != nil {
		log.Fatal("Error: couldn't create file %q : %q", TREE_PNG_FNAME, err)
		os.Exit(1)
	}
	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		log.Fatal("Error: couldn't encode PNG %q : %q", TREE_PNG_FNAME, err)
		os.Exit(1)
	}
}


func test_case_logger(msg string) (func(passed bool, ps []Particle), func()) {
	n_passed := 0
	n_failed := 0
	return func(passed bool, ps []Particle) {
		if passed {
			fmt.Printf("Passed test - ok\n")
			n_passed++
		} else {
			fmt.Printf("Failed test - %q!\n\t got %v\n", msg, ps)
			n_failed++
		}
	}, func() {
		fmt.Println("=====================")
		fmt.Printf ("Test Summary:\n")
		fmt.Printf ("   Failed %v/%v tests\n", n_failed, n_passed + n_failed)
		fmt.Printf ("   Passed %v/%v tests\n", n_passed, n_passed + n_failed)
		fmt.Printf ("=====================\n")
	}
}

func test_cases() {

	tl, summary := test_case_logger("Partition()")

	ps := [...]Particle {
		{pos: Vec2{1.0, 1.0}},
		{pos: Vec2{0.9, 0.9}},
		{pos: Vec2{0.8, 0.8}},
		{pos: Vec2{0.7, 0.7}},
	}

    // func Partition (ps []Particle, orientation Orientation, middle float64) (a, b []Particle)

	a, b := Partition(ps[:], Vertical, 0.5)
	if len(a) != 0 { tl(false, a) } else { tl(true, a) }
	if len(b) != 4 { tl(false, b) } else { tl(true, b) }

	a, b = Partition(ps[:], Horizontal, 0.5)
	if len(a) != 0 { tl(false, a) } else { tl(true, a) }
	if len(b) != 4 { tl(false, b) } else { tl(true, b) }


	a, b = Partition(ps[:], Vertical, 0.85)
	if len(a) != 2 { tl(false, a) } else { tl(true, a) }
	if len(b) != 2 { tl(false, b) } else { tl(true, b) }

	a, b = Partition(ps[:], Horizontal, 0.85)
	if len(a) != 2 { tl(false, a) } else { tl(true, a) }
	if len(b) != 2 { tl(false, b) } else { tl(true, b) }


	ps1 := [...]Particle {
		{pos: Vec2{0.0, 0.9}},
		{pos: Vec2{0.5, -0.8}},
		{pos: Vec2{1.7, 0.1}},
		{pos: Vec2{0.7, -0.1}},
		{pos: Vec2{-0.7, 0.1}},
	}

	a, b = Partition(ps1[:], Vertical, 0.100000000001)
	if len(a) != 4 { tl(false, a) } else { tl(true, a) }
	if len(b) != 1 { tl(false, b) } else { tl(true, b) }

	a, b = Partition(ps1[:], Horizontal, 0.601)
	if len(a) != 3 { tl(false, a) } else { tl(true, a) }
	if len(b) != 2 { tl(false, b) } else { tl(true, b) }

	a, b = Partition(ps1[:], Vertical, -100)
	if len(a) != 0 { tl(false, a) } else { tl(true, a) }
	if len(b) != 5 { tl(false, b) } else { tl(true, b) }

	a, b = Partition(ps1[:], Horizontal, 100)
	if len(a) != 5 { tl(false, a) } else { tl(true, a) }
	if len(b) != 0 { tl(false, b) } else { tl(true, b) }


	ps1 = [...]Particle {
		{pos: Vec2{0.9,  0.0}},
		{pos: Vec2{-0.8,  0.5}},
		{pos: Vec2{0.1,  1.7}},
		{pos: Vec2{-0.1,  0.7}},
		{pos: Vec2{0.1, -0.7}},
	}

	a, b = Partition(ps1[:], Horizontal, 0.100000000001)
	if len(a) != 4 { tl(false, a) } else { tl(true, a) }
	if len(b) != 1 { tl(false, b) } else { tl(true, b) }

	a, b = Partition(ps1[:], Vertical, 0.601)
	if len(a) != 3 { tl(false, a) } else { tl(true, a) }
	if len(b) != 2 { tl(false, b) } else { tl(true, b) }

	a, b = Partition(ps1[:], Horizontal, -100)
	if len(a) != 0 { tl(false, a) } else { tl(true, a) }
	if len(b) != 5 { tl(false, b) } else { tl(true, b) }

	a, b = Partition(ps1[:], Vertical, 100)
	if len(a) != 5 { tl(false, a) } else { tl(true, a) }
	if len(b) != 0 { tl(false, b) } else { tl(true, b) }


	ps2 := [...]Particle {}
	a, b = Partition(ps2[:], Horizontal, 0.85)
	if len(a) != 0 { tl(false, a) } else { tl(true, a) }
	if len(b) != 0 { tl(false, b) } else { tl(true, b) }

	summary()
}

func main() {

	test_cases()

	var particles [N_PARTICLES]Particle
	init_uniformly(particles[:])

	root := Cell{
		lower_left: Vec2{0, 0},
		upper_right: Vec2{1, 1},
		particles: particles[:],
	}

	root.Treebuild(Vertical)
	//root.Dumptree(0)

	make_tree_png(root.particles[:], &root)
	fmt.Printf("Created %s", TREE_PNG_FNAME)

}