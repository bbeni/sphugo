package sim

import (
	"math"
	"log"
	//"time"
    "fmt"
    "sync"
)

var _ = fmt.Print

type Simulation struct {

	Config SphConfig

	Root *Cell // Tree structure for keeping track of spatial cells of particles
	Particles []Particle
	CurrentStep int

	IsBusy  sync.Mutex
}

func MakeSimulation() (Simulation){

	sim := Simulation {
		Config: MakeConfig(),
	}

	spawner := MakeUniformRectSpawner()
	sim.Particles = spawner.Spawn(0)
	sim.Root = MakeCells(sim.Particles, Vertical)

	return sim
}

func MakeSimulationFromConfig(configFilePath string) (error, Simulation) {
	err, conf:= MakeConfigFromFile(configFilePath)
	if err != nil {
		return err, MakeSimulation()
	}
	sim := MakeSimulationFromConf(conf)
	return nil, sim
}

func MakeSimulationFromConf(conf SphConfig) (Simulation){
	sim := Simulation {
		Config: conf,
	}

	ps := make([]Particle, 0, 100000)

	for _, startSpawner := range sim.Config.Start {
		ps = append(ps, startSpawner.Spawn(0)...)
	}

	sim.Particles = ps
	sim.Root = MakeCells(sim.Particles, Vertical)

	return sim
}

func (sim *Simulation) Run() {
	for step := range sim.Config.NSteps {
		sim.Step()
		log.Printf("Calculated step %v/%v", step, sim.Config.NSteps)
	}
}

