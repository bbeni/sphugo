package main

import (
	"math"

	"github.com/bbeni/sphugo/gx"
	"github.com/bbeni/sphugo/sim"
)

func calcAndDrawDensity(sph *sim.Simulation, kernel sim.Kernel, canvas *gx.Canvas, side, positionIndex int, colorRamp func(uint8) gx.Color) {
	// Calculate Nearest Neighbor Density Rho
	for i, _ := range sph.Root.Particles {
		p := &sph.Root.Particles[i]
		p.Rho = sim.Density2D(p, sph, kernel)
	}

	// draw it for all particles
	for _, particle := range sph.Root.Particles {
		offX, offY := (side*positionIndex)%canvas.W, (side*positionIndex)/canvas.W*side
		x := float32(particle.Pos.X)*float32(side) + float32(offX)
		y := float32(particle.Pos.Y)*float32(side) + float32(offY)

		colorFormula := float64(particle.Rho / 6000 * 255)
		color_index := uint8(math.Min(colorFormula, 255))

		//color := gx.ParaRamp(color_index)
		//color := gx.HeatRamp(color_index)
		//color := gx.ToxicRamp(color_index)
		//color := gx.RainbowRamp(color_index)
		color := colorRamp(color_index)

		if color_index > 255 {
			nnRadius := float32(particle.NNDists[0]) * float32(canvas.W)
			canvas.DrawCircle(x, y, nnRadius, 2, gx.WHITE)
		}

		canvas.DrawDisk(float32(x), float32(y), 4, color)
	}
}

func main() {
	kernelT := sim.TopHat2D
	kernelM := sim.Monahan2D
	kernelW := sim.Wendtland2D

	sph := sim.Simulation{
		Config: sim.MakeConfig(),
	}

	spawner1 := sim.UniformRectSpawner{
		UpperLeft:  sim.Vec2{0, 0},
		LowerRight: sim.Vec2{1, 1},
		NParticles: 1000,
	}

	spawner2 := sim.UniformRectSpawner{
		UpperLeft:  sim.Vec2{0.1, 0},
		LowerRight: sim.Vec2{0.3, 0.4},
		NParticles: 200,
	}

	ps := spawner1.Spawn(0)
	ps = append(ps, spawner2.Spawn(0)...)

	sph.Root = sim.MakeCells(ps, sim.Vertical)
	sph.Root.Treebuild(sim.Vertical)
	sph.Root.BoundingSpheres()

	// claculate all nearest neighbours
	for i, _ := range sph.Root.Particles {
		sph.Root.Particles[i].FindNearestNeighboursPeriodic(sph.Root, [2]float64{0, 1}, [2]float64{0, 1})
	}

	side := 420
	w := 3 * side
	h := 2 * side

	canvas := gx.NewCanvas(w, h)
	canvas.Clear(gx.BLACK)

	calcAndDrawDensity(&sph, kernelT, &canvas, side, 0, gx.HeatRamp)
	calcAndDrawDensity(&sph, kernelM, &canvas, side, 1, gx.HeatRamp)
	calcAndDrawDensity(&sph, kernelW, &canvas, side, 2, gx.HeatRamp)

	calcAndDrawDensity(&sph, kernelT, &canvas, side, 3, gx.ParaRamp)
	calcAndDrawDensity(&sph, kernelM, &canvas, side, 4, gx.ParaRamp)
	calcAndDrawDensity(&sph, kernelW, &canvas, side, 5, gx.ParaRamp)

	// separation lines in white
	canvas.DrawLine(gx.Vec2i{side, 0}, gx.Vec2i{side, h}, gx.WHITE)
	canvas.DrawLine(gx.Vec2i{2 * side, 0}, gx.Vec2i{2 * side, h}, gx.WHITE)
	canvas.DrawLine(gx.Vec2i{0, side}, gx.Vec2i{w, side}, gx.WHITE)

	canvas.ToPNG("density_compare.png")

	periodicVisualTest()
}

func periodicVisualTest() {

	kernel := sim.TopHat2D

	sph := sim.Simulation{
		Config: sim.MakeConfig(),
	}

	spawner1 := sim.UniformRectSpawner{
		UpperLeft:  sim.Vec2{0.1, 0.1},
		LowerRight: sim.Vec2{0.9, 0.9},
		NParticles: 10000,
	}

	spawner2 := sim.UniformRectSpawner{
		UpperLeft:  sim.Vec2{0.85, 0.4},
		LowerRight: sim.Vec2{0.9, 0.9},
		NParticles: 1200,
	}

	ps := spawner1.Spawn(0)
	ps = append(ps, spawner2.Spawn(0)...)

	fname := [2]string{"density_test.png", "density_test_periodic.png"}
	for i := 0; i < 2; i++ {
		sph.Root = sim.MakeCells(ps, sim.Vertical)
		sph.Root.Treebuild(sim.Vertical)
		sph.Root.BoundingSpheres()

		// claculate all nearest neighbours
		if i == 0 {
			for i, _ := range sph.Root.Particles {
				sph.Root.Particles[i].FindNearestNeighbours(sph.Root)
			}
		} else {
			for i, _ := range sph.Root.Particles {
				sph.Root.Particles[i].FindNearestNeighboursPeriodic(sph.Root, [2]float64{0.1, 0.9}, [2]float64{0.1, 0.9})
			}
		}

		for i, _ := range sph.Root.Particles {
			p := &sph.Root.Particles[i]
			p.Rho = sim.Density2D(p, &sph, kernel)
		}

		// draw it for all particles
		canvas := gx.NewCanvas(700, 350)
		canvas.Clear(gx.BLACK)
		for _, particle := range sph.Root.Particles {
			x := float32(particle.Pos.X) * float32(canvas.W)
			y := float32(particle.Pos.Y) * float32(canvas.H)

			colorFormula := float64(particle.Rho / 32000 * 255)
			color_index := uint8(math.Min(colorFormula, 255))

			color := gx.ToxicRamp(color_index)
			//color := gx.RainbowRamp(color_index)

			if color_index > 255 {
				nnRadius := float32(particle.NNDists[0]) * float32(canvas.W)
				canvas.DrawCircle(x, y, nnRadius, 2, gx.WHITE)
			}

			canvas.DrawDisk(float32(x), float32(y), 2, color)
		}

		canvas.ToPNG(fname[i])
	}

}
