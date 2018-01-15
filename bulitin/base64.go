package bulitin

import "encoding/base64"

type Base64 struct{}

func (Base64) Encode(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}
func (Base64) Decode(str string) (string, error) {
	return decode(str)
}

func decode(str string) (string, error) {
	b, e := base64.StdEncoding.DecodeString(str)
	return string(b), e
}
