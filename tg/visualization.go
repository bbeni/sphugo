/* Visualization stuff for Simulations */
package tg

import (
	"github.com/bbeni/sphugo/gx"
)

func MakeTreePlot(root* Cell, w, h int) (gx.Canvas) {

	var canvas = gx.NewCanvas(w, h)
	canvas.Clear(gx.BLACK)

	// Draw the compartiment cells
	PlotCells(canvas, root, 0, root.Depth())

	// Draw all particles in white
	for _, particle := range root.Particles {
		x := int(particle.Pos.X * float64(canvas.W))
		y := int(particle.Pos.Y * float64(canvas.H))
		canvas.DrawPoint(x, y, gx.WHITE)
	}

	// draw a rainbow on lower left corner
	for i := range 256 {
		lower_left := gx.Vec2i{i, h}
		upper_right := gx.Vec2i{i, h-10}
		canvas.DrawLine(lower_left, upper_right, gx.RainbowRamp(uint8(i)))
	}
	return canvas
}

// draws rectangles for each cell. the higher the level,
// the color of the rect is more towards violet in rainbow
func PlotCells(canvas gx.Canvas, root *Cell, color_index, max_color_index int) {
	x1 := int(root.LowerLeft.X * float64(canvas.W))
	y1 := int(root.LowerLeft.Y * float64(canvas.H))
	x2 := int(root.UpperRight.X * float64(canvas.W))
	y2 := int(root.UpperRight.Y * float64(canvas.H))

	if root.Lower != nil {
		PlotCells(canvas, root.Lower, color_index + 1, max_color_index)
	}

	if root.Upper != nil {
		PlotCells(canvas, root.Upper, color_index + 1, max_color_index)
	}

	canvas.DrawRect(gx.Vec2i{x1, y1}, gx.Vec2i{x2 - 1, y2 - 1}, gx.RainbowRamp(uint8(color_index*256/max_color_index)))
}

// Draw Bounding Circles of leaf nodes in Cell
// max_depth should be >= 1
func PlotBoundingCircles (c gx.Canvas, root *Cell, max_depth int, color gx.Color) {

	// only plot up to max_depth
	if root.Depth() <= max_depth {

		x := float32(root.BCenter.X*float64(c.W))
		y := float32(root.BCenter.Y*float64(c.H))
		r := float32(root.BRadius)*float32(c.W) // incorrect: use width for now

		if r < 2 {
			r = 2
		}
		c.DrawCircle(x, y, r, 1.0, color)
	}

	if root.Upper != nil {
		PlotBoundingCircles(c, root.Upper, max_depth, color)
	}

	if root.Lower != nil {
		PlotBoundingCircles(c, root.Lower, max_depth, color)
	}
}