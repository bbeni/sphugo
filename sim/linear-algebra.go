/* Linear Algebra functions and types
*/
package sim


import (
	"math"
)

// Math functions for int
func Abs(x int) int {
	if x < 0 { return -x }
	return x
}

func Max(x, y int) int {
	if x >= y {return x}
	return y
}

func Min(x, y int) int {
	if x <= y {return x}
	return y
}

// Linear algebra
type Vec2i struct {
	X, Y int
}

type Vec2 struct {
	X, Y float64
}

func (v *Vec2) Add(other *Vec2) Vec2 {
	return Vec2{v.X + other.X, v.Y + other.Y}
}

func (v *Vec2) Sub(other *Vec2) Vec2 {
	return Vec2{v.X - other.X, v.Y - other.Y}
}

func (v *Vec2) Dot(other *Vec2) float64 {
	return v.X * other.X + v.Y * other.Y
}

func (v Vec2) Mul(f float64) Vec2 {
	return Vec2{v.X*f, v.Y*f}
}


func (v *Vec2) Norm() float64 {
	return math.Sqrt(v.Dot(v))
}

func (v *Vec2) Normed() Vec2 {
	return v.Mul(1 / v.Norm())
}


func DistSq(a, b Vec2) float64 {
	dx, dy := a.X - b.X, a.Y - b.Y
	return dx*dx + dy*dy
}

func Dist(a, b Vec2) float64 {
	dx, dy := a.X - b.X, a.Y - b.Y
	return math.Hypot(dx, dy)
}
