package template

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

type inputReader func(path string) ([]byte, error)
type relation struct {
	Name  string      `json:"name"`
	Child []*relation `json:"child"`
}

func (r *relation) insertRelation(relation *relation) {
	r.Child = append(r.Child, relation)
}
func (p *relation) String() string {
	if byte, err := json.Marshal(p); err != nil {
		return "解析关系出错"
	} else {
		return string(byte)
	}
}

type parser struct {
	name        string
	input       []byte
	left, right string
	err         error
	lex         *lex
	tree        []node
	cur         item
	pre         item
	readFn      inputReader
	deep        int
	relation    *relation
}

func newParser(name string, left, right string, read inputReader) (*parser, error) {
	p := &parser{
		name:     name,
		readFn:   read,
		left:     left,
		right:    right,
		relation: &relation{Name: name, Child: make([]*relation, 0)},
	}
	input, err := p.readFn(name)
	if err != nil {
		return nil, fmt.Errorf("加载=>%s模板出错:%s", name, err.Error())
	}
	p.input = input
	p.lex = lexer(input, left, right)
	return p, nil
}
func newInclude(name string, left, right string, reader inputReader, r *relation) (*parser, error) {
	p := &parser{
		name:     name,
		readFn:   reader,
		left:     left,
		right:    right,
		relation: r,
	}
	input, err := p.readFn(name)
	if err != nil {
		return nil, fmt.Errorf("加载=>%s模板出错:%s", name, err.Error())
	}
	p.input = input
	p.lex = lexer(input, left, right)
	return p, nil
}
func (p *parser) error(err string) {
	col := p.cur.p - bytes.LastIndex(p.input[:p.cur.p], []byte("\n"))
	msg := fmt.Sprintf("文件名称:%s\n%s\n行:%d,列:%d", p.name, err, p.cur.l, col)
	p.err = fmt.Errorf(msg)
	panic(msg)
}
func (p *parser) error0(err string) {
	col := p.pre.p - bytes.LastIndex(p.input[:p.pre.p], []byte("\n"))
	msg := fmt.Sprintf("文件名称:%s\n%s\n行:%d,列:%d", p.name, err, p.pre.l, col)
	p.err = fmt.Errorf(msg)
	panic(msg)
}
func (p *parser) next() {
	if item, ok := <-p.lex.item; ok {
		p.pre = p.cur
		p.cur = item
	}
}
func (p *parser) expect(t token, err string) (i item) {
	i = p.cur
	if p.cur.t == t {
		p.next()
	} else {
		if p.cur.t == DT_ILLEGAL {
			p.error(p.cur.v)
		} else {
			p.error(err)
		}
	}
	return
}
func (p *parser) append(node ...node) {
	p.tree = append(p.tree, node...)
}
func (p *parser) start() ([]node, error) {
	func(err *error) {
		defer recoverHandle(err)
		p.next()
		p.parseNode()

	}(&p.err)
	return p.tree, p.err
}
func recoverHandle(err *error) {
	if e := recover(); e != nil {
		*err = fmt.Errorf("%s", e)
	}
}
func (p *parser) parseNode() {
	for p.cur.t != DT_EOF {
		switch p.cur.t {
		case DT_HTMLTEXT:
			p.append(p.parseHTML())
		case DT_LEFTDEL:
			p.append(p.parseScrpit()...)
		}
	}
}
func (p *parser) parseHTML() (h *htmlExpr) {
	h = &htmlExpr{
		pos: p.cur.p,
		val: p.cur.v,
	}
	p.next()
	return h
}
func (p *parser) parseScrpit() (n []node) {
	p.next()
	for p.cur.t != DT_EOF && p.cur.t != DT_RIGHTDEL {
		n = append(n, p.parseStmt(false, false))
	}
	p.expect(DT_RIGHTDEL, "缺少右侧脚本分隔符:"+p.right)
	return n
}
func (p *parser) parseStmt(isfun, isfor bool) (s stmt) {
	switch p.cur.t {
	case DT_VAR:
		s = p.parseDecVar()
	case DT_FUNC:
		if isfun {
			p.error("声明错误:自定义函数内无法再声明函数")
		}
		s = p.parseDecFunc()
	case DT_IDENT, DT_GLOBAL:
		s = p.parseAssignStmt()
	case DT_IF:
		s = p.parseIfStmt(isfun, isfor)
	case DT_FOR:
		s = p.parseForStmt(isfun)
	case DT_RETURN:
		s = p.parseReturnStmt()
	case DT_ECHO:
		s = p.parseEcho()
	case DT_INCLUDE:
		s = p.parseInclude(&p.deep, p.relation)
	case DT_BREAK:
		if !isfor {
			p.error("break应该被for语句包裹")
		}
		s = p.parseBreakStmt()
	case DT_ILLEGAL:
		p.error("词法错误:不支持的词法格式==>" + p.cur.v)
	default:
		p.error("语法错误:不支持的语法==>" + p.cur.v)
	}
	if p.cur.t == DT_SEMICOLON {
		p.next()
	}
	return
}
func (p *parser) parseExpr() (e expr) {
	e = p.parseLogicalORExpr()
	return
}
func (p *parser) parseLogicalORExpr() (e expr) {
	e = p.parseLogicalANDExpr()
	for p.cur.t == DT_LOGIC_OR {
		pos := p.cur.p
		p.next()
		rhs := p.parseLogicalANDExpr()
		e = &binaryExpr{left: e, opPos: pos, op: DT_LOGIC_OR, right: rhs}
	}
	return e
}
func (p *parser) parseLogicalANDExpr() (e expr) {
	e = p.parseEqualityExpr()
	for p.cur.t == DT_LOGIC_AND {
		pos := p.cur.p
		p.next()
		rhs := p.parseEqualityExpr()
		e = &binaryExpr{left: e, opPos: pos, op: DT_LOGIC_AND, right: rhs}
	}
	return e
}
func (p *parser) parseEqualityExpr() (e expr) {
	e = p.parseRelationalExpr()
	for p.cur.t == DT_NEQ || p.cur.t == DT_EQL {
		token := p.cur.t
		pos := p.cur.p
		p.next()
		rhs := p.parseRelationalExpr()
		e = &binaryExpr{left: e, opPos: pos, op: token, right: rhs}
	}
	return e
}
func (p *parser) parseRelationalExpr() (e expr) {
	e = p.parseAdditiveExpr()
	for p.cur.t == DT_LSS || p.cur.t == DT_LEQ || p.cur.t == DT_GTR || p.cur.t == DT_GEQ {
		token := p.cur.t
		pos := p.cur.p
		p.next()
		rhs := p.parseAdditiveExpr()
		e = &binaryExpr{left: e, opPos: pos, op: token, right: rhs}
	}
	return e
}
func (p *parser) parseAdditiveExpr() (e expr) {
	e = p.parseMultiplicativeExpr()
	for p.cur.t == DT_ADD || p.cur.t == DT_SUB {
		token := p.cur.t
		pos := p.cur.p
		p.next()
		rhs := p.parseMultiplicativeExpr()
		e = &binaryExpr{left: e, opPos: pos, op: token, right: rhs}
	}
	return e
}
func (p *parser) parseMultiplicativeExpr() (e expr) {
	e = p.parseUnaryExpr()
	for p.cur.t == DT_MUL || p.cur.t == DT_QUO || p.cur.t == DT_REM {
		token := p.cur.t
		pos := p.cur.p
		p.next()
		rhs := p.parseUnaryExpr()
		e = &binaryExpr{left: e, opPos: pos, op: token, right: rhs}
	}
	return e
}
func (p *parser) parseUnaryExpr() (e expr) {
	switch p.cur.t {
	case DT_NOT:
		token := p.cur.t
		pos := p.cur.p
		p.next()
		e = &unaryExpr{pos, token, p.parseUnaryExpr()}
	case DT_ADD:
		token := p.cur.t
		pos := p.cur.p
		p.next()
		e = &unaryExpr{pos, token, p.parseUnaryExpr()}
	case DT_SUB:
		token := p.cur.t
		pos := p.cur.p
		p.next()
		e = &unaryExpr{pos, token, p.parseUnaryExpr()}
	default:
		e = p.parsePostfixExpr()
	}
	return e
}
func (p *parser) parsePostfixExpr() (e expr) {
	e = p.parsePrimaryExpr()
	switch p.cur.t {
	case DT_INC, DT_DEC:
		switch e.(type) {
		case *ident, *memberExpr, *indexExpr:
		default:
			p.error(`语法错误:++,--操作的左手表达式无效`)
		}
		token := p.cur.t
		pos := p.cur.p
		e = &unaryExpr{pos: pos, op: token, e: e}
		p.next()
	}
	return e
}
func (p *parser) parsePrimaryExpr() (e expr) {
	switch p.cur.t {
	case DT_LBRACE: //{
		e = p.parseObjectLiteral()
		return
	case DT_LPAREN: //(
		p.next()
		e = p.parseExpr()
		p.expect(DT_RPAREN, "missing )")
		return
	case DT_LBRACK, DT_INT, DT_FLOAT, DT_BOOL, DT_STRING: //[
		e = p.parseArrayLiteral()
		return
	case DT_IDENT:
		e = p.parseIdent()
		return
	case DT_GLOBAL:
		e = p.parseGlobalExpr()
		return
	case DT_STRING_LIT, DT_RAWSTRING_LIT, DT_INT_LIT, DT_FLOAT_LIT, DT_BOOL_LIT, DT_NULL:
		e = p.parseLiteral()
	case DT_ILLEGAL:
		p.error(p.cur.v)
	default:
		p.error0("语法错误:请检查语句")
	}
	p.next()
	return e
}
func (p *parser) parseObjectLiteral() (e expr) {
	left := p.expect(DT_LBRACE, "{").p
	if p.cur.t == DT_RBRACE {
		right := p.cur.p
		p.next()
		return &object{lbpos: left, properties: nil, rbpos: right}
	}
	propertys := p.parsePropertys()
	right := p.expect(DT_RBRACE, `语法错误:缺少符号"}"`).p
	return &object{left, propertys, right}
}
func (p *parser) parsePropertys() (e []property) {
	e = append(e, p.parseProperty())
	for p.cur.t == DT_COMMA {
		p.next()
		e = append(e, p.parseProperty())
	}
	return e
}
func (p *parser) parseProperty() (e property) {
	key := p.parseKeyName()
	pos := p.expect(DT_COLON, `语法错误:缺少符号 ":"`).p
	value := p.parseExpr()
	return property{key: key, colon: pos, value: value}
}
func (p *parser) parseKeyName() (e expr) {
	switch p.cur.t {
	case DT_STRING_LIT:
		e = p.parseLiteral()
	case DT_IDENT:
		e = p.parseExpr()
	default:
		p.error("语法错误:map的键只能是字符串或者标识符")
	}
	p.next()
	return
}
func (p *parser) parseLiteral() (l *literal) {
	switch p.cur.t {
	case DT_STRING_LIT:
		str, e := strconv.Unquote(string(p.cur.v))
		if e != nil {
			p.error(fmt.Sprintf("非法字符串字面量"))
		}
		return &literal{pos: p.cur.p, typ: DT_STRING_LIT, val: str}
	case DT_RAWSTRING_LIT:
		raw, e := strconv.Unquote(p.cur.v)
		if e != nil {
			p.error(fmt.Sprintf("非法的rawString字面量"))
		}
		return &literal{pos: p.cur.p, typ: DT_RAWSTRING_LIT, val: raw}
	case DT_INT_LIT:
		return &literal{pos: p.cur.p, typ: DT_INT_LIT, val: p.cur.v}
	case DT_FLOAT_LIT:
		return &literal{pos: p.cur.p, typ: DT_FLOAT_LIT, val: p.cur.v}
	case DT_BOOL_LIT:
		return &literal{pos: p.cur.p, typ: DT_BOOL_LIT, val: p.cur.v}
	case DT_NULL:
		return &literal{pos: p.cur.p, typ: DT_NULL, val: "null"}
	}
	return l
}
func (p *parser) parseArrayLiteral() (e expr) {
	switch p.cur.t {
	case DT_LBRACK:
		left := p.expect(DT_LBRACK, "")
		if p.cur.t == DT_RBRACK {
			right := p.cur.p
			p.next()
			return &slice{typ: -1, lbpos: left.p, element: nil, rbpos: right}
		} else {
			eles := make([]expr, 0)
			eles = append(eles, p.parseExpr())
			for p.cur.t == DT_COMMA {
				p.next()
				eles = append(eles, p.parseExpr())
			}
			right := p.expect(DT_RBRACK, `语法错误:缺少符号 "]"`).p
			return &slice{lbpos: left.p, element: eles, rbpos: right}
		}
	default:
		typ := p.cur.t
		p.next()
		left := p.expect(DT_LBRACK, `语法错误缺少:缺少符号"["`)
		if p.cur.t == DT_RBRACK {
			right := p.cur.p
			p.next()
			return &slice{typ: typ, lbpos: left.p, element: nil, rbpos: right}
		} else {
			eles := make([]expr, 0)
			eles = append(eles, p.parseExpr())
			for p.cur.t == DT_COMMA {
				p.next()
				eles = append(eles, p.parseExpr())
			}
			right := p.expect(DT_RBRACK, `语法错误:缺少符号 "]"`).p
			return &slice{typ: typ, lbpos: left.p, element: eles, rbpos: right}
		}
	}
}
func (p *parser) parseIdent() (e expr) {
	e = &ident{pos: p.cur.p, name: p.cur.v, data: nil}
	id := p.expect(DT_IDENT, "语法错误:缺少标识符")
	for p.cur.t == DT_PERIOD || p.cur.t == DT_LBRACK || p.cur.t == DT_LPAREN {
		switch p.cur.t {
		case DT_PERIOD:
			p.next()
			selId := p.expect(DT_IDENT, "语法错误:缺少标识符")
			sel := &ident{pos: selId.p, name: selId.v, data: nil}
			if p.cur.t == DT_LPAREN {
				p.next()
				mc := &memBerCallExpr{x: e, name: sel.name, lppos: p.cur.p}
				if p.cur.t == DT_RPAREN {
					mc.args = nil
				} else {
					mc.args = p.parseValues()
				}
				mc.rppos = p.expect(DT_RPAREN, `语法错误:缺少符号 ")"`).p
				e = mc
			} else {
				e = &memberExpr{sel: sel, x: e}
			}
		case DT_LBRACK: //string,array,map
			iExpr := &indexExpr{x: e, lbpos: p.cur.p, index: nil, colon: 0, index1: nil, rbpos: 0}
			p.next()
			iExpr.index = p.parseExpr()
			if p.cur.t == DT_COLON {
				iExpr.colon = DT_COLON
				p.next()
				if p.cur.t != DT_RBRACK {
					iExpr.index1 = p.parseExpr()
				}
			}
			iExpr.rbpos = p.expect(DT_RBRACK, `语法错误:缺少符号"]"`).p
			e = iExpr
		case DT_LPAREN:
			ce := &callExpr{x: e, lppos: p.cur.p, name: id.v}
			p.next()
			if p.cur.t == DT_RPAREN {
				ce.args = nil
			} else {
				ce.args = p.parseValues()
			}
			ce.rppos = p.expect(DT_RPAREN, `语法错误:缺少符号 ")"`).p
			e = ce
		}
	}
	return
}
func (p *parser) parseGlobalExpr() (e expr) {
	g := &globalIdentExpr{pos: p.cur.p}
	p.next()
	g.period = p.expect(DT_PERIOD, `全局变量定义错误:缺少符号"."`).p
	id := p.expect(DT_IDENT, "语法错误:缺少标识符")
	if p.cur.t == DT_LPAREN {
		call := &globalCallExpr{}
		call.period = g.period
		call.name = id.v
		call.lppos = p.expect(DT_LPAREN, `语法错误:缺少符号 ")"`).p
		if p.cur.t == DT_RPAREN {
			call.args = nil
		} else {
			call.args = p.parseValues()
		}
		call.rppos = p.expect(DT_RPAREN, `语法错误:缺少符号 ")"`).p
		return call
	} else {
		g.ident = &ident{pos: id.p, name: id.v, data: nil}
		return g
	}
}
func (p *parser) parseDecVar() (s stmt) {
	v_p := p.expect(DT_VAR, "").p
	idents := p.parseIdents()
	op_p := p.expect(DT_ASSIGN, `变量声明错误:变量声明时必须赋值缺少符号"="`).p
	values := p.parseValues()
	return &varDecStmt{v_p, idents, op_p, values}
}
func (p *parser) parseIdents() (ids []*ident) {
	first := p.expect(DT_IDENT, "语法错误:缺少标识符")
	ids = append(ids, &ident{p.cur.p, first.v, true, nil})
	for p.cur.t == DT_COMMA {
		p.next()
		if p.cur.t != DT_IDENT {
			p.error("语法错误:缺少标识符")
		}
		ids = append(ids, &ident{p.cur.p, p.cur.v, true, nil})
		p.next()
	}
	return
}
func (p *parser) parseValues() (exprs []expr) {
	value := p.parseExpr()
	exprs = append(exprs, value)
	for p.cur.t == DT_COMMA {
		p.next()
		value = p.parseExpr()
		exprs = append(exprs, value)
	}
	return
}
func (p *parser) parseDecFunc() (s stmt) {
	fun := &funDecStmt{}
	p.expect(DT_FUNC, "缺少关键字 func")
	fun.name = p.expect(DT_IDENT, "语法错误:func后缺少标识符").v
	fun.lbpos = p.expect(DT_LPAREN, "缺少符号 (").p
	if p.cur.t == DT_RPAREN {
		fun.lbpos = p.expect(DT_RPAREN, "缺少符号 )").p
	} else {
		fun.args = p.parseIdents()
		fun.lbpos = p.expect(DT_RPAREN, "缺少符号 )").p
	}
	fun.body = p.parseBlockStmt(true, false)
	return fun
}
func (p *parser) parseBlockStmt(isfunc, isfor bool) (s *blockStmt) {
	block := &blockStmt{}
	block.left = p.expect(DT_LBRACE, "缺少符号 {").p
loop:
	for {
		switch p.cur.t {
		case DT_EOF, DT_RBRACE, DT_RIGHTDEL:
			break loop
		default:
			block.list = append(block.list, p.parseStmt(isfunc, isfor))
		}
	}
	block.right = p.expect(DT_RBRACE, "缺少符号 }").p
	return block
}
func (p *parser) parseAssignStmt() (s stmt) {
	//TODO 暂时先不管吧，之后细节出来了在考虑强制左右手类型
	lhs := p.parseExpr()
	if p.cur.t == DT_ASSIGN {
		p.next()
		s = &assignStmt{lhs: lhs, opPos: p.cur.p, op: p.cur.t, rhs: p.parseExpr()}
	} else {
		switch call := lhs.(type) {
		case *globalCallExpr, *memBerCallExpr, *callExpr:
			s = &exprStmt{x: call}
		default:
			s = &badStmt{s: lhs.begin(), e: lhs.end(), text: "非法的表达式调用"}
		}
	}
	return s
}
func (p *parser) parseIfStmt(isfunc, isfor bool) (s stmt) {
	ifstmt := &ifStmt{}
	ifstmt.pos = p.expect(DT_IF, "").p
	ifstmt.cond = p.parseExpr()
	ifstmt.body = p.parseBlockStmt(isfunc, isfor)
	if p.cur.t == DT_ELSE {
		p.next()
		switch p.cur.t {
		case DT_LBRACE:
			ifstmt.branch = p.parseBlockStmt(isfunc, isfor)
		case DT_IF:
			ifstmt.branch = p.parseIfStmt(isfunc, isfor)
		}
	}
	return ifstmt
}
func (p *parser) parseForStmt(isfunc bool) (s stmt) {
	pos := p.expect(DT_FOR, "").p
	ids := p.parseIdents()
	switch len(ids) {
	case 1:
		fors := &forStmt{}
		fors.pos = pos
		fors.key = ids[0]
		fors.keywords = p.expect(DT_RANGE, `语法错误:for语句缺少关键字"range"`).p
		p.expect(DT_LPAREN, `语法错误:for i range() 缺少符号"("`)
		ics := p.parseValues()
		if len(ics) != 3 {
			p.error("语法错误:range内需要3个数字类型例如range(1,100,1)")
		}
		p.expect(DT_RPAREN, `语法错误:for i range() 缺少符号")"`)
		fors.init, fors.cond, fors.step = ics[0], ics[1], ics[2]
		fors.body = p.parseBlockStmt(isfunc, true)
		return fors
	case 2:
		forr := &forRange{}
		forr.pos = pos
		forr.key, forr.val = ids[0], ids[1]
		forr.keywords = p.expect(DT_RANGE, `语法错误:for语句缺少关键字"range"`).p
		forr.target = p.parseExpr()
		forr.body = p.parseBlockStmt(isfunc, true)
		return forr
	default:
		p.error("语法错误:for语句只能是 for i range()或者for k,v range ident")
		return
	}
}
func (p *parser) parseReturnStmt() (s stmt) {
	r := &returnStmt{}
	r.pos = p.expect(DT_RETURN, "").p
	switch p.cur.t {
	case DT_LBRACE, DT_LPAREN, DT_LBRACK,
		DT_IDENT, DT_GLOBAL, DT_STRING_LIT,
		DT_RAWSTRING_LIT, DT_INT_LIT, DT_FLOAT_LIT, DT_BOOL_LIT, DT_NULL:
		r.x = p.parseValues()
	default:
		r.x = nil
	}
	return r
}
func (p *parser) parseEcho() (s stmt) {
	echo := &echoStmt{}
	echo.pos = p.expect(DT_ECHO, "").p
	echo.lbpos = p.expect(DT_LPAREN, `语法错误:@语句缺少符号"("`).p
	echo.exprs = p.parseValues()
	echo.rbpos = p.expect(DT_RPAREN, `语法错误:@语句缺少符号")"`).p
	return echo
}
func (p *parser) parseInclude(deep *int, r *relation) (s stmt) {
	if *deep > 4 {
		p.error("嵌套层数过深")
	}
	*deep++
	icd := &include{}
	p.expect(DT_INCLUDE, "")
	str, e := strconv.Unquote(p.expect(DT_STRING_LIT, `语法错误:include语句缺少文件名，语法应该是#"targetName"(expr....)`).v)
	if e != nil {
		p.error("不合法的文件路径")
	}
	icd.name = str
	child := &relation{
		Name:  str,
		Child: make([]*relation, 0),
	}
	p.relation.insertRelation(child)
	icd.lbpos = p.expect(DT_LPAREN, `语法错误:include语句缺少符号"("，语法应该是#"targetName"(expr....)`).p
	if p.cur.t != DT_RPAREN {
		icd.args = p.parseValues()
	}
	parse, err := newInclude(str, p.left, p.right, p.readFn, child)
	if err != nil {
		p.error(err.Error())
	}
	parse.deep = p.deep
	if trees, err := parse.start(); err != nil {
		p.error(err.Error())
	} else {
		icd.tree = trees
	}
	icd.rbpos = p.expect(DT_RPAREN, `语法错误:include语句缺少符号")"，语法应该是#"targetName"(expr....)`).p
	*deep = 0
	return icd
}
func (p *parser) parseBreakStmt() (s stmt) {
	b := &breakStmt{p.cur.p}
	p.next()
	return b
}
