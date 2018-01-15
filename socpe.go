package template

type scope struct {
	top     *scope
	global  map[string]*ident //后加的需求。。
	objects map[string]*ident
}

var Functions map[string]*ident

func newScope(s *scope) *scope {
	if s == nil || s.global == nil {
		return &scope{nil, make(map[string]*ident), make(map[string]*ident)}
	}
	return &scope{s, s.global, make(map[string]*ident)}
}
func newTopScope(c interface{}) *scope {
	scope := newScope(nil)
	scope.objects = clone()
	scope.objects["ctx"] = &ident{pos: 0, name: "ctx", canset: false, data: c}
	return scope
}
func newIncludeScope(c interface{}, global map[string]*ident) *scope {
	scope := newScope(nil)
	scope.objects = clone()
	scope.objects["ctx"] = &ident{pos: 0, name: "ctx", canset: false, data: c}
	scope.global = global
	return scope
}
func (s *scope) isExist(name string) (*ident, bool) {
	if s == nil {
		return nil, false
	}
	s1 := *s
	for s1.top != nil {
		if obj, ok := s1.objects[name]; ok {
			return obj, true
		}
		s1 = *s1.top
	}
	if obj, ok := s1.objects[name]; ok {
		return obj, true
	}
	return nil, false
}
func (s *scope) isExistGlobal(name string) (*ident, bool) {
	if obj, ok := s.global[name]; ok {
		return obj, true
	}
	return nil, false
}
func (s *scope) insert(ident ...*ident) {
	for _, v := range ident {
		s.objects[v.name] = v
	}
}
func (s *scope) insertGlobal(ident ...*ident) {
	for _, v := range ident {
		s.global[v.name] = v
	}
}
func clone() map[string]*ident {
	c := make(map[string]*ident)
	for k, v := range Functions {
		c[k] = v
	}
	return c
}
func init() {
	fns := BuildFunc()
	Functions = make(map[string]*ident)
	for k, v := range fns {
		Functions[k] = &ident{pos: 0, name: k, canset: false, data: v}
	}
}
