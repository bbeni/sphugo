package main

import (
	"github.com/bbeni/sphugo/gfx"
)

func main() {

	cmaps := []func(uint8)(gfx.Color){gfx.RainbowRamp, gfx.ParaRamp, gfx.HeatRamp, gfx.ToxicRamp}

	var width = 20

	c := gfx.NewCanvas(width*len(cmaps), 256)

	for j, cmap := range cmaps {
		for i := range 256 {
			c.DrawRect(gfx.Vec2i{j*width,255-i}, gfx.Vec2i{(j+1)*width,255-i}, cmap(uint8(i)))
		}
	}

	c.ToPNG("customColorMaps.png")
}