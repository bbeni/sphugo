package tg

import (
	"math"
	"log"
)

type Simulation struct {
	Root Cell
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

	sim.Root = MakeCellsUniform(10_000, Vertical)

	return sim
}

// SPH
func (sim Simulation) Run() {

	// TODO(#2): check if simulation initialized

	// initialization drift dt=0
	for _, p := range sim.Root.Particles {
		p.VPred = p.Vel
		p.EPred = p.E
	}

	sim.CalculateForces()

	for step := range sim.NSteps {

		// drift 1 for leapfrog dt/2
		for _, p := range sim.Root.Particles {
			vdt    := p.Vel.Mul(sim.DeltaTHalf)
			p.Pos   = p.Pos.Add(&vdt)
			adt    := p.VDot.Mul(sim.DeltaTHalf)
			p.VPred = p.Vel.Add(&adt)
			p.EPred = p.E + p.EDot * sim.DeltaTHalf
		}

		sim.CalculateForces()

		// kick dt
		for _, p := range sim.Root.Particles {
			adt  := p.VDot.Mul(2*sim.DeltaTHalf)
			p.Vel = p.Vel.Add(&adt)
			p.E   = p.E + p.EDot*2*sim.DeltaTHalf
		}

		// drift 2 for leapfrog dt/2
		for _, p := range sim.Root.Particles {
			vdt    := p.Vel.Mul(sim.DeltaTHalf)
			p.Pos   = p.Pos.Add(&vdt)
		}

		log.Printf("Calculated step %v/%v", sim.NSteps, step)
	}

}

func (sim Simulation) CalculateForces() {

	// rebuild the tree to perserve data locality
	sim.Root.Treebuild(Vertical)

	// Calculate Nearest Neighbor Density Rho
	{
		// TODO: implement NN density
	}

	// Calculate speed of sound
	// c = sqrt(gamma(gamma-1)ePred)
	// gamma heat capacity ratio = 1 + 2/f
	{
		factor := sim.Gamma * (sim.Gamma - 1)
		for _, p := range sim.Root.Particles {
			math.Sqrt(factor * p.EPred)
		}
	}

	// Calculate Nearest Neighbor SPH forces
	// VDot, EDot
	{
		// TODO: implement NN forces
	}

}
