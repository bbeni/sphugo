package sim

import (
	"math"
	"log"
	"time"
 	"math/rand"
    "fmt"
)

var _ = fmt.Print

type Simulation struct {

	// Per timestep values, change the course of the simulation
	Root *Cell // Tree structure for keeping track of spatial cells of particles
	Particles []Particle
	CurrentStep int

	// Constants
	Gamma float64 // heat capacity ratio = 1+2/f
	DeltaTHalf float64
	NSteps int
	ConstAcc Vec2
	ParticleMass float64
}


// TODO(#1): Make it more customizable
func MakeSimulation() (Simulation){

	var sim Simulation

	sim.Gamma = 1.6666
	sim.NSteps = 10000
	sim.DeltaTHalf = 0.006
	sim.ConstAcc = Vec2{0, 2.2}
	sim.ParticleMass = 0.5

	nParticles := 10000
	particles := make([]Particle, nParticles)
	//InitSpecial(particles)
	//InitEvenly(particles)
	//InitEvenlyVelGradient(particles)
	InitBlock(particles)

	sim.Particles = particles
	sim.Root = MakeCells(particles, Vertical)

	return sim
}

func InitSpecial(particles []Particle) {

	if USE_RANDOM_SEED {
		rand.Seed(time.Now().UnixNano())
	} else {
	    rand.Seed(12345678)
	}

	for i, _ := range particles {
		particles[i].Pos = Vec2{rand.Float64(), rand.Float64()}
	}

	for i, _ := range particles {
		particles[i].Z   = rand.Int()
		zNormalized := float64(particles[i].Z)/float64(math.MaxInt)
		particles[i].Vel = Vec2{rand.Float64()*0.02, -rand.Float64()*0.15}.Mul(1/zNormalized)
	}
}


func InitEvenlyVelGradient(particles []Particle) {
	if USE_RANDOM_SEED {
		rand.Seed(time.Now().UnixNano())
	} else {
	    rand.Seed(12345678)
	}

	for i, _ := range particles {
		particles[i].Pos = Vec2{rand.Float64(), rand.Float64()}
	}
	for i, _ := range particles {
		particles[i].Z   = rand.Int()
		particles[i].Vel = Vec2{0.05 + particles[i].Pos.X * 0.1, (rand.Float64() - 0.5)*0.1}
	}
}


func InitEvenly(particles []Particle) {
	if USE_RANDOM_SEED {
		rand.Seed(time.Now().UnixNano())
	} else {
	    rand.Seed(12345678)
	}

	for i, _ := range particles {
		particles[i].Pos = Vec2{rand.Float64(), rand.Float64()}
	}
	for i, _ := range particles {
		particles[i].Z = rand.Int()
		particles[i].E = 0.01
	}
}

func InitBlock(particles []Particle) {
	if USE_RANDOM_SEED {
		rand.Seed(time.Now().UnixNano())
	} else {
	    rand.Seed(12345678)
	}

	for i, _ := range particles {
		particles[i].Pos = Vec2{rand.Float64()/6, rand.Float64()}
	}
	for i, _ := range particles {
		particles[i].Z = rand.Int()
		particles[i].E = 0.01
	}
}




func (sim *Simulation) Run() {
	for step := range sim.NSteps {
		sim.Step()
		log.Printf("Calculated step %v/%v", step, sim.NSteps)
	}
}


