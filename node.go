package template

import (
	"bytes"
	"fmt"
)

type node interface {
	begin() int //start pos
	end() int   // end pos
	//Line() int
	//Col() int
}
type expr interface {
	node
	exprNode()
}
type stmt interface {
	node
	stmtNode()
}

//custom expression collection
type (
	badExpr struct {
		s, e, l, c int
		text       string
	}
	literal struct {
		pos int
		typ token
		val string
	}
	ident struct {
		pos    int
		name   string
		canset bool
		data   interface{}
	}
	slice struct {
		typ     token
		lbpos   int
		element []expr
		rbpos   int
	}
	property struct {
		key   expr
		colon int
		value expr
	}
	object struct {
		lbpos      int
		properties []property
		rbpos      int
	}
	unaryExpr struct {
		pos int
		op  token
		e   expr
	}
	binaryExpr struct {
		left  expr
		opPos int
		op    token
		right expr
	}
	memberExpr struct {
		x   expr
		sel *ident
	}
	callExpr struct {
		x     expr
		name  string
		lppos int
		args  []expr
		rppos int
	}
	memBerCallExpr struct {
		x     expr
		name  string
		lppos int
		args  []expr
		rppos int
	}
	indexExpr struct {
		x      expr
		lbpos  int
		index  expr
		colon  token
		index1 expr
		rbpos  int
	}
	globalIdentExpr struct {
		pos    int
		period int
		ident  *ident
	}
	globalCallExpr struct {
		period int
		name   string
		lppos  int
		args   []expr
		rppos  int
	}
	htmlExpr struct {
		pos int
		val string
	}
)

func (bad *badExpr) begin() int {
	return bad.s
}
func (bad *badExpr) end() int {
	return bad.e
}
func (bad *badExpr) line() int {
	return bad.l
}
func (bad *badExpr) col() int {
	return bad.c
}
func (lit *literal) begin() int {
	return lit.pos
}
func (lit *literal) end() int {
	return lit.pos + len(lit.val)
}
func (ident *ident) begin() int {
	return ident.pos
}
func (ident *ident) end() int {
	return ident.pos + len(ident.name)
}
func (s *slice) begin() int {
	return s.lbpos
}
func (s *slice) end() int {
	return s.rbpos
}
func (p *property) begin() int {
	return p.key.begin()
}
func (p *property) end() int {
	return p.value.end()
}
func (o *object) begin() int {
	return o.lbpos
}
func (o *object) end() int {
	return o.rbpos
}
func (u *unaryExpr) begin() int {
	return u.pos
}
func (u *unaryExpr) end() int {
	return u.e.end()
}
func (b *binaryExpr) begin() int {
	return b.left.begin()
}
func (b *binaryExpr) end() int {
	return b.right.end()
}
func (m *memberExpr) begin() int {
	return m.x.begin()
}
func (m *memberExpr) end() int {
	return m.x.end()
}
func (c *callExpr) begin() int {
	return c.x.begin()
}
func (c *callExpr) end() int {
	return c.x.end()
}
func (mc *memBerCallExpr) begin() int {
	return mc.x.begin()
}
func (mc *memBerCallExpr) end() int {
	return mc.rppos
}
func (i *indexExpr) begin() int {
	return i.lbpos
}
func (i *indexExpr) end() int {
	return i.rbpos
}
func (g *globalIdentExpr) begin() int {
	return g.ident.begin()
}
func (g *globalIdentExpr) end() int {
	return g.ident.end()
}
func (g *globalCallExpr) begin() int {
	return g.period
}
func (g *globalCallExpr) end() int {
	return g.rppos
}
func (h *htmlExpr) begin() int {
	return h.pos
}
func (h *htmlExpr) end() int {
	return h.pos + len(h.val)
}
func (bad *badExpr) exprNode()       {}
func (lit *literal) exprNode()       {}
func (ident *ident) exprNode()       {}
func (s *slice) exprNode()           {}
func (p *property) exprNode()        {}
func (o *object) exprNode()          {}
func (u *unaryExpr) exprNode()       {}
func (b *binaryExpr) exprNode()      {}
func (m *memberExpr) exprNode()      {}
func (m *callExpr) exprNode()        {}
func (i *indexExpr) exprNode()       {}
func (g *globalIdentExpr) exprNode() {}
func (g *globalCallExpr) exprNode()  {}
func (m *memBerCallExpr) exprNode()  {}
func (h *htmlExpr) exprNode()        {}
func (this *badExpr) String() string {
	buff := bytes.NewBuffer(make([]byte, 0))
	buff.WriteString(fmt.Sprintf("BadExpression:%s", this.text))
	return buff.String()
}
func (this *literal) String() string {
	buff := bytes.NewBuffer(make([]byte, 0))
	buff.WriteString(fmt.Sprintf("Literal:%+v", *this))
	return buff.String()
}
func (i *ident) String() string {
	buff := bytes.NewBuffer(make([]byte, 0))
	buff.WriteString(fmt.Sprintf("Ident:%+v", *i))
	return buff.String()
}
func (this *slice) String() string {
	buff := bytes.NewBuffer(make([]byte, 0))
	buff.WriteString(fmt.Sprintf("SliceExpr:%+v", *this))
	return buff.String()
}
func (this *property) String() string {
	buff := bytes.NewBuffer(make([]byte, 0))
	buff.WriteString(fmt.Sprintf("Property:%+v", *this))
	return buff.String()
}
func (this *object) String() string {
	buff := bytes.NewBuffer(make([]byte, 0))
	buff.WriteString(fmt.Sprintf("ObjectExpr:%+v", *this))
	return buff.String()
}
func (this *unaryExpr) String() string {
	buff := bytes.NewBuffer(make([]byte, 0))
	buff.WriteString(fmt.Sprintf("UnaryExpr:%+v", *this))
	return buff.String()
}
func (this *binaryExpr) String() string {
	buff := bytes.NewBuffer(make([]byte, 0))
	buff.WriteString(fmt.Sprintf("BinaryExpr:%+v", *this))
	return buff.String()
}
func (this *memberExpr) String() string {
	buff := bytes.NewBuffer(make([]byte, 0))
	buff.WriteString(fmt.Sprintf("parseMemberExpr:%+v", *this))
	return buff.String()
}
func (this *callExpr) String() string {
	buff := bytes.NewBuffer(make([]byte, 0))
	buff.WriteString(fmt.Sprintf("callExpr:%+v", *this))
	return buff.String()
}
func (this *indexExpr) String() string {
	buff := bytes.NewBuffer(make([]byte, 0))
	buff.WriteString(fmt.Sprintf("IndexExpr:%+v", *this))
	return buff.String()
}
func (this *globalIdentExpr) String() string {
	buff := bytes.NewBuffer(make([]byte, 0))
	buff.WriteString(fmt.Sprintf("GlobalIdentExpr:%+v", *this))
	return buff.String()
}
func (this *globalCallExpr) String() string {
	buff := bytes.NewBuffer(make([]byte, 0))
	buff.WriteString(fmt.Sprintf("GlobalCallExpr:%+v", *this))
	return buff.String()
}
func (this *memBerCallExpr) String() string {
	buff := bytes.NewBuffer(make([]byte, 0))
	buff.WriteString(fmt.Sprintf("MemBerCallExpr:%+v", *this))
	return buff.String()
}
func (this *htmlExpr) String() string {
	buff := bytes.NewBuffer(make([]byte, 0))
	buff.WriteString(fmt.Sprintf("HtmlExpr:%+v", *this))
	return buff.String()
}

