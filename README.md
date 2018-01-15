# go 模板引擎
# init
go get github.com/mandahu/tmpl
```
  func main() {
	var reader = func(str string) ([]byte, error) {
		return ioutil.ReadFile(str)
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ctx := template.NewContext(r, w)
		buf := bytes.NewBufferString("")
		t := template.NewTemplate("example.html", reader, ctx, buf)
		if err := t.Compile(); err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
		if err := t.Exec(); err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
		w.WriteHeader(200)
		w.Write(buf.Bytes())
	})
	http.ListenAndServe("127.0.0.1:80", nil)

}
```

