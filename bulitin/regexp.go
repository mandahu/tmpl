package bulitin

import "regexp"

type RegExp struct{}

func (RegExp) Match(pattern, str string) (bool, error) {
	return regexp.MatchString(pattern, str)
}
func (RegExp) FindAll(pattern, str string) ([]string, error) {
	return findAll(pattern, str)
}
func findAll(pattern, str string) ([]string, error) {
	r, e := regexp.Compile(pattern)
	if e != nil {
		return nil, e
	}
	s := r.FindAllString(str, -1)
	return s, nil
}
