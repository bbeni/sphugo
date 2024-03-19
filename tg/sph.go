package tg

import (
	"math"
	"log"
//	"time"
//	"fmt"
//	"github.com/bbeni/treego/gx"
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

	sim.NSteps = 200

	sim.DeltaTHalf = 0.01

	sim.Root = MakeCellsUniform(100000, Vertical)

	for i, _ := range sim.Root.Particles {
		sim.Root.Particles[i].Vel = Vec2{0.1, 0.01}
	}

	return sim
}

// SPH
func (sim *Simulation) Run() {

	// TODO(#2): check if simulation initialized

	// initialization drift dt=0
	for _, p := range sim.Root.Particles {
		p.VPred = p.Vel
		p.EPred = p.E
	}

	sim.CalculateForces()

	//canvas := gx.NewCanvas(1024, 1024)

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

		log.Printf("Calculated step %v/%v", sim.NSteps, step)

		/*canvas.Clear(gx.BLACK)

		//PlotBoundingCircles(canvas, sim.Root, 1, gx.GREEN)
		// Draw all particles in white
		for _, particle := range sim.Root.Particles {
			x := int(particle.Pos.X * float64(canvas.W))
			y := int(particle.Pos.Y * float64(canvas.H))
			canvas.DrawPoint(x, y, gx.WHITE)
		}

		canvas.ToPNG(fmt.Sprintf("./out/%v.png", step))

		*/
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