//custom statement collection

type (
	badStmt struct {
		s, e, l, c int
		text       string
	}
	varDecStmt struct {
		keywords int //keywords=var
		ident    []*ident
		opPos    int
		x        []expr
	}
	blockStmt struct {
		left  int
		list  []node
		right int
	}
	funDecStmt struct {
		pos   int
		name  string
		lbpos int
		args  []*ident
		rbpos int
		body  *blockStmt
	}
	assignStmt struct {
		lhs   expr
		opPos int
		op    token
		rhs   expr
	}
	exprStmt struct {
		x expr
	}
	ifStmt struct {
		pos    int
		cond   expr
		body   *blockStmt
		branch stmt
	}
	forStmt struct {
		pos              int
		key              *ident
		keywords         int //keywords=range
		lbpos            int
		init, cond, step expr
		rbpos            int
		body             *blockStmt
	}
	forRange struct {
		pos      int
		key, val *ident
		keywords int //keywords=range
		target   expr
		body     *blockStmt
	}
	breakStmt struct {
		pos int
	}
	returnStmt struct {
		pos int
		x   []expr
	}
	echoStmt struct {
		//@(...expr)
		pos   int //@ is echo start symbol
		lbpos int
		exprs []expr
		rbpos int
	}
	include struct {
		name  string
		lbpos int
		tree  []node
		rbpos int
		args  []expr
	}
)

