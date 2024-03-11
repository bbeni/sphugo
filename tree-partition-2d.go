package main

import (
	"fmt"
	"math/rand"
)

type Orientation int
const (
	Vertical Orientation = iota
	Horizontal
)

func (orientation Orientation) other() (Orientation) {
	if orientation == Vertical {
		return Horizontal
	} else {
		return Vertical
	}
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

func init_uniformly(particles []Particle) {
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
			if ps[i].pos.y > middle && ps[j].pos.y < middle {
				ps[i], ps[j] = ps[j], ps[i]
				j--
				i++
				continue
			}
			if ps[i].pos.y <= middle { i++ }
			if ps[j].pos.y >= middle { j-- }
		}
	} else {
		for i < j {
			if ps[i].pos.x > middle && ps[j].pos.x < middle {
				ps[i], ps[j] = ps[j], ps[i]
				j--
				i++
				continue
			}
			if ps[i].pos.x <= middle { i++ }
			if ps[j].pos.x >= middle { j-- }
		}
	}
	return ps[:i+1], ps[i:]
}

func Treebuild (root *Cell, orientation Orientation) {
	
	var mid float64
	if orientation == Vertical {
		mid = 0.5 * (root.lower_left.y + root.upper_right.y)
	} else {
		mid = 0.5 * (root.lower_left.x + root.upper_right.x)
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

		if len(a) > 8 {
			Treebuild(root.lower, orientation.other())
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

		if len(b) > 8 {
			Treebuild(root.upper, orientation.other())
		}
	}
}

func (root *Cell) Dump_tree(level int) {
	for i := 0; i < level; i++ { fmt.Print("  ") }
	fmt.Println(root.lower_left, root.upper_right)
	if root.upper != nil {
		root.upper.Dump_tree(level + 1)
	}
	if root.lower != nil {
		root.lower.Dump_tree(level + 1)
	}
}

func main() {
	var particles [100]Particle
	init_uniformly(particles[:])

	root := Cell{
		lower_left: Vec2{0, 0},
		upper_right: Vec2{1, 1},
		particles: particles[:],
	}

	Treebuild(&root, Vertical)
	root.Dump_tree(0)

}