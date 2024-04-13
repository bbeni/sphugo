package main

import (
	"math"
	"github.com/bbeni/sphugo/sim"
	"github.com/bbeni/sphugo/gfx"
)

func calcAndDrawDensity(sph *sim.Simulation, kernel sim.Kernel, canvas *gfx.Canvas, side, positionIndex int, colorRamp func (uint8) gfx.Color) {
	// Calculate Nearest Neighbor Density Rho
	for i, _ := range sph.Root.Particles {
		p := &sph.Root.Particles[i]
		p.Rho = sim.Density2D(p, sph, kernel)
	}

	// draw it for all particles
	for _, particle := range sph.Root.Particles {
		offX, offY := (side*positionIndex) % canvas.W, (side*positionIndex) / canvas.W * side
		x := float32(particle.Pos.X) * float32(side) + float32(offX)
		y := float32(particle.Pos.Y) * float32(side) + float32(offY)

		colorFormula := float64(particle.Rho/6000 * 255)
		color_index := uint8(math.Min(colorFormula, 255))

		//color := gfx.ParaRamp(color_index)
		//color := gfx.HeatRamp(color_index)
		//color := gfx.ToxicRamp(color_index)
		//color := gfx.RainbowRamp(color_index)
		color   := colorRamp(color_index)

		if color_index > 255 {
			nnRadius := float32(particle.NNDists[0])*float32(canvas.W)
			canvas.DrawCircle(x, y, nnRadius, 2, gfx.WHITE)
		}

		canvas.DrawDisk(float32(x), float32(y), 4, color)
	}
}

func main() {
	kernelT := sim.TopHat2D
	kernelM := sim.Monahan2D
	kernelW := sim.Wendtland2D

	sph := sim.Simulation {
		Config: sim.MakeConfig(),
	}

	spawner1 := sim.UniformRectSpawner {
		UpperLeft:  sim.Vec2{0, 0},
		LowerRight: sim.Vec2{1, 1},
		NParticles: 1000,
	}

	spawner2 := sim.UniformRectSpawner {
		UpperLeft:  sim.Vec2{0.1, 0},
		LowerRight: sim.Vec2{0.3, 0.4},
		NParticles: 200,
	}

	sph.Particles = spawner1.Spawn(0)
	sph.Particles = append(sph.Particles, spawner2.Spawn(0)...)

	sph.Root = sim.MakeCells(sph.Particles, sim.Vertical)
	sph.Root.Treebuild(sim.Vertical)
	sph.Root.BoundingSpheres()

	// claculate all nearest neighbours
	for i, _ := range sph.Root.Particles {
		sph.Root.Particles[i].FindNearestNeighboursPeriodic(sph.Root, [2]float64{0, 1}, [2]float64{0, 1})
	}

	side := 420
	w := 3 * side
	h := 2 * side

	canvas := gfx.NewCanvas(w, h)
	canvas.Clear(gfx.BLACK)

	calcAndDrawDensity(&sph, kernelT, &canvas, side, 0, gfx.HeatRamp)
	calcAndDrawDensity(&sph, kernelM, &canvas, side, 1, gfx.HeatRamp)
	calcAndDrawDensity(&sph, kernelW, &canvas, side, 2, gfx.HeatRamp)

	calcAndDrawDensity(&sph, kernelT, &canvas, side, 3, gfx.ParaRamp)
	calcAndDrawDensity(&sph, kernelM, &canvas, side, 4, gfx.ParaRamp)
	calcAndDrawDensity(&sph, kernelW, &canvas, side, 5, gfx.ParaRamp)

	 // separation lines in white
	canvas.DrawLine(gfx.Vec2i{  side, 0}, gfx.Vec2i{  side, h}, gfx.WHITE)
	canvas.DrawLine(gfx.Vec2i{2*side, 0}, gfx.Vec2i{2*side, h}, gfx.WHITE)
	canvas.DrawLine(gfx.Vec2i{0,   side}, gfx.Vec2i{w,   side}, gfx.WHITE)

	canvas.ToPNG("density_compare.png")
}