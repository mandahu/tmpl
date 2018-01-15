package bulitin

import (
	"strconv"
)

type Strconv struct{}

func (Strconv) ParseInt(s string) (int, error) {
	return strconv.Atoi(s)
}
func (Strconv) ParseFloat(str string) (float64, error) {
	return strconv.ParseFloat(str, 64)
}
