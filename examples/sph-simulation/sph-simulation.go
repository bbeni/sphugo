package main

import (
	"github.com/bbeni/sphugo/sim"

	"bufio"
	"fmt"
	"os"
)

/* Generates all frames of the example SPH simulation defined in sim/sph.go.
The frames are stored as .png in ./out/

To make a video use FFMPEG with the following command:

'''
	ffmpeg -framerate 30 -s 1280x720 -pix_fmt rgba -pattern_type glob -i './out/*.png' -c:v libx264  -pix_fmt yuv420p ./animation.mp4
'''
*/

// TODO: make compression work fine
// Bad quality ...   ffmpeg -framerate 30 -pattern_type glob -i './out/*.png' -c:v libx264 -vb 2500k -c:a aac -ab 200k -pix_fmt yuv420p ./animation.mp4
//

func main() {

	if _, err := os.Stat("./out"); !os.IsNotExist(err) {

		fmt.Println("The folder ./out/ exists. Do you want to overwrite it? [y/n]")
		reader := bufio.NewReader(os.Stdin)
		char, _, err := reader.ReadRune()

		if err != nil {
			panic(err)
		}

		if char == 'y' || char == 'Y' {
			os.RemoveAll("out")
		} else {
			fmt.Println("Not overweriting 'out'. Aborting.")
			return
		}
	}

	err := os.Mkdir("out", 0755)
	if err != nil {
		panic(err)
	}

	sph := sim.MakeSimulation()
	animator := sim.MakeAnimator(&sph)

	for i := range 10000 {
		sph.Step()
		canvas := animator.CurrentFrame()
		out_file := fmt.Sprintf("./out/%.4v.png", i)
		canvas.ToPNG(out_file)
	}

}
