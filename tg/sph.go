package tg

import (
	"math"
	"log"
	"time"
 	"math/rand"
	"github.com/bbeni/sphugo/gx"
  //"fmt"
	"image"
	"sync"
)

type Simulation struct {
	Root *Cell
	DeltaTHalf float64
	NSteps int
	NParticles int

	// Constants
	Gamma float64 // heat capacity ratio = 1+2/f

	//frames for rendering
	Frames   []image.Image
	FramesMu sync.Mutex

	renderingParticleArray []*Particle
}


// TODO(#1): Make it more customizable
func MakeSimulation() (Simulation){

	var sim Simulation

	sim.Gamma = 1.8
	sim.NSteps = 1000
	sim.DeltaTHalf = 0.002
	sim.NParticles = 500

	particles := make([]Particle, sim.NParticles)
	//InitSpecial(particles)
	InitEvenly(particles)
	sim.Root = MakeCells(particles, Vertical)


	sim.Frames = make([]image.Image, 0, sim.NSteps)

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
		particles[i].E = 0.001
	}
}


func (sim *Simulation) Run() {

	sim.Init()

	for step := range sim.NSteps {
		sim.Step(step)
	}

}

func (sim *Simulation) Init() {
	// TODO(#2): check if simulation initialized
	if sim.Root == nil || len(sim.Root.Particles) == 0  {
		panic("int Run(): Simulation not initialized!")
	}

	// used to order particles acoording to z-value before rendering
	sim.renderingParticleArray = make([]*Particle, len(sim.Root.Particles))
	for i, _ := range sim.Root.Particles {
		sim.renderingParticleArray[i] = &sim.Root.Particles[i]
	}

	// initialization drift dt=0
	for _, p := range sim.Root.Particles {
		p.VPred = p.Vel
		p.EPred = p.E
	}

	sim.CalculateForces()
}

// SPH
func (sim *Simulation) Step(step int) {

	{
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
			if p.Pos.X >= 1 {
				p.Pos.X -= 1
			}
			if p.Pos.Y >= 1 {
				p.Pos.Y -= 1
			}

			if p.Pos.X < 0 {
				p.Pos.X += 1
			}

			if p.Pos.Y < 0 {
				p.Pos.Y += 1
			}
		}

		log.Printf("Calculated step %v/%v", step, sim.NSteps)


		// Rendering the frame in this block
		{
			canvas := gx.NewCanvas(1280, 720)

			canvas.Clear(gx.BLACK)

			extractZindex := func(p *Particle) int {

				//return int(p.Rho*100000)
				return -p.Z
			}

			QuickSort(sim.renderingParticleArray, extractZindex)

			for _, particle := range sim.renderingParticleArray{
				x := float32(particle.Pos.X) * float32(canvas.W)
				y := float32(particle.Pos.Y) * float32(canvas.H)

				//zNormalized := float32(particle.Z)/float32(math.MaxInt)
				//color_index := 255 - uint8((particle.Rho - 1)*64)
				//color_index := 255 - uint8(zNormalized * 256)

				color_index := uint8(math.Min(float64(particle.Rho/float64(sim.NParticles*3)*256), 255))
				color := gx.ParaRamp(color_index)
				//color := gx.HeatRamp(color_index)
				//color := gx.ToxicRamp(color_index)
				//color := gx.RainbowRamp(color_index)


				if color_index > 255 {
					nnRadius := float32(particle.NNDists[0])*float32(canvas.W)
					canvas.DrawCircle(x, y, nnRadius, 2, gx.WHITE)
				}

				//canvas.DrawDisk(float32(x), float32(y), zNormalized*zNormalized*20+1, color)
				canvas.DrawDisk(float32(x), float32(y), 14, color)
			}

			//canvas.ToPNG(fmt.Sprintf("./out/%.4v.png", step))

			img := canvas.Img

			//sim.FramesMu.Lock()
			sim.Frames = append(sim.Frames, img)
			//sim.FramesMu.Unlock()
		}

	}

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


// lets assume mass 1 per particle, so the density is just the 1/volume of sphere
func DensityMonahan2D(p *Particle) (float64) {
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

	return acc*6*40/(math.Pi*maxR*maxR*7)
}

// lets assume mass 1 per particle, so the density is just the 1/volume of sphere
// viscosity PIab not implemented
// - Sum [ (Pa/rhoa^2       + Pb/rhob^2     + PIab )]
//			contribution A  + contributionB
func AccelerationAndEDotMonahan2D(p *Particle, gammaFactor float64) {
	maxR := p.NNDists[0]

	contributionA := p.C*p.C / (gammaFactor*p.Rho)
	contributionB := 0.0

	nablaAKernel := 0.0

	acc_ax := 0.0
	acc_ay := 0.0
	acc_edot := 0.0

	var x float64
	var i int
	for i = range NN_SIZE {
		nn := p.NearestNeighbours[i]
		x = p.NNDists[i]/maxR 			// r/h in lecture

		if x > 1 || x < 0 {
			panic("unreachable")
		}

		if x < 0.5 {
			nablaAKernel = (3*x*x - 2*x) / p.NNDists[i]
		} else {
		 	nablaAKernel = -(1 - x)*(1 - x) / p.NNDists[i]
		}

		// clamp kernel
		clamp := 0.9
		nablaAKernel = math.Max(math.Min(nablaAKernel, clamp), -clamp)

		contributionB = nn.C*nn.C / (gammaFactor*nn.Rho)

		rX := (p.Pos.X - nn.Pos.X)
		rY := (p.Pos.Y - nn.Pos.Y)
		acc_ax += rX * (contributionA + contributionB) * nablaAKernel
		acc_ay += rY * (contributionA + contributionB) * nablaAKernel
		acc_edot += (rY * (p.VPred.X - nn.VPred.X) + rY * (p.VPred.Y - nn.VPred.Y)) * nablaAKernel

	}

	// clamp acceleration
	clamp := 0.9
	acc_ax = math.Max(math.Min(acc_ax, clamp), -clamp)
	acc_ay = math.Max(math.Min(acc_ay, clamp), -clamp)

	// clamp energy change
	acc_edot = math.Max(math.Min(acc_edot, 10), 0)

	acc := Vec2{acc_ax, acc_ay}
    acc = acc.Mul(-6*40/(math.Pi*maxR*maxR*maxR*7))
	p.VDot = acc
	p.EDot = acc_edot
	//fmt.Println(p.EDot, p.VDot, p.VPred, p.C)
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
			p.Rho = DensityMonahan2D(p)
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
			AccelerationAndEDotMonahan2D(&sim.Root.Particles[i], sim.Gamma)
		}
	}

}