// SPH
func (sim *Simulation) Step() {

	sim.IsBusy.Lock()

	// constants
	dtHalf := sim.Config.DeltaTHalf

	// sources spawn particles
	{
		t := float64(sim.CurrentStep) * dtHalf * 2

		i := -1
		for i = range sim.Config.Sources {
			spwn := &sim.Config.Sources[i]
			newParticles := (*spwn).Spawn(t)
			sim.Particles = append(sim.Particles, newParticles...)
		}

		if i != -1 {
			sim.Root = MakeCells(sim.Particles, Vertical)
		}

	}

	// step 0 of SPH needs special work done before real step is done
	if sim.CurrentStep == 0 {

		// TODO(#2): check if simulation initialized
		if sim.Root == nil || len(sim.Root.Particles) == 0  {
			panic("int Run(): Simulation not initialized!")
		}

		// initialization drift dt=0
		for i, p := range sim.Root.Particles {
			sim.Root.Particles[i].VPred = p.Vel
			sim.Root.Particles[i].EPred = p.E
		}

		sim.CalculateForces()
	}

	// real work done here
	{
		// drift 1 for leapfrog dt/2
		for i, _ := range sim.Root.Particles {
			p      := &sim.Root.Particles[i]

			vdt    := p.Vel.Mul(dtHalf)
			p.Pos   = p.Pos.Add(&vdt)

			adt    := p.VDot.Mul(dtHalf)
			p.VPred = p.Vel.Add(&adt)
			p.EPred = p.E + p.EDot * dtHalf
		}

		sim.CalculateForces()

		// kick dt
		for i, _ := range sim.Root.Particles {
			p := &sim.Root.Particles[i]
			adt  := p.VDot.Mul(2*dtHalf)
			p.Vel = p.Vel.Add(&adt)
			p.E   = p.E + p.EDot*2*dtHalf
		}

		// drift 2 for leapfrog dt/2
		for i, _ := range sim.Root.Particles {
			p := &sim.Root.Particles[i]

			vdt    := p.Vel.Mul(dtHalf)
			p.Pos   = p.Pos.Add(&vdt)
		}

		// Boundary: particles outside boundary get moved around
		//  x1              x2
		// x
		// x ->   x2-x1 + x
		//
		//  x1              x2
		//                     x
		// x -> -(x1-x2) + x
		//

		for i, _ := range sim.Root.Particles {
			p := &sim.Root.Particles[i]
			if p.Pos.X < sim.Config.HorPeriodicity[0] {
				p.Pos.X += (sim.Config.HorPeriodicity[1] - sim.Config.HorPeriodicity[0])
				continue
			}

			if p.Pos.X > sim.Config.HorPeriodicity[1] {
				p.Pos.X -= (sim.Config.HorPeriodicity[1] - sim.Config.HorPeriodicity[0])
				continue
			}

			if p.Pos.Y < sim.Config.VertPeriodicity[0] {
				p.Pos.Y += (sim.Config.VertPeriodicity[1] - sim.Config.VertPeriodicity[0])
				continue
			}

			if p.Pos.Y > sim.Config.VertPeriodicity[1] {
				p.Pos.Y -= (sim.Config.VertPeriodicity[1] - sim.Config.VertPeriodicity[0])
			}
		}

		// TODO: unhardcode refelction boundaries
		for i, _ := range sim.Root.Particles {
			p := &sim.Root.Particles[i]
			if p.Pos.Y > 0.94 {
				p.Pos.Y -= p.Pos.Y - 0.94
				p.Vel.Y = -p.Vel.Y
			}
		}
	}

	sim.CurrentStep += 1
	sim.IsBusy.Unlock()
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


type Kernel struct {
	F func(q float64) float64
	FPrefactor float64
	DF func(q float64) float64
	DFPrefactor float64
}

var TopHat2D = Kernel {
	F: func(q float64) float64 {
		return 1
	},

	FPrefactor: 1 / math.Pi,

	DF: func(q float64) float64 {
		panic("not defined. derivative is delta distribution!")
	},

	DFPrefactor: 1,
}

var Monahan2D = Kernel {
	F: func(q float64) float64 {
		if q < 0.5 {
			return q*q*q - q*q + 1.0/6
		}
		return (1 - q)*(1 - q)*(1 - q) / 3
	},

	FPrefactor:  6 * 40 / (math.Pi * 7),

	DF: func(q float64) float64 {
		if q < 0.5 {
			return (3*q*q - 2*q)
		}
	 	return -(1 - q)*(1 - q)
	},

	DFPrefactor: 6 * 40 / (math.Pi * 7),
}

// change varibales from lecture to adapt for factor of 2 in h
// q -> 2*q
// 1/h -> 2/rmax
// 2d F  prefactor -> 4 * ..
// 2d DF prefactor -> 8 * ..

var Wendtland2D = Kernel {
	F: func(q float64) float64 {
		if q <= 1 {
			return (1 - q)*(1 - q)*(1 - q)*(1 - q)*(1 + 4*q)
		} else {
			panic("not good..")
		}
	},

	FPrefactor: 4 * 7 / (math.Pi * 4),

	DF: func(q float64) float64 {
		if q <= 1 {
			return -10 * q * (1 - q)*(1 - q)*(1 - q)
		} else {
			panic("not good..")
		}
	},

	DFPrefactor: 8 * 7 / (math.Pi * 4),
}


func Density2D(p *Particle, sim *Simulation, kernel Kernel) (float64) {
	maxR := p.NNDists[0]

	acc := 0.0
	var x float64

	var i int
	for i = range NN_SIZE {
		x = p.NNDists[i]/maxR

		if x > 1 || x < 0 {
			panic("unreachable")
		}
		acc += kernel.F(x)
	}

	return kernel.FPrefactor*sim.Config.ParticleMass*acc / (maxR*maxR)
}

// - Sum [ (Pa/rhoa^2       + Pb/rhob^2     + PIab )]
//			contribution A  + contributionB
func AccelerationAndEDot2D(p *Particle, sim *Simulation, kernel Kernel) {
	gamma := sim.Config.Gamma
	maxR  := p.NNDists[0]

	// PA / rhoA^2
	contributionA := p.C*p.C / (gamma*p.Rho)
	contributionB := 0.0

	dRKernel := 0.0

	acc_ax := 0.0
	acc_ay := 0.0
	acc_edot := 0.0

	var q float64
	var i int
	for i = range NN_SIZE {
		nn := p.NearestNeighbours[i]

		// TODO: stupid fix because NN are nil sometimes...
		if nn == nil {
			break
		}

		q   = p.NNDists[i]/maxR 			// r/h in lecture

		if q > 1 || q < 0 {
			panic("kernel parameter q not in [0, 1]!")
		}

		dRKernel = kernel.DF(q)


		// PB / rhoB^2
		contributionB = nn.C*nn.C / (gamma*nn.Rho)

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
			rhoAB := 0.5 * (p.Rho + nn.Rho)
			hAB   := 0.5 * (p.NNDists[0] + nn.NNDists[0])
			muAB  := dot * hAB / (rAB.Dot(&rAB) + etaSq)
			piAB  = (-alpha*cAB*muAB + beta*muAB*muAB) / rhoAB
		}

		acc_ax += rAB.X * (piAB + contributionA + contributionB) * dRKernel / p.NNDists[i]
		acc_ay += rAB.Y * (piAB + contributionA + contributionB) * dRKernel / p.NNDists[i]
		acc_edot += dot * dRKernel
	}

	acc := Vec2{acc_ax, acc_ay}
    acc = acc.Mul(sim.Config.ParticleMass * kernel.DFPrefactor / (maxR*maxR*maxR))
    acc = acc.Add(&sim.Config.Acceleration)
	p.VDot = acc
	p.EDot = contributionA * acc_edot * sim.Config.ParticleMass // Benz formulation
}



func (sim *Simulation) CalculateForces() {

	// rebuild the tree to perserve data locality
	sim.Root.Treebuild(Vertical)

	sim.Root.BoundingSpheres()

	// claculate all nearest neighbours
	for i, _ := range sim.Root.Particles {

		// !!!
		// TODO: is the boundary a bug? we should actually use the config values in the boundary!!!
		// !!!

		sim.Root.Particles[i].FindNearestNeighboursPeriodic(sim.Root, [2]float64{-1, 2}, [2]float64{-1, 2})
	}

	// Calculate Nearest Neighbor Density Rho
	for i, _ := range sim.Root.Particles {
		p := &sim.Root.Particles[i]
		p.Rho = Density2D(p, sim, sim.Config.Kernel)
	}


	// Calculate speed of sound
	// c = sqrt(gamma(gamma-1)ePred)
	// gamma heat capacity ratio = 1 + 2/f
	{
		factor := sim.Config.Gamma * (sim.Config.Gamma - 1)
		for i, _ := range sim.Root.Particles {
			c := math.Sqrt(factor * sim.Root.Particles[i].EPred)
			sim.Root.Particles[i].C = c
		}
	}

	// Calculate Nearest Neighbor SPH forces
	// VDot, EDot
	for i, _ := range sim.Root.Particles {
		AccelerationAndEDot2D(&sim.Root.Particles[i], sim, sim.Config.Kernel)
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