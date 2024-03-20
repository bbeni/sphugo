package tg

import (
	"math"
	"log"
 	"fmt"
	"github.com/bbeni/sphugo/gx"
)

type Simulation struct {
	Root *Cell
	DeltaTHalf float64
	NSteps int

	// Constants
	Gamma float64 // heat capacity ratio = 1+2/f
}


// TODO(#1): Make it more customizable
func MakeSimulation() (Simulation){

	var sim Simulation

	sim.Gamma = 1.6

	sim.NSteps = 10000

	sim.DeltaTHalf = 0.01

	sim.Root = MakeCellsUniform(10_000, Vertical)

	return sim
}




// SPH
func (sim *Simulation) Run() {

	// TODO(#2): check if simulation initialized


	// used to order particles acoording to z-value before rendering
	renderingParticleArray := make([]*Particle, len(sim.Root.Particles))
	for i, _ := range sim.Root.Particles {
		renderingParticleArray[i] = &sim.Root.Particles[i]
	}

	// initialization drift dt=0
	for _, p := range sim.Root.Particles {
		p.VPred = p.Vel
		p.EPred = p.E
	}

	sim.CalculateForces()

	canvas := gx.NewCanvas(1920, 1080)

	for step := range sim.NSteps {

		// drift 1 for leapfrog dt/2
		for i, _ := range sim.Root.Particles {
			p := &sim.Root.Particles[i]
			vdt    := p.Vel.Mul(sim.DeltaTHalf)
			p.Pos   = p.Pos.Add(&vdt)
			adt    := p.VDot.Mul(sim.DeltaTHalf)
			p.VPred = p.Vel.Add(&adt)
			p.EPred = p.E + p.EDot * sim.DeltaTHalf
		}

		sim.CalculateForces()

		// kick dt
		for i, _ := range sim.Root.Particles {
			p := &sim.Root.Particles[i]
			adt  := p.VDot.Mul(2*sim.DeltaTHalf)
			p.Vel = p.Vel.Add(&adt)
			p.E   = p.E + p.EDot*2*sim.DeltaTHalf
		}

		// drift 2 for leapfrog dt/2
		for i, _ := range sim.Root.Particles {
			p := &sim.Root.Particles[i]
			vdt    := p.Vel.Mul(sim.DeltaTHalf)
			p.Pos   = p.Pos.Add(&vdt)
		}

		// Periodic Boundary: particles outside boundary get moved back
		for i, _ := range sim.Root.Particles {
			p := &sim.Root.Particles[i]
			if p.Pos.X >= 1.05 {
				p.Pos.X -= 1.1
			}
			if p.Pos.Y >= 1.05 {
				p.Pos.Y -= 1.1
			}

			if p.Pos.X < -0.05 {
				p.Pos.X += 1.1
			}

			if p.Pos.Y < -0.05 {
				p.Pos.Y += 1.1
			}
		}

		log.Printf("Calculated step %v/%v", sim.NSteps, step)

		canvas.Clear(gx.BLACK)


		extractZindex := func(p *Particle) int {
			return -p.Z
		}

		QuickSort(renderingParticleArray, extractZindex)

		for _, particle := range renderingParticleArray{
			x := float32(particle.Pos.X) * float32(canvas.W)
			y := float32(particle.Pos.Y) * float32(canvas.H)

			zNormalized := float32(particle.Z)/float32(math.MaxInt)

			//color_index := 255 - uint8((particle.Rho - 1)*64)
			color_index := 255 - uint8(zNormalized * 256)
			color := gx.ToxicRamp(color_index)

			canvas.DrawDisk(float32(x), float32(y), zNormalized*zNormalized*20+1, color)
		}

		canvas.ToPNG(fmt.Sprintf("./out/%.4v.png", step))
	}

}

func (sim *Simulation) CalculateForces() {

	// rebuild the tree to perserve data locality
	sim.Root.Treebuild(Vertical)

	sim.Root.BoundingSpheres()

	// claculate all nearest neighbours
	for i, _ := range sim.Root.Particles {
		sim.Root.Particles[i].FindNearestNeighboursPeriodic(sim.Root)
	}

	// Calculate Nearest Neighbor Density Rho
	{
		// TODO(#3): implement NN density

	}

	// Calculate speed of sound
	// c = sqrt(gamma(gamma-1)ePred)
	// gamma heat capacity ratio = 1 + 2/f
	{
		factor := sim.Gamma * (sim.Gamma - 1)
		for i, _ := range sim.Root.Particles {
			c := math.Sqrt(factor * sim.Root.Particles[i].EPred)
			sim.Root.Particles[i].C = c
		}
	}

	// Calculate Nearest Neighbor SPH forces
	// VDot, EDot
	{
		// TODO(#4): implement NN forces
	}

}
