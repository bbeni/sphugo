package tg

import (
	"testing"
)

func isInsideAny(pos Vec2, cell *Cell) bool {

	x := pos.Sub(&cell.BCenter)

	if x.Dot(&x) <= cell.BRadiusSq {
		return true
	}

	if cell.Upper != nil {
		if isInsideAny(pos, cell.Upper) {
			return true
		}
	}

	if cell.Lower != nil {
		if isInsideAny(pos, cell.Lower) {
			return true
		}
	}

	return false
}

func TestInsideAnySphere60(t *testing.T) {

	cell := MakeCellsUniform(60, Vertical)
	cell.BoundingSpheres()

	for i, p := range cell.Particles {
		if !isInsideAny(p.Pos, &cell) {
			t.Fatalf("Particle %v `%v` is not inside any Cell!", i, p.Pos)
		}
	}
}

func TestInsideAnySphere1(t *testing.T) {

	cell := MakeCellsUniform(1, Vertical)
	cell.BoundingSpheres()

	for i, p := range cell.Particles {
		if !isInsideAny(p.Pos, &cell) {
			t.Fatalf("Particle %v `%v` is not inside any Cell!", i, p.Pos)
		}
	}
}

func TestInsideAnySphere6000(t *testing.T) {

	cell := MakeCellsUniform(6000, Vertical)
	cell.BoundingSpheres()

	for i, p := range cell.Particles {
		if !isInsideAny(p.Pos, &cell) {
			t.Fatalf("Particle %v `%v` is not inside any Cell!", i, p.Pos)
		}
	}
}