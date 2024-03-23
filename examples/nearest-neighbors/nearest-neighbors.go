/* K-nearest-neighbor search Example

Author: Benjamin Frölich

Goal:
	Implement the k nearest neighbor search. Use the priority queue given in the
	Python template and implement “replace” and “key” functions.
	Use the particle to cell distance function from the lecture notes
	or the celldist2() given in the Python template. Are they the same?
	Optional: Also implement the ball search algorithm given in the lecture notes.

*/

package main

import (
	"github.com/bbeni/sphugo/sim"
	"github.com/bbeni/sphugo/gfx"
)

func NonPeriodic() {
	root := sim.MakeCellsUniform(220, sim.Vertical)
	root.BoundingSpheres()

	w, h := 1000, 1000
	canvas := gfx.NewCanvas(w, h)
	canvas.Clear(gfx.BLACK)

	//sim.PlotBoundingCircles(canvas, root, 1, gfx.WHITE)
	sim.PlotCells(canvas, root, 1, 1)

	for _, p := range root.Particles {
		x, y := p.Pos.X*float64(w), p.Pos.Y*float64(h)
		canvas.DrawDisk(float32(x), float32(y), 3.4, gfx.ORANGE)
	}

	// pick one Particle and plot it
	p0 := root.Particles[14]
	x, y := p0.Pos.X*float64(w), p0.Pos.Y*float64(h)
	canvas.DrawDisk(float32(x), float32(y), 10, gfx.GREEN)


	// Find the nearest neighbors of the picked particle and plot them
	p0.FindNearestNeighbours(root)
	for i := range sim.NN_SIZE {
		pn := *p0.NearestNeighbours[i]
		x, y := pn.Pos.X*float64(w), pn.Pos.Y*float64(h)
		canvas.DrawDisk(float32(x), float32(y), 4.4, gfx.GREEN)
	}

	// Draw green circle
	radius := float32(p0.NNDists[0]*float64(w))
	canvas.DrawCircle(float32(x), float32(y), radius, 2, gfx.GREEN)

	canvas.ToPNG("nearest_neighbours.png")
}

func Periodic() {
	root := sim.MakeCellsUniform(220, sim.Vertical)
	root.BoundingSpheres()

	w, h := 1000, 1000
	canvas := gfx.NewCanvas(w, h)
	canvas.Clear(gfx.BLACK)

	sim.PlotBoundingCircles(canvas, root, 1, gfx.WHITE)

	for _, p := range root.Particles {
		x, y := p.Pos.X*float64(w), p.Pos.Y*float64(h)
		canvas.DrawDisk(float32(x), float32(y), 3.4, gfx.ORANGE)
	}

	// pick one Particle and plot it
	p0 := root.Particles[14]
	x, y := p0.Pos.X*float64(w), p0.Pos.Y*float64(h)
	canvas.DrawDisk(float32(x), float32(y), 10, gfx.GREEN)


	// Find the nearest neighbors of the picked particle and plot them
	p0.FindNearestNeighboursPeriodic(root)
	for i := range sim.NN_SIZE {
		pn := *p0.NearestNeighbours[i]
		x, y := pn.Pos.X*float64(w), pn.Pos.Y*float64(h)
		canvas.DrawDisk(float32(x), float32(y), 4.4, gfx.GREEN)
	}

	// Draw green circles periodic
	for i := -1.0; i<=1; i++ {
		for j := -1.0; j<=1; j++ {
			radius := float32(p0.NNDists[0]*float64(w))
			pixel_x := float32(x) + float32(float64(w)*i)
			pixel_y := float32(y) + float32(float64(h)*j)
			canvas.DrawCircle(pixel_x, pixel_y, radius, 2, gfx.GREEN)
		}
	}

	canvas.ToPNG("nearest_neighbours_periodic.png")
}

func main() {
	NonPeriodic()
	Periodic()
}
