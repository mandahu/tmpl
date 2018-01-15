package main

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/mandahu/tmpl"
)

func main() {
	var reader = func(str string) ([]byte, error) {
		return ioutil.ReadFile(str)
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		buf := bytes.NewBufferString("")
		t := template.NewTemplate("example/example.html", reader, buf)
		if err := t.Compile(); err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
		ctx := template.NewContext(r, w)
		if err := t.Exec(ctx); err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
		w.WriteHeader(200)
		w.Write(buf.Bytes())
	})
	http.ListenAndServe("127.0.0.1:80", nil)

}
