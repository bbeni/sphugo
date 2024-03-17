/* K-nearest-neighbor search Example

Author: Benjamin Frölich

Goal:
	Implement the k nearest neighbor search. Use the priority queue given in the
	Python template and implement “replace” and “key” functions.
	Use the particle to cell distance function from the lecture notes
	or the celldist2() given in the Python template. Are they the same?

	Optional: Also implement the ball search algorithm given in the lecture notes.

TODO: find out how include heap methods from own package
TODO: extract logic from tree-partition-2d and import

*/

package main

import (
	"github.com/bbeni/treego/tg"
	"math"
)


type PrioQValue tg.Cell

type PrioQ struct {
	keys []float32
	values []PrioQValue
}

type PrioQer interface {
	Len() int
	Less(i, j int) bool
	Swap(i, j int)
	MinKey() float32
	Replace()
}

func (pq PrioQ) Len() int {
	return len(pq.keys)
}

func (pq PrioQ) Less(i, j int) bool {
	return pq.keys[i] < pq.keys[j]
}

func (pq PrioQ) Swap(i, j int) {
	pq.keys[i], pq.keys[j] = pq.keys[j], pq.keys[i]
	pq.values[i], pq.values[j] = pq.values[j], pq.values[i]
}

func BuildHeap(pq PrioQ) {
	if pq.Len() < 2 {
		return
	}
	for i := pq.Len()/2 - 1; i >= 0; i-- {
		Heapify(pq, i)
	}
}

func Heapify(pq PrioQ, i int) {
	for {
		l := i*2 + 1
		r := i*2 + 2
		min_index := i

		if l < pq.Len() && pq.Less(l, min_index) {
			min_index = l
		}

		if r < pq.Len() && pq.Less(l, min_index) {
			min_index = r
		}

		if min_index == i {
			break
		}

		pq.Swap(min_index, i)
		i = min_index
	}
}

func (pq PrioQ) FindMin() (value PrioQValue, ok bool) {
	if pq.Len() == 0 {
		var def PrioQValue
		return def, false
	}
	return pq.values[0], true
}

func main() {
	particles := make([]tg.Particle, 60)
	tg.InitUniformly(particles)

	root := tg.Cell{
		LowerLeft: tg.Vec2{0, 0},
		UpperRight: tg.Vec2{1, 1},
		Particles: particles[:],
	}

	root.Treebuild(tg.Vertical)
	root.BoundingBalls()


	w, h := 1000, 1000
	c := tg.NewCanvas(w, h)
	c.Clear(tg.BLACK)

	var PlotBalls func (root *tg.Cell)

	PlotBalls = func (root *tg.Cell) {
		if root.Upper == nil && root.Lower == nil {

			x := float32(root.Center.X*float64(w))
			y := float32(root.Center.Y*float64(h))
			r := float32(math.Sqrt(root.BMaxSquared)) * float32(w)

			c.DrawCircle(x, y, r, 1.0, tg.BLUE)
		}

		if root.Upper != nil {
			PlotBalls(root.Upper)
		}

		if root.Lower != nil {
			PlotBalls(root.Lower)
		}
	}


	PlotBalls(&root)

	for _, p := range particles {

		x, y := p.Pos.X*float64(w), p.Pos.Y*float64(h)
		c.DrawDisk(float32(x), float32(y), 2.4, tg.GREEN)
	}
	c.AsPNG("test.png")

}

/*
func main() {

	keys := make([]float32, 0)
	for i := range 21 {
		keys = append(keys, float32(i*17 % 10 + 10))
	}

	values := make([]PrioQValue, 21)

	var pq PrioQ
	pq.values = values
	pq.keys = keys

	tg.BuildHeap(keys)
	tg.DumpHeap(keys)

	BuildHeap(pq)
	tg.DumpHeap(pq.keys)




	fmt.Println("Hello K. Nearest Neighbors!")
}
*/