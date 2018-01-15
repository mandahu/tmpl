package bulitin

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
)

type Crypto struct{}

func (Crypto) MD5(str string) string {
	return _md5(str)
}
func (Crypto) SHA1(str string) string {
	return _sha1(str)
}
func (Crypto) SHA256(str string) string {
	return _sha256(str)
}
func _md5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
func _sha1(str string) string {
	h := sha1.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
func _sha256(str string) string {
	h := sha256.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