func (b *badStmt) begin() int {
	return b.s
}
func (b *badStmt) end() int {
	return b.e
}
func (b *badStmt) line() int {
	return b.l
}
func (b *badStmt) col() int {
	return b.c
}
func (v *varDecStmt) begin() int {
	return v.keywords
}
func (v *varDecStmt) end() int {
	return v.x[len(v.x)].end()
}
func (b *blockStmt) begin() int {
	return b.left
}
func (b *blockStmt) end() int {
	return b.end()
}
func (f *funDecStmt) begin() int {
	return f.pos
}
func (f *funDecStmt) end() int {
	return f.body.end()
}
func (a *assignStmt) begin() int {
	return a.lhs.begin()
}
func (a *assignStmt) end() int {
	return a.rhs.end()
}
func (a *exprStmt) begin() int {
	return a.x.begin()
}
func (a *exprStmt) end() int {
	return a.x.end()
}
func (i *ifStmt) begin() int {
	return i.pos
}
func (i *ifStmt) end() int {
	return i.cond.end()
}
func (f *forStmt) begin() int {
	return f.pos
}
func (f *forStmt) end() int {
	return f.rbpos
}
func (f *forRange) begin() int {
	return f.pos
}
func (f *forRange) end() int {
	return f.target.end()
}
func (b *breakStmt) begin() int {
	return b.pos
}
func (b *breakStmt) end() int {
	return b.pos + len(DT_BREAK.String())
}
func (r *returnStmt) begin() int {
	return r.pos
}
func (r *returnStmt) end() int {
	return r.x[len(r.x)].end()
}
func (e *echoStmt) begin() int {
	return e.pos
}
func (e *echoStmt) end() int {
	return e.rbpos
}
func (i *include) begin() int {
	return i.rbpos
}
func (i *include) end() int {
	return i.lbpos
}
func (b *badStmt) stmtNode()    {}
func (v *varDecStmt) stmtNode() {}
func (b *blockStmt) stmtNode()  {}
func (f *funDecStmt) stmtNode() {}
func (a *assignStmt) stmtNode() {}
func (a *exprStmt) stmtNode()   {}
func (i *ifStmt) stmtNode()     {}
func (f *forStmt) stmtNode()    {}
func (f *forRange) stmtNode()   {}
func (b *breakStmt) stmtNode()  {}
func (r *returnStmt) stmtNode() {}
func (e *echoStmt) stmtNode()   {}
func (i *include) stmtNode()    {}
func (this *badStmt) String() string {
	buff := bytes.NewBuffer(make([]byte, 0))
	buff.WriteString(fmt.Sprintf("BadStatement:%v", this.text))
	return buff.String()
}
func (this *varDecStmt) String() string {
	buff := bytes.NewBuffer(make([]byte, 0))
	buff.WriteString(fmt.Sprintf("varDecStmt:%+v", *this))
	return buff.String()
}
func (this *funDecStmt) String() string {
	buff := bytes.NewBuffer(make([]byte, 0))
	buff.WriteString(fmt.Sprintf("funDecStmt:%+v", *this))
	return buff.String()
}
func (this *assignStmt) String() string {
	buff := bytes.NewBuffer(make([]byte, 0))
	buff.WriteString(fmt.Sprintf("AssignStmt:%+v", *this))
	return buff.String()
}
func (this *exprStmt) String() string {
	buff := bytes.NewBuffer(make([]byte, 0))
	buff.WriteString(fmt.Sprintf("exprStmt:%+v", *this))
	return buff.String()
}
func (this *ifStmt) String() string {
	buff := bytes.NewBuffer(make([]byte, 0))
	buff.WriteString(fmt.Sprintf("IfStmt:%+v", *this))
	return buff.String()
}
func (this *forStmt) String() string {
	buff := bytes.NewBuffer(make([]byte, 0))
	buff.WriteString(fmt.Sprintf("ForStmt:%+v", *this))
	return buff.String()
}
func (this *forRange) String() string {
	buff := bytes.NewBuffer(make([]byte, 0))
	buff.WriteString(fmt.Sprintf("ForRange:%+v", *this))
	return buff.String()
}
func (this *breakStmt) String() string {
	buff := bytes.NewBuffer(make([]byte, 0))
	buff.WriteString(fmt.Sprintf("break"))
	return buff.String()
}
func (this *returnStmt) String() string {
	buff := bytes.NewBuffer(make([]byte, 0))
	buff.WriteString(fmt.Sprintf("Return%+v:", *this))
	return buff.String()
}
func (this *echoStmt) String() string {
	buff := bytes.NewBuffer(make([]byte, 0))
	buff.WriteString(fmt.Sprintf("Echo:%+v", *this))
	return buff.String()
}
func (this *blockStmt) String() string {
	buff := bytes.NewBuffer(make([]byte, 0))
	buff.WriteString(fmt.Sprintf("BlockStmt:%+v", *this))
	return buff.String()
}
func (this *include) String() string {
	buff := bytes.NewBuffer(make([]byte, 0))
	buff.WriteString(fmt.Sprintf("Include:%+v", *this))
	return buff.String()
}
