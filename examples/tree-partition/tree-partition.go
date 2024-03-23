/* Tree code and Partition Example

Author: Benjamin Frölich

For more info and implementation details see tg/tree-partition.go

*/

package main

import (
	"github.com/bbeni/sphugo/sim"
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

func main() {

	var particles [N_PARTICLES]sim.Particle
	sim.InitUniformly(particles[:])

	root := sim.Cell{
		LowerLeft: sim.Vec2{0, 0},
		UpperRight: sim.Vec2{1, 1},
		Particles: particles[:],
	}

	root.Treebuild(sim.Vertical)
	//root.Dumptree(0)

	canvas := sim.MakeTreePlot(&root, IMAGE_W, IMAGE_H)
	canvas.ToPNG(TREE_PNG_FNAME)

}