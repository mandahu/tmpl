package bulitin

import (
	"strings"
)

type String struct{}

func (String) Join(s []string, sep string) string {
	return strings.Join(s, sep)
}
func (String) Split(s, sep string) []string {
	return strings.Split(s, sep)
}
func (String) Index(s, sep string) int {
	return strings.Index(s, sep)
}
func (String) LastIndex(s, sep string) int {
	return strings.LastIndex(s, sep)
}
func (String) ToUpper(s string) string {
	return strings.ToUpper(s)
}
func (String) ToLower(s string) string {
	return strings.ToLower(s)
}
func (String) Count(s, sub string) int {
	return strings.Count(s, sub)
}
