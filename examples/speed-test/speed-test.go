package main

import (
	"github.com/bbeni/sphugo/sim"
	"fmt"
	"time"
)



func main() {

	spwn := sim.MakeUniformRectSpawner()
	spwn.NParticles = 10000

	conf := sim.MakeConfig()
	conf.Start = append(conf.Start, spwn)
	conf.DeltaTHalf = 0.02
	conf.Acceleration = sim.Vec2{0, 0.2}

	sph := sim.MakeSimulationFromConf(conf)

	previous := time.Now()
	total := 0.0

	for i := range 200 {
		sph.Step()

		elapsed := time.Since(previous).Seconds()
		previous = time.Now()

		total += elapsed
		fmt.Println("Step", i, "FPS", 1/elapsed)
	}

	fmt.Printf("Took %.4v seconds, and got an average FPS of %.4v", total, 200/total)
}