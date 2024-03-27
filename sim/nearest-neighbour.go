package sim

import (
	"math"
	"fmt"
)

// TODO: remove
var _ = fmt.Print


// recursively find all nearest neighbors of a particle based on position
// make sure you call it on the top/root Cell!!
//
// TODO: implement it using a loop
func (sim *Simulation) FindNearestNeighbours() {

	sim.NNQueueInitSentinel()

	for i := range sim.Particles {
		sim.findNNRec(sim.Root, i, Vec2{0, 0})
	}

	// make dists use Sqrt
	// and set h value
	for i := range sim.Particles {
		for j := range NN_SIZE {
			sim.NNDists[i][j] = math.Sqrt(sim.NNDists[i][j])
		}
		sim.Particles[i].h = sim.NNDists[i][0]
	}
}

// Periodic version
// assuming particles are between x = (0, 1], y = (0, 1]
func (sim *Simulation) FindNearestNeighboursPeriodic() {

	sim.NNQueueInitSentinel()

	for k := range sim.Particles {
		for i:=-1.0; i<=1; i++ {
			for j:=-1.0; j<=1; j++ {
				sim.findNNRec(sim.Root, k, Vec2{i, j})
			}
		}
	}
	// make dists use Sqrt
	// and set h value
	for i := range sim.Particles {
		for j := range NN_SIZE {
			sim.NNDists[i][j] = math.Sqrt(sim.NNDists[i][j])
		}
		sim.Particles[i].h = sim.NNDists[i][0]
	}
}


// TODO: @Speed fix Sqrts
func (sim *Simulation) findNNRec(root *Cell, particleIndex int, offset Vec2) {

	pos := sim.Particles[particleIndex].Pos.Add(&offset)
	particle := &sim.Particles[particleIndex]

	if root.Upper == nil && root.Lower == nil {
		for i := range root.Particles {
			d2 := DistSq(pos, root.Particles[i].Pos)

			// if the dist is lower than max dist and the particle is not itself!
			if d2 < sim.NNQueuePeekKey(particleIndex) && particle != &root.Particles[i] {
				sim.NNQueueInsert(particleIndex, d2, &root.Particles[i], root.Particles[i].Pos.Sub(&offset))
			}
		}
		return
	}

	if root.Upper != nil && root.Lower != nil {
		distUpper := Dist(root.Upper.BCenter, pos)
		distLower := Dist(root.Lower.BCenter, pos)

		// sqrt call ...
		maxDist := math.Sqrt(sim.NNQueuePeekKey(particleIndex))

		if distLower < distUpper {
			if distLower - root.Lower.BRadius < maxDist {
				sim.findNNRec(root.Lower, particleIndex, offset)
			}
			if distUpper - root.Upper.BRadius < maxDist {
				sim.findNNRec(root.Upper, particleIndex, offset)
			}
		} else {
			if distUpper - root.Upper.BRadius < maxDist {
				sim.findNNRec(root.Upper, particleIndex, offset)
			}
			if distLower - root.Lower.BRadius < maxDist {
				sim.findNNRec(root.Lower, particleIndex, offset)
			}

		}
		return
	}

	if root.Upper != nil {
		sim.findNNRec(root.Upper, particleIndex, offset)
	}

	if root.Lower != nil {
		sim.findNNRec(root.Lower, particleIndex, offset)
	}
}



// not used/ not sure if it is correct?
// dist squared to cell
func (cell *Cell) DistSquared(to *Vec2) float64 {
	d1 := to.Sub(&cell.UpperRight)
	d2 := cell.LowerLeft.Sub(to)
	maxx := math.Max(d1.X, d2.X)
	maxy := math.Max(d1.Y, d2.Y)
	return maxx*maxx + maxy*maxy
}


func (sim *Simulation) NNQueuePeekKey(particleIndex int) float64 {
	return sim.NNDists[particleIndex][0]
}

func (sim *Simulation) NNQueueInsert(particleIndex int, dist float64, neighbour *Particle, realPos Vec2) {
	sim.NNDists[particleIndex][0] = dist
	sim.NearestNeighbours[particleIndex][0] = neighbour
	sim.NNPos[particleIndex][0] = realPos
	sim.NNQueueHeapify(particleIndex, 0)
}

// TODO(#6): @Speed use memcopy
func (sim *Simulation) NNQueueInitSentinel() {
	for particleIndex := range sim.Particles{
		for j := range NN_SIZE {
			sim.NNDists[particleIndex][j] = math.MaxFloat64
			sim.NearestNeighbours[particleIndex][j] = nil
			sim.NNPos[particleIndex][j] = Vec2{}
		}
	}
}


func (sim *Simulation) NNQueueHeapify(pi int, i int) {
	for {
		l := i*2 + 1
		r := i*2 + 2
		max_index := i

		if l < NN_SIZE && sim.NNDists[pi][max_index] < sim.NNDists[pi][l]{
			max_index = l
		}

		if r < NN_SIZE && sim.NNDists[pi][max_index] < sim.NNDists[pi][r] {
			max_index = r
		}

		if max_index == i {
			break
		}

		// swap elements
		sim.NNDists[pi][i], sim.NNDists[pi][max_index] = sim.NNDists[pi][max_index], sim.NNDists[pi][i]
		sim.NearestNeighbours[pi][i], sim.NearestNeighbours[pi][max_index] = sim.NearestNeighbours[pi][max_index], sim.NearestNeighbours[pi][i]
		sim.NNPos[pi][i], sim.NNPos[pi][max_index] = sim.NNPos[pi][max_index], sim.NNPos[pi][i]

		i = max_index
	}
}
