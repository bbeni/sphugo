package main

import (
	"fmt"
	"github.com/bbeni/sphugo/sim"
	"os"
	"runtime/pprof"
	"time"
)

func main() {

	f, err := os.Create("myprogram.prof")
	if err != nil {
		fmt.Println(err)
		return
	}

	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	spwn := sim.MakeUniformRectSpawner()
	spwn.NParticles = 100000

	conf := sim.MakeConfig()
	conf.Start = append(conf.Start, spwn)
	conf.DeltaTHalf = 0.02
	conf.Acceleration = sim.Vec2{0, 0.2}

	sph := sim.MakeSimulationFromConf(conf)

	previous := time.Now()
	total := 0.0

	for i := range 20 {
		sph.Step()

		elapsed := time.Since(previous).Seconds()
		previous = time.Now()

		total += elapsed
		fmt.Println("Step", i, "FPS", 1/elapsed)
	}

	fmt.Printf("Took %.4v seconds, and got an average FPS of %.4v", total, 20/total)
}
