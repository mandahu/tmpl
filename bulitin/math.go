package bulitin

import (
	"math"
	"math/rand"
)

type Math struct{}

func (Math) Max(i, j float64) float64 {
	return math.Max(i, j)
}
func (Math) Min(i, j float64) float64 {
	return math.Min(i, j)
}
func (Math) Ceil(i float64) float64 {
	return math.Ceil(i)
}
func (Math) Floor(i float64) float64 {
	return math.Floor(i)
}
func (Math) Random() float64 {
	return rand.Float64()
}
