package template

import (
	"fmt"
	"reflect"
	"template/bulitin"
)

func length(i interface{}) int {
	v := reflect.Indirect(reflect.ValueOf(i))
	switch v.Kind() {
	case reflect.String, reflect.Map, reflect.Slice:
		return v.Len()
	default:
		return 0
	}
}
func typeof(i interface{}) string {
	if i != nil {
		switch reflect.ValueOf(i).Kind() {
		case reflect.Map, reflect.String, reflect.Int, reflect.Float64:
			return reflect.ValueOf(i).Kind().String()
		case reflect.Slice:
			return reflect.TypeOf(i).String()
		}
	}
	return "null"
}
func itof(i int) float64 {
	return float64(i)
}
func ftoi(i float64) int {
	return int(i)
}
func BuildFunc() map[string]interface{} {
	funcs := make(map[string]interface{})
	funcs["printf"] = fmt.Sprintf
	funcs["len"] = length
	funcs["type"] = typeof
	funcs["Int"] = ftoi
	funcs["Float"] = itof
	funcs["strings"] = bulitin.String{}
	funcs["strconv"] = bulitin.Strconv{}
	funcs["arrays"] = bulitin.Arrays{}
	funcs["math"] = bulitin.Math{}
	funcs["time"] = bulitin.Time{}
	funcs["regexp"] = bulitin.RegExp{}
	funcs["url"] = bulitin.URL{}
	funcs["json"] = bulitin.JSON{}
	funcs["http"] = bulitin.Http{}
	funcs["crypto"] = bulitin.Crypto{}
	funcs["html"] = bulitin.HTML{}
	return funcs
}
