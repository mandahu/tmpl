package bulitin

import "net/url"

type URL struct{}

func (URL) EscapeComponent(str string) string {
	return url.QueryEscape(str)
}
func (URL) Escape(str string) (string, error) {
	u, e := url.Parse(str)
	return u.String(), e
}
func (URL) UnEscape(str string) (string, error) {
	return url.QueryUnescape(str)
}
