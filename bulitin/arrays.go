package bulitin

import (
	"reflect"
	"sort"
)

type Arrays struct{}

func (Arrays) ParseStringArray(i interface{}) []string {
	return convtos(i)
}
func (Arrays) ParseIntArray(i interface{}) []int {
	return convtoi(i)
}
func (Arrays) ParseBoolArray(i interface{}) []bool {
	return convtob(i)
}
func (Arrays) ParseFloatArray(i interface{}) []float64 {
	return convtof(i)
}
func (Arrays) Sort(in interface{}) {
	if reflect.ValueOf(in).Kind() == reflect.Slice {
		if reflect.TypeOf(in).Elem().Kind() == reflect.Interface {
			sort.Slice(in, func(i, j int) bool {
				switch iv := reflect.ValueOf(in).Index(i).Elem(); {
				case iv.Kind() == reflect.Int:
					if jv := reflect.ValueOf(in).Index(j); jv.Kind() == reflect.Int {
						return iv.Int() < jv.Int()
					} else {
						return true
					}
				case iv.Kind() == reflect.String:
					if jv := reflect.ValueOf(in).Index(j).Elem(); jv.Kind() == reflect.String {
						return iv.String() < jv.String()
					} else {
						return false
					}
				case iv.Kind() == reflect.Float64:
					if jv := reflect.ValueOf(in).Index(j).Elem(); jv.Kind() == reflect.Float64 {
						return iv.Float() < jv.Float()
					} else {
						return true
					}
				}
				return false
			})
		} else {
			sort.Slice(in, func(i, j int) bool {
				switch iv := reflect.ValueOf(in).Index(i); {
				case iv.Kind() == reflect.Int:
					if jv := reflect.ValueOf(in).Index(j); jv.Kind() == reflect.Int {
						return iv.Int() < jv.Int()
					} else {
						return true
					}
				case iv.Kind() == reflect.String:
					if jv := reflect.ValueOf(in).Index(j); jv.Kind() == reflect.String {
						return iv.String() < jv.String()
					} else {
						return false
					}
				case iv.Kind() == reflect.Float64:
					if jv := reflect.ValueOf(in).Index(j); jv.Kind() == reflect.Float64 {
						return iv.Float() < jv.Float()
					} else {
						return true
					}
				}
				return false
			})
		}
	}
}
func (Arrays) Reverse(in interface{}) {
	if reflect.ValueOf(in).Kind() == reflect.Slice {
		if reflect.TypeOf(in).Elem().Kind() == reflect.Interface {
			sort.Slice(in, func(i, j int) bool {
				switch iv := reflect.ValueOf(in).Index(i).Elem(); {
				case iv.Kind() == reflect.Int:
					if jv := reflect.ValueOf(in).Index(j).Elem(); jv.Kind() == reflect.Int {
						return iv.Int() > jv.Int()
					} else {
						return false
					}
				case iv.Kind() == reflect.String:
					if jv := reflect.ValueOf(in).Index(j).Elem(); jv.Kind() == reflect.String {
						return iv.String() > jv.String()
					} else {
						return true
					}
				case iv.Kind() == reflect.Float64:
					if jv := reflect.ValueOf(in).Index(j).Elem(); jv.Kind() == reflect.Float64 {
						return iv.Float() > jv.Float()
					} else {
						return false
					}
				default:
					return false
				}
			})
		} else {
			sort.Slice(in, func(i, j int) bool {
				switch iv := reflect.ValueOf(in).Index(i); {
				case iv.Kind() == reflect.Int:
					if jv := reflect.ValueOf(in).Index(j); jv.Kind() == reflect.Int {
						return iv.Int() > jv.Int()
					} else {
						return false
					}
				case iv.Kind() == reflect.String:
					if jv := reflect.ValueOf(in).Index(j); jv.Kind() == reflect.String {
						return iv.String() > jv.String()
					} else {
						return true
					}
				case iv.Kind() == reflect.Float64:
					if jv := reflect.ValueOf(in).Index(j); jv.Kind() == reflect.Float64 {
						return iv.Float() > jv.Float()
					} else {
						return false
					}
				default:
					return false
				}
			})
		}
	}
}
func convtos(i interface{}) []string {
	s := make([]string, 0)
	if input := reflect.ValueOf(i); input.Kind() == reflect.Slice {
		for i := 0; i < input.Len(); i++ {
			if input.Index(i).Kind() == reflect.Interface {
				if input.Index(i).Elem().Kind() == reflect.String {
					s = append(s, input.Index(i).Elem().String())
				}
			} else {
				if input.Index(i).Kind() == reflect.String {
					s = append(s, input.Index(i).String())
				}
			}
		}
	}
	return s
}
func convtoi(i interface{}) []int {
	s := make([]int, 0)
	if input := reflect.ValueOf(i); input.Kind() == reflect.Slice {
		for i := 0; i < input.Len(); i++ {
			if input.Index(i).Kind() == reflect.Interface {
				if input.Index(i).Elem().Kind() == reflect.Int {
					s = append(s, int(input.Index(i).Elem().Int()))
				}
			} else {
				if input.Index(i).Kind() == reflect.Int {
					s = append(s, int(input.Index(i).Int()))
				}
			}
		}
	}
	return s
}
func convtof(i interface{}) []float64 {
	s := make([]float64, 0)
	if input := reflect.ValueOf(i); input.Kind() == reflect.Slice {
		for i := 0; i < input.Len(); i++ {
			if input.Index(i).Kind() == reflect.Interface {
				if input.Index(i).Elem().Kind() == reflect.Float64 {
					s = append(s, input.Index(i).Elem().Float())
				}
			} else {
				if input.Index(i).Kind() == reflect.Float64 {
					s = append(s, input.Index(i).Float())
				}
			}
		}
	}
	return s
}
func convtob(i interface{}) []bool {
	s := make([]bool, 0)
	if input := reflect.ValueOf(i); input.Kind() == reflect.Slice {
		for i := 0; i < input.Len(); i++ {
			if input.Index(i).Kind() == reflect.Interface {
				if input.Index(i).Elem().Kind() == reflect.Bool {
					s = append(s, input.Index(i).Elem().Bool())
				}
			} else {
				if input.Index(i).Kind() == reflect.Bool {
					s = append(s, input.Index(i).Bool())
				}
			}
		}
	}
	return s
}
