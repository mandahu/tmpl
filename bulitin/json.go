package bulitin

import "encoding/json"

//JSON解析库
type JSON struct{}

func (JSON) Parse(str string) (interface{}, error) {
	var j interface{}
	err := json.Unmarshal([]byte(str), &j)
	return j, err
}
func (JSON) String(i interface{}) (s string, err error) {
	if b, err := json.Marshal(i); err != nil {
		return string(""), err
	} else {
		return string(b), nil
	}
}
