package template

import (
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Context struct {
	req      *http.Request
	res      http.ResponseWriter
	Protocol string //使用的http协议版本
	Host     string //客户端访问的域名
	Path     string //客户端访问的path
	Method   string
	ClientIp string
	Header   map[string]interface{}
	Cookie   map[string]interface{}
	Query    map[string]interface{} //客户端请求的url参数
	Body     map[string]interface{} //客户端请求的body参数
	RawBody  string                 //raw body
	Params   map[string]interface{} //:id 或者*path
}

func NewContext(req *http.Request, res http.ResponseWriter) *Context {
	c := &Context{}
	c.set(req, res)
	c.Protocol = req.Proto
	c.Host = req.Host
	c.Path = req.URL.Path
	c.Method = req.Method
	c.ClientIp = clientIp(c.req)
	c.Query = make(map[string]interface{})
	c.Body = make(map[string]interface{})
	c.Header = make(map[string]interface{})
	c.Cookie = make(map[string]interface{})
	c.Params = make(map[string]interface{})
	for k, v := range req.Header {
		p := make([]string, 0)
		for _, v1 := range v {
			p = append(p, v1)
		}
		c.Header[k] = p
	}
	for _, v := range req.Cookies() {
		c.Cookie[v.Name] = v.Value
	}
	for k, v := range req.URL.Query() {
		c.Query[k] = v[0]
	}
	if err := req.ParseMultipartForm(2 << 10); err != nil {
		if len(req.PostForm) == 0 {
			//Raw
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				c.RawBody = ""
			} else {
				c.RawBody = string(body)
			}
		} else {
			//x-www-form-urlencoded
			for k, v := range req.PostForm {
				c.Body[k] = v[0]
			}
		}
	}
	return c
}
func (c *Context) set(req *http.Request, res http.ResponseWriter) {
	c.req = req
	c.res = res
}
func (c *Context) SetHeader(k, v string) {
	c.res.Header().Set(k, v)
}
func (c *Context) SetCookie(name, value string, expires int, secure, httpOnly bool) {
	path := c.req.Header.Get("aw-original-uri")
	domain := c.req.Header.Get("aw-original-host")
	u, _ := url.QueryUnescape(path)
	cookie := http.Cookie{
		Name:     name,
		Value:    url.QueryEscape(value),
		Path:     u,
		Expires:  time.Now().Add(time.Duration(expires) * time.Second),
		Domain:   domain,
		Secure:   secure,   //为true时 除非使用ssl和https否则则无法保存，
		HttpOnly: httpOnly, //为true时 JavaScript和xhr以及feth.request不能访问
	}
	if ck := cookie.String(); ck != "" {
		c.res.Header().Add("Set-Cookie", ck)
	}
}
func (c *Context) Redirect(code int, location string) string {
	http.Redirect(c.res, c.req, location, code)
	return ""
}
func clientIp(req *http.Request) string {
	clientIP := strings.TrimSpace(req.Header.Get("leonid-real-ip"))
	if len(clientIP) > 0 {
		return clientIP
	}
	clientIP = strings.TrimSpace(req.Header.Get("ali-cdn-real-ip"))
	if len(clientIP) > 0 {
		return clientIP
	}
	clientIP = req.Header.Get("X-Forwarded-For")
	if index := strings.IndexByte(clientIP, ','); index >= 0 {
		clientIP = clientIP[0:index]
	}
	clientIP = strings.TrimSpace(clientIP)
	if len(clientIP) > 0 {
		return clientIP
	}
	if ip, _, err := net.SplitHostPort(strings.TrimSpace(req.RemoteAddr)); err == nil {
		return ip
	}
	return ""
}
func (c *Context) Status(code int) {
	c.res.WriteHeader(code)
}