// SPH
func (sim *Simulation) Step() {

	// step 0 of SPH needs special work done before real step is done
	if sim.CurrentStep == 0 {

		// TODO(#2): check if simulation initialized
		if sim.Root == nil || len(sim.Root.Particles) == 0  {
			panic("int Run(): Simulation not initialized!")
		}

		// initialization drift dt=0
		for _, p := range sim.Root.Particles {
			p.VPred = p.Vel
			p.EPred = p.E
		}

		sim.CalculateForces()
	}

	// real work done here
	{
		// drift 1 for leapfrog dt/2
		for i, _ := range sim.Root.Particles {
			p      := &sim.Root.Particles[i]

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

		// Boundary: particles outside boundary get moved back
		for i, _ := range sim.Root.Particles {
			p := &sim.Root.Particles[i]
			if p.Pos.X >= 1 {
				p.Pos.X -= 1

				// refelction kind
				// and set pos back to almost 1
				p.Vel.X *= -1
				p.Pos.X  = 0.99999
			}
			if p.Pos.Y >= 1 {
				// periodic
				// p.Pos.Y -= 1

				// refelction kind
				// and set pos back to almost 1
				p.Vel.Y *= -1
				p.Pos.Y  = 0.99999
			}

			if p.Pos.X < 0 {
				p.Pos.X += 1

				// refelction kind
				// and set pos back to almost 1
				p.Vel.X *= -1
				p.Pos.X  = 0.01
			}

			if p.Pos.Y < 0 {

				//p.Pos.Y += 1

			}
		}


		/*for i, _ := range sim.Root.Particles {
			p := &sim.Root.Particles[i]

			if p.Pos.X >= 1 {
				p.Vel.X = -p.Vel.X
				p.Pos.X = -p.Pos.X
			}

			if p.Pos.Y >= 1 {
				p.Vel.Y = -p.Vel.Y*0.8
				p.Pos.Y = -p.Pos.Y
			}

			if p.Pos.X < 0 {
				p.Vel.X = -p.Vel.X
				p.Pos.X = -p.Pos.X
			}

			//if p.Pos.Y < 0 {
			//	p.Vel.Y = -p.Vel.Y*0.9
			//}
		}*/
	}

	sim.CurrentStep += 1
}


// lets assume mass 1 per particle, so the density is just the 1/volume of sphere
func DensityTopHat3D(p *Particle) (float64) {
	maxR := p.NNDists[0]
	return 3 * NN_SIZE / (4*math.Pi*maxR*maxR*maxR)
}


// lets assume mass 1 per particle, so the density is just the 1/volume of sphere
func DensityMonahan3D(p *Particle) (float64) {
	maxR := p.NNDists[0]

	acc := 0.0
	var x float64

	var i int
	for i = range NN_SIZE {
		x = p.NNDists[i]/maxR

		if x > 1 || x < 0 {
			panic("unreachable")
		}

		if x < 0.5 {
			acc += x*x*x - x*x + 1.0/6
			continue
		}
		acc += (1 - x)*(1 - x)*(1 - x) / 3
	}

	return acc * 6 * 8 / (math.Pi*maxR*maxR*maxR)
}


// lets assume mass 1 per particle, so the density is just the 1/volume of sphere
func DensityTopHat2D(p *Particle) (float64) {
	maxR := p.NNDists[0]
	return  NN_SIZE / (math.Pi*maxR*maxR)
}


func DensityMonahan2D(p *Particle, sim *Simulation) (float64) {
	maxR := p.NNDists[0]

	acc := 0.0
	var x float64

	var i int
	for i = range NN_SIZE {
		x = p.NNDists[i]/maxR

		if x > 1 || x < 0 {
			panic("unreachable")
		}

		if x < 0.5 {
			acc += x*x*x - x*x + 1.0/6
			continue
		}
		acc += (1 - x)*(1 - x)*(1 - x) / 3
	}

	return sim.ParticleMass*acc*6*40/(math.Pi*maxR*maxR*7)
}


// - Sum [ (Pa/rhoa^2       + Pb/rhob^2     + PIab )]
//			contribution A  + contributionB
func AccelerationAndEDotMonahan2D(p *Particle, sim *Simulation) {
	maxR := p.NNDists[0]

	// PA / rhoA^2
	contributionA := p.C*p.C / (sim.Gamma*p.Rho)
	contributionB := 0.0

	dRKernel := 0.0

	acc_ax := 0.0
	acc_ay := 0.0
	acc_edot := 0.0

	var q float64
	var i int
	for i = range NN_SIZE {
		nn := p.NearestNeighbours[i]
		q   = p.NNDists[i]/maxR 			// r/h in lecture

		if q > 1 || q < 0 {
			panic("kernel parameter q not in [0, 1]!")
		}

		//
		// Kernel
		//

		// TODO: make it exchangable easily
		if q < 0.5 {
			dRKernel = (3*q*q - 2*q)
		} else {
		 	dRKernel = -(1 - q)*(1 - q)
		}

		// PB / rhoB^2
		contributionB = nn.C*nn.C / (sim.Gamma*nn.Rho)


		vA := p.VPred
		vB := nn.VPred
		//vA := p.Vel
		//vB := nn.Vel

		rA := p.Pos
		rB := p.NNPos[i]

		//
		// Viscosity Term
		//

		vAB   := vB.Sub(&vA)
		rAB   := rB.Sub(&rA)
		dot   := vAB.Dot(&rAB)
		piAB  := 0.0
		if dot < 0 {
			const (
				alpha = 0.75
				beta  = 1.5
				etaSq = 0.01
			)
			cAB   := 0.5 * (p.C + nn.C)
			rhoAB := 0.5 * (p.Rho + p.Rho)
			hAB   := 0.5 * (p.NNDists[0] + nn.NNDists[0])
			muAB  := dot * hAB / (rAB.Dot(&rAB) + etaSq)
			piAB  = (-alpha*cAB*muAB + beta*muAB*muAB) / rhoAB
		}



		if rAB.X > 0.5 || rAB.Y > 0.5 {
			panic("rX or rY is bigger than expected. more than half boundary")
		}

		acc_ax += rAB.X * (piAB + contributionA + contributionB) * dRKernel / p.NNDists[i]
		acc_ay += rAB.Y * (piAB + contributionA + contributionB) * dRKernel / p.NNDists[i]
		acc_edot += dot * dRKernel
	}

	acc := Vec2{acc_ax, acc_ay}
    acc = acc.Mul(6*40*sim.ParticleMass/(math.Pi*maxR*maxR*maxR*7))
    acc = acc.Add(&sim.ConstAcc)
	p.VDot = acc
	p.EDot = contributionA * acc_edot * sim.ParticleMass // Benz formulation
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
		for i, _ := range sim.Root.Particles {
			p := &sim.Root.Particles[i]
			//p.Rho = DensityTopHat2D(p)
			p.Rho = DensityMonahan2D(p, sim)
			//fmt.Println(p.Rho)
		}
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
		for i, _ := range sim.Root.Particles {
			AccelerationAndEDotMonahan2D(&sim.Root.Particles[i], sim)
		}
	}

}

//
// profiling functions
//


func (sim *Simulation) TotalEnergy() float64 {
	tot := 0.0
	for i := range sim.Particles {
		tot += sim.Particles[i].E
	}
	return tot
}

func (sim *Simulation) TotalDensity() float64 {
	tot := 0.0
	for i := range sim.Particles {
		tot += sim.Particles[i].Rho
	}
	return tot
}

func (sim *Simulation) TotalMomentum() float64 {
	tot := 0.0
	for i := range sim.Particles {
		tot = sim.Particles[i].Vel.Norm()
	}
	return tot
}