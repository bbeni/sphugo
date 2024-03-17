package tg

import (
	"fmt"
	"math/rand"
	"math"
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

func (v *Vec2) Add(other *Vec2) Vec2 {
	return Vec2{v.X + other.X, v.Y + other.Y}
}

func (v *Vec2) Sub(other *Vec2) Vec2 {
	return Vec2{v.X - other.X, v.Y - other.Y}
}

func (v *Vec2) Dot(other *Vec2) float64 {
	return v.X*other.X + v.Y*other.Y
}

func (v Vec2) Mul(f float64) Vec2 {
	return Vec2{v.X*f, v.Y*f}
}


type Particle struct {
	Pos, Vel Vec2
	Rho float64
}

type Cell struct {
	LowerLeft Vec2
	UpperRight Vec2
	Particles []Particle
	Lower *Cell
	Upper *Cell

	//BoundingBall
	Center Vec2
	BMaxSquared float64
}

func (cell *Cell) DistSquared(to *Vec2) float64 {
	d1 := to.Sub(&cell.UpperRight)
	d2 := cell.LowerLeft.Sub(to)
	maxx := math.Max(d1.X, d2.X)
	maxy := math.Max(d1.Y, d2.Y)
	return maxx*maxx + maxy*maxy
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
		root.Lower = &Cell{
			Particles: a,
			LowerLeft: root.LowerLeft,
			UpperRight: root.UpperRight,
		}

		if orientation == Vertical {
			root.Lower.UpperRight.Y = mid
		} else {
			root.Lower.UpperRight.X = mid
		}

		if len(a) > MAX_PARTICLES_PER_CELL {
			root.Lower.Treebuild(orientation.other())
		}
	}

	if len(b) > 0 {
		root.Upper = &Cell{
			Particles: b,
			LowerLeft: root.LowerLeft,
			UpperRight: root.UpperRight,
		}

		if orientation == Vertical {
			root.Upper.LowerLeft.Y = mid
		} else {
			root.Upper.LowerLeft.X = mid
		}

		if len(b) > MAX_PARTICLES_PER_CELL {
			root.Upper.Treebuild(orientation.other())
		}
	}
}

func (root *Cell) BoundingBalls() {
	if root.Upper == nil && root.Lower == nil {
		if len(root.Particles) == 1 {
			root.Center = root.Particles[0].Pos
			root.BMaxSquared = 0
		} else {
			dSquaredMax := 0.0
			var pA, pB Vec2

			// N^2 time
			// stupid: double checking
			for _, p1 := range root.Particles {
				for _, p2 := range root.Particles {
					x := p2.Pos.Sub(&p1.Pos)
					dSq := x.Dot(&x)
					if dSq > dSquaredMax {
						dSquaredMax = dSq
						pA, pB = p1.Pos, p2.Pos
					}
				}
			}

			// the vector that connects outermost points half
			rMax := pB.Sub(&pA).Mul(0.5)
			root.Center = rMax.Add(&pA)
			root.BMaxSquared = rMax.Dot(&rMax)
		}
	}

	if root.Upper != nil {
		root.Upper.BoundingBalls()
	}

	if root.Lower != nil {
		root.Lower.BoundingBalls()
	}
}


func (root *Cell) Dumptree(level int) {
	for i := 0; i < level; i++ { fmt.Print("  ") }
	fmt.Println(root.LowerLeft, root.UpperRight)
	if root.Upper != nil {
		root.Upper.Dumptree(level + 1)
	}
	if root.Lower != nil {
		root.Lower.Dumptree(level + 1)
	}
}

func (root *Cell) Countlevel() int {
	a, b := 0, 0
	if root.Upper != nil { a = root.Upper.Countlevel() }
	if root.Lower != nil { b = root.Lower.Countlevel() }
	return Max(a, b) + 1
}
