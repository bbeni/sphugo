package sim

import (
	"math"
)


// recursively find all nearest neighbors of a particle based on position
// make sure you call it on the top/root Cell!!
//
// TODO: implement it using a loop
func (particle *Particle) FindNearestNeighbours(root *Cell) {
	particle.NNQueueInitSentinel()
	particle.findNNRec(root, Vec2{0, 0})

	// make dists use Sqrt
	for i := range NN_SIZE {
		particle.NNDists[i] = math.Sqrt(particle.NNDists[i])
	}
}

// Periodic version
// assuming particles are between x = (0, 1], y = (0, 1]
func (particle *Particle) FindNearestNeighboursPeriodic(root *Cell) {
	particle.NNQueueInitSentinel()
	for i:=-1.0; i<=1; i++ {
		for j:=-1.0; j<=1; j++ {
			particle.findNNRec(root, Vec2{i, j})
		}
	}
	// make dists use Sqrt
	for i := range NN_SIZE {
		particle.NNDists[i] = math.Sqrt(particle.NNDists[i])
	}
}


// TODO: @Speed fix Sqrts
func (particle *Particle) findNNRec(root *Cell, offset Vec2) {

	pos := particle.Pos.Add(&offset)
	if root.Upper == nil && root.Lower == nil {
		for i := range root.Particles {
			d2 := DistSq(pos, root.Particles[i].Pos)

			// if the dist is lower than max dist and the particle is not itself!
			if d2 < particle.NNQueuePeekKey() && particle != &root.Particles[i] {
				particle.NNQueueInsert(d2, &root.Particles[i], root.Particles[i].Pos.Sub(&offset))
			}
		}
		return
	}

	if root.Upper != nil && root.Lower != nil {
		distUpper := Dist(root.Upper.BCenter, pos)
		distLower := Dist(root.Lower.BCenter, pos)

		// sqrt call ...
		maxDist := math.Sqrt(particle.NNQueuePeekKey())

		if distLower < distUpper {
			if distLower - root.Lower.BRadius < maxDist {
				particle.findNNRec(root.Lower, offset)
			}
			if distUpper - root.Upper.BRadius < maxDist {
				particle.findNNRec(root.Upper, offset)
			}
		} else {
			if distUpper - root.Upper.BRadius < maxDist {
				particle.findNNRec(root.Upper, offset)
			}
			if distLower - root.Lower.BRadius < maxDist {
				particle.findNNRec(root.Lower, offset)
			}

		}
		return
	}

	if root.Upper != nil {
		particle.findNNRec(root.Upper, offset)
	}

	if root.Lower != nil {
		particle.findNNRec(root.Lower, offset)
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

func (p *Particle) NNQueuePeekKey() float64 {
	return p.NNDists[0]
}


// This is actually faster than the heapque. it's probably beacuse we only have 32 NN's
// TODO: compare for other NN_SIZEs than 32
func (p *Particle) NNQueueInsert(dist float64, neighbour *Particle, realPos Vec2) {
	var i uint8 = 1
	for ;i<NN_SIZE && p.NNDists[i] > dist; i++ {
		p.NNDists[i-1] = p.NNDists[i]
		p.NearestNeighbours[i-1] = p.NearestNeighbours[i]
		p.NNPos[i-1] = p.NNPos[i]
	}
	p.NNDists[i-1] = dist
	p.NearestNeighbours[i-1] = neighbour
	p.NNPos[i-1] = realPos
}

func (p *Particle) NNQueueInitSentinel() {
	for i := range NN_SIZE {
		// This assumes the particles don't move more than
		// this value can be optimized, but actually might introduce errors in low density regions
		// @Inclomple be careful when changing coordinate system
		// TODO: make sure it's not too bad
		//
		//p.NNDists[i] = math.MaxFloat64
		p.NNDists[i] += 0.4
	}
}