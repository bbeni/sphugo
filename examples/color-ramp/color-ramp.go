package main

import (
	"github.com/bbeni/sphugo/gx"
)

func main() {

	cmaps := []func(uint8) gx.Color{gx.RainbowRamp, gx.ParaRamp, gx.HeatRamp, gx.ToxicRamp}

	var width = 20

	c := gx.NewCanvas(width*len(cmaps), 256)

	for j, cmap := range cmaps {
		for i := range 256 {
			c.DrawRect(gx.Vec2i{j * width, 255 - i}, gx.Vec2i{(j + 1) * width, 255 - i}, cmap(uint8(i)))
		}
	}

	c.ToPNG("customColorMaps.png")
}
