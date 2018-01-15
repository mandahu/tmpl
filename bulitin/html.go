package bulitin

import (
	"fmt"
	"html"
)

type HTML struct{}

func (HTML) Escape(format string, arg ...interface{}) string {
	return html.EscapeString(fmt.Sprintf(format, arg...))
}
func (HTML) Unescape(str string) string {
	return html.UnescapeString(str)
}
