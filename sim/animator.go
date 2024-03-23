package sim

import (
	"image"
	"math"

	"github.com/bbeni/sphugo/gfx"
	//"sync"
)


type Animator struct {

	//frames for rendering
	Frames   []image.Image
	//FramesMu sync.Mutex

	// as refernce to simulation variables
	Simulation *Simulation

	// reference to particles
	// also used to order particles acoording to z-value before rendering
	renderingParticleArray []*Particle

	// TODO: Gui state, should be moved to simviewer.go!
	ActiveFrame int
}

func MakeAnimator(simulation *Simulation) Animator {

	if simulation.Root == nil || len(simulation.Root.Particles) == 0  {
		panic("int Run(): Simulation not initialized!")
	}

	ani := Animator {
		renderingParticleArray: make([]*Particle, len(simulation.Root.Particles)),
	}

	for i, _ := range simulation.Particles {
		ani.renderingParticleArray[i] = &simulation.Particles[i]
	}

	ani.Simulation = simulation
	ani.Frames = make([]image.Image, 0, simulation.NSteps)

	ani.ActiveFrame = -1

	return ani
}


// render the last frame of the simultion
func (ani *Animator) Frame() {

	//
	// order according to z-value
	//

	extractZindex := func(p *Particle) int {
		//return int(p.Rho*100000)
		return -p.Z
	}
	QuickSort(ani.renderingParticleArray, extractZindex)

	canvas := gfx.NewCanvas(1280, 720)
	canvas.Clear(gfx.BLACK)

	for _, particle := range ani.renderingParticleArray {
		x := float32(particle.Pos.X) * float32(canvas.W)
		y := float32(particle.Pos.Y) * float32(canvas.H)

		//zNormalized := float32(particle.Z)/float32(math.MaxInt)
		//color_index := 255 - uint8((particle.Rho - 1)*64)
		//color_index := 255 - uint8(zNormalized * 256)

		color_index := uint8(math.Min(float64(particle.Rho/float64(len(ani.Simulation.Particles)*3)*256), 255))
		//color := gfx.ParaRamp(color_index)
		//color := gfx.HeatRamp(color_index)
		color := gfx.ToxicRamp(color_index)
		//color := gfx.RainbowRamp(color_index)


		if color_index > 255 {
			nnRadius := float32(particle.NNDists[0])*float32(canvas.W)
			canvas.DrawCircle(x, y, nnRadius, 2, gfx.WHITE)
		}

		//canvas.DrawDisk(float32(x), float32(y), zNormalized*zNormalized*20+1, color)
		canvas.DrawDisk(float32(x), float32(y), 8, color)
	}

	//canvas.ToPNG(fmt.Sprintf("./out/%.4v.png", sim.Step))

	img := canvas.Img

	//ani.FramesMu.Lock()
	ani.Frames = append(ani.Frames, img)
	//ani.FramesMu.Unlock()
}
