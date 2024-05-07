/*
	Tests for Partition() function
*/

package sim

import (
	"testing"
)

var psEven = [...]Particle{
	{Pos: Vec2{1.0, 1.0}},
	{Pos: Vec2{0.9, 0.9}},
	{Pos: Vec2{0.8, 0.8}},
	{Pos: Vec2{0.7, 0.7}},
}

func TestPartitionEvenVert1(t *testing.T) {

	a, b := Partition(psEven[:], Vertical, 0.5)

	if len(a) != 0 {
		t.Fatalf("expected len(a)=0, got %d", len(a))
	}
	if len(b) != 4 {
		t.Fatalf("expected len(b)=4, got %d", len(a))
	}
}

func TestPartitionEvenHor1(t *testing.T) {

	a, b := Partition(psEven[:], Horizontal, 0.5)

	if len(a) != 0 {
		t.Fatalf("expected len(a)=0, got %d", len(a))
	}
	if len(b) != 4 {
		t.Fatalf("expected len(b)=4, got %d", len(a))
	}
}

func TestPartitionEvenVert2(t *testing.T) {

	a, b := Partition(psEven[:], Vertical, 0.85)

	if len(a) != 2 {
		t.Fatalf("expected len(a)=2, got %d", len(a))
	}
	if len(b) != 2 {
		t.Fatalf("expected len(b)=2, got %d", len(a))
	}
}

func TestPartitionEvenHor2(t *testing.T) {

	a, b := Partition(psEven[:], Horizontal, 0.85)

	if len(a) != 2 {
		t.Fatalf("expected len(a)=2, got %d", len(a))
	}
	if len(b) != 2 {
		t.Fatalf("expected len(b)=2, got %d", len(a))
	}
}

var psOdd = [...]Particle{
	{Pos: Vec2{0.0, 0.9}},
	{Pos: Vec2{0.5, -0.8}},
	{Pos: Vec2{1.7, 0.1}},
	{Pos: Vec2{0.7, -0.1}},
	{Pos: Vec2{-0.7, 0.1}},
}

func TestPartitionOddVer1(t *testing.T) {

	a, b := Partition(psOdd[:], Vertical, 0.100000000001)

	if len(a) != 4 {
		t.Fatalf("expected len(a)=4, got %d", len(a))
	}
	if len(b) != 1 {
		t.Fatalf("expected len(b)=1, got %d", len(a))
	}
}

func TestPartitionOddHor1(t *testing.T) {

	a, b := Partition(psOdd[:], Horizontal, 0.601)

	if len(a) != 3 {
		t.Fatalf("expected len(a)=3, got %d", len(a))
	}
	if len(b) != 2 {
		t.Fatalf("expected len(b)=2, got %d", len(a))
	}
}

func TestPartitionOddVer2(t *testing.T) {

	a, b := Partition(psOdd[:], Vertical, -100)

	if len(a) != 0 {
		t.Fatalf("expected len(a)=0, got %d", len(a))
	}
	if len(b) != 5 {
		t.Fatalf("expected len(b)=5, got %d", len(a))
	}
}

func TestPartitionOddHor2(t *testing.T) {

	a, b := Partition(psOdd[:], Horizontal, 100)

	if len(a) != 5 {
		t.Fatalf("expected len(a)=5, got %d", len(a))
	}
	if len(b) != 0 {
		t.Fatalf("expected len(b)=0, got %d", len(a))
	}
}

func expect(lenA, lenB, a, b int, t *testing.T) {
	if lenA != a || lenB != b {
		t.Fatalf("expected len(a)=%d got %d and len(b)=%d got %d", lenA, a, lenB, b)
	}
}

var psVariation = [...]Particle{
	{Pos: Vec2{0.9, 0.0}},
	{Pos: Vec2{-0.8, 0.5}},
	{Pos: Vec2{0.1, 1.7}},
	{Pos: Vec2{-0.1, 0.7}},
	{Pos: Vec2{0.1, -0.7}},
}

func TestPartitionVariationHor1(t *testing.T) {
	a, b := Partition(psVariation[:], Horizontal, 0.100000000001)
	expect(4, 1, len(a), len(b), t)
}

func TestPartitionVariationVer1(t *testing.T) {
	a, b := Partition(psVariation[:], Vertical, 0.601)
	expect(3, 2, len(a), len(b), t)
}

func TestPartitionVariationHor2(t *testing.T) {
	a, b := Partition(psVariation[:], Horizontal, -100)
	expect(0, 5, len(a), len(b), t)
}

func TestPartitionVariationVer2(t *testing.T) {
	a, b := Partition(psVariation[:], Vertical, 100)
	expect(5, 0, len(a), len(b), t)
}

func TestPartitionEmpty(t *testing.T) {
	psEmpty := [...]Particle{}
	a, b := Partition(psEmpty[:], Horizontal, 0.85)
	expect(0, 0, len(a), len(b), t)
}
