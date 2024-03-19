package tg

import (
	"fmt"
	"math/rand"
	"math"
	"time"
)

// Configuration
const (
	N_PARTICLES 			= 2200
	MAX_PARTICLES_PER_CELL  = 8
	SPLIT_FRACTION 			= 0.5   // Fraction of left to total space for Treebuild(), usually 0.5.
	USE_RANDOM_SEED 		= false // for generating randomly distributed particles in init_uniformly()
	NN_SIZE 				= 32    // Nearest Neighbour Size
)

// For 10 Mio particles
// 10 000 000 * 104 bytes ~~ 1.3 GB

type Particle struct {
	Pos Vec2
	Vel Vec2
	Rho float64   // Density
	C float64     // Speed of sound
	E float64     // Specific internal energy
	H float64     // NN parameter

	// Temporary values filled by Simulation
	EDot float64  // dE/dt
	VDot Vec2     // Acceleration
	EPred float64 // Predicted internal energy
	VPred Vec2    // Predicted Velicty
	// 104 bytes until now

	// @Speed might be bad because we have a big particle size now
	// should rather keep it seperate to prevent cache misses?
	// for now we just naively implement it like this
	// 32*(8+8)  = 512 bytes
	NearestNeighbours [NN_SIZE]*Particle
	NNDists 		  [NN_SIZE]float64

}

// Tree structure every leaf holds
// at most MAX_PARTICLES_PER_CELL
type Cell struct {
	Particles []Particle

	// Bounds of Cell
	LowerLeft  Vec2
	UpperRight Vec2

	// Minimum Bounding Sphere
	BCenter Vec2
	BRadius float64

	// Children
	Lower *Cell
	Upper *Cell
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
	}
	for i, _ := range particles {
		particles[i].Rho = rand.Float64()*6 + 1.0
		particles[i].Vel = Vec2{rand.Float64()*0.05, rand.Float64()*0.05}
	}

}


func MakeCellsUniform(nParticles int, ori Orientation) (*Cell) {
	particles := make([]Particle, nParticles)

	InitUniformly(particles[:])

	root := Cell{
		LowerLeft: Vec2{0, 0},
		UpperRight: Vec2{1, 1},
		Particles: particles[:],
	}

	root.Treebuild(ori)
	root.BoundingSpheres()

	return &root
}

func MakeCell(numberParticles int, initalizer func(index int) Vec2 ) (root *Cell) {
	panic("Not Implemented")
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

// TODO(#5): @Bug stackoverflow @Leak Memory Maybe memory not initalized
// when recomputing Treebuild ?
// mayne the bug is even in Partition ?
func (root *Cell) Treebuild (orientation Orientation) {
	

	// dirty fix: particles out of [0, 1] x [0, 1] will grow cells indefinitely
	// MAX_PARTICLES_PER_CELL not satisfied...
	if root.LowerLeft.Y == root.UpperRight.Y && root.LowerLeft.X == root.UpperRight.X {
		return
	}

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

		if len(a) > MAX_PARTICLES_PER_CELL{
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


// Adapted idea from: (might be worse)
// 1990, Jack Ritter proposed a simple algorithm to find a non-minimal bounding sphere.
// https://en.wikipedia.org/wiki/Bounding_sphere, 2024
func (root *Cell) BoundingSpheres() {
	if root.Upper == nil && root.Lower == nil {
		if len(root.Particles) == 1 {
			root.BCenter = root.Particles[0].Pos
			root.BRadius = 0
		} else {
			dSquaredMax := 0.0
			var pA, pB Vec2

			// naive idea
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
			root.BCenter = rMax.Add(&pA)
			BRadiusSq := rMax.Dot(&rMax)

			// step 3: include outliers
			// naive idea
			for _, p := range root.Particles {
				x := root.BCenter.Sub(&p.Pos)
				rNewSq := x.Dot(&x)
				if rNewSq > BRadiusSq {
					BRadiusSq = rNewSq
				}
			}

			root.BRadius = math.Sqrt(BRadiusSq)
		}
		return
	}


	if root.Upper != nil {
		root.Upper.BoundingSpheres()
	}

	if root.Lower != nil {
		root.Lower.BoundingSpheres()
	}

	// at max one is not nil!
	// if unbalanced, just copy the values from the one
	if root.Upper == nil {
		root.BRadius = root.Lower.BRadius
		root.BCenter = root.Lower.BCenter
		return
	}

	if root.Lower == nil {
		root.BRadius = root.Upper.BRadius
		root.BCenter = root.Upper.BCenter
		return
	}

	if root.Lower == nil || root.Upper == nil {
		panic("unrachable!")
	}

	// we are sure to have all leafs calculated here
	// calculating the bounding cricle C that encloses circles A and B
	if root.Upper != nil && root.Lower != nil {
		AB := root.Upper.BCenter.Sub(&root.Lower.BCenter)
		ABNorm := AB.Norm()

		rA := root.Lower.BRadius
		rB := root.Upper.BRadius
		rC := (rA + rB + ABNorm) * 0.5
		mid := AB.Mul((rB - rC) / ABNorm)

		root.BCenter = mid.Add(&root.Upper.BCenter)
		root.BRadius = rC
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

func (root *Cell) Depth() int {
	a, b := 0, 0
	if root.Upper != nil { a = root.Upper.Depth() }
	if root.Lower != nil { b = root.Lower.Depth() }
	return Max(a, b) + 1
}
