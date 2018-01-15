package template

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
)

type template struct {
	name    string
	parser  *parser
	reader  inputReader
	writer  io.Writer
	scope   *scope //解析器在解析时需要的作用域
	context interface{}
	tree    *[]node
	size    int
	err     error
	check   map[string]bool
}

func NewTemplate(name string, reader inputReader, writer io.Writer) *template {
	return &template{name: name, writer: writer, reader: reader}
}
func (t *template) Compile() error {
	parser, err := newParser(t.name, "{%", "%}", t.reader)
	if err != nil {
		return err
	}
	t.parser = parser
	tree, err := t.parser.start()
	if err != nil {
		return err
	}
	t.tree = &tree
	return nil
}
func createScope(s *scope) *scope {
	return newScope(s)
}
func closeScope(s *scope) {
	s = nil
}
func (t *template) checkIdent(s *scope, i ...*ident) {
	for _, v := range i {
		if id, ok := s.isExist(v.name); ok {
			if !id.canset && id.name != "args" {
				t.error(v, "声明错误:%s是关键字或函数库名不能使用,请更换变量名", v.name)
			}
		}
	}
}
func (t *template) Exec(data interface{}) error {
	t.check = make(map[string]bool)
	t.size = 3 << 20
	t.context=data
	t.scope = newTopScope(data)
	t.executeWriteByte()
	defer closeScope(t.scope)
	return t.err
}
func (t *template) executeWriteByte() {
	defer recoverHandle(&t.err)
	if t.tree != nil {
		r, b := false, false
		t.execStmts(&b, &r, *t.tree, t.scope)
	}
}
func (t *template) error(e node, format string, args ...interface{}) {
	err_msg := ""
	if e != nil {
		msg := fmt.Sprintf(format, args...)
		err_msg = fmt.Sprintf("文件名称:%s 开始位:%d 终止位:%d =>%s", t.name, e.begin(), e.end(), msg)
	} else {
		err_msg = fmt.Sprintf(format, args...)
	}
	panic(err_msg)
}
func (t *template) execStmts(b, r *bool, n []node, s *scope) (val interface{}, outNum int) {
	for i := 0; i < len(n) && !*b && !*r; i++ {
		switch stmt := n[i].(type) {
		case *varDecStmt:
			t.execDecVarStmt(stmt, s)
		case *assignStmt:
			t.execAssignStmt(stmt, s)
		case *ifStmt:
			news := createScope(s)
			t.execIfStmt(b, r, stmt, news)
			closeScope(news)
		case *forStmt:
			news := createScope(s)
			t.execForStmt(b, r, stmt, news)
			closeScope(news)
		case *forRange:
			news := createScope(s)
			t.execForRange(b, r, stmt, news)
			closeScope(news)
		case *echoStmt:
			t.execEcho(stmt, s)
		case *include:
			t.execInclude(b, r, stmt, s)
		case *blockStmt:
			val, outNum = t.execStmts(b, r, stmt.list, s)
		case *breakStmt:
			*b = true
		case *returnStmt:
			val, outNum = t.execReturnStmt(stmt, s)
			*r = true
		case *htmlExpr:
			t.writer.Write([]byte(stmt.val))
		case *funDecStmt:
			t.execDecFuncStmt(stmt, s)
		case *exprStmt:
			t.evalExprStmt(stmt, s)
		case *badStmt:
			t.error(stmt, stmt.text)
		default:
			t.error(stmt, "不支持的语句调用,联系下开发人员,看下是否需要增加此语法")
		}
	}
	return
}
func (t *template) execDecVarStmt(d *varDecStmt, s *scope) {
	valus := make([]interface{}, 0)
	for k, _ := range d.x {
		switch v, n := t.evalExpr(d.x[k], s); {
		case n == 0:
			t.error(d.x[k], "语法错误:表达式没有任何返回值")
			return
		case n == 1:
			valus = append(valus, v)
		case n > 1:
			if fv, ok := v.([]reflect.Value); ok {
				for _, v := range fv {
					valus = append(valus, v.Interface())
				}
			}
		}
	}
	if len(d.ident) != len(valus) {
		t.error(d.ident[0], "声明错误:右手表达式返回的值的数量为%d个与变量声明的数量不一致", len(valus))
	}
	for k, _ := range d.ident {
		if d.ident[k].name != "_" {
			t.checkIdent(s, d.ident[k])
			if id, ok := s.isExist(d.ident[k].name); ok {
				id.data = valus[k]
			} else {
				d.ident[k].canset = true
				d.ident[k].data = valus[k]
				s.insert(d.ident[k])
			}
		}
	}
}
func (t *template) execAssignStmt(d *assignStmt, s *scope) {
	switch rhs, n := t.evalExpr(d.rhs, s); {
	case n == 0:
		t.error(d.rhs, "赋值错误:操作符=右侧表达式没有返回任何值")
		return
	case n == 1:
		switch expr := d.lhs.(type) {
		case *ident:
			if v, ok := s.isExist(expr.name); ok {
				if v.canset {
					v.data = rhs
				} else {
					t.error(expr, "赋值错误:赋值语句只能作用在自定义变量上")
				}
			} else {
				t.error(expr, "赋值错误:变量未声明无法赋值")
			}
		case *indexExpr:
			x, _ := t.evalExpr(expr.x, s)
			xval := reflect.Indirect(reflect.ValueOf(x))
			switch xval.Kind() {
			case reflect.Slice:
				low, n := t.evalExpr(expr.index, s)
				t.checkExprValue(expr.index, &low, n, 1)
				index := t.toInt(expr.index, reflect.ValueOf(low))
				if expr.index1 != nil {
					t.error(expr.index1, "赋值错误:数组重新赋值,角标个数存在多个")
					return
				}
				if xval.Index(index).Kind() != reflect.Interface && xval.Index(index).Kind() != reflect.ValueOf(rhs).Kind() {
					t.error(expr.index, "赋值错误:有类型的数组重新赋值时必须保持与声明一致声明时为%s类型", xval.Index(index).Kind().String())
				}
				xval.Index(index).Set(reflect.ValueOf(rhs))
			case reflect.Map:
				if expr.index1 != nil {
					t.error(expr.index1, "赋值错误:map重新赋值,key的定义不明确")
				}
				key, n := t.evalExpr(expr.index, s)
				t.checkExprValue(expr.index, &key, n, 1)
				kv := reflect.ValueOf(key)
				if kv.Kind() != reflect.String {
					t.error(expr.index, "赋值错误:key只能是string类型")
				}
				xval.SetMapIndex(kv, reflect.ValueOf(rhs))
			default:
				t.error(expr, "语法错误:元素赋值只能是数组和map类型")
			}
		case *memberExpr:
			a, _ := t.evalExpr(expr, s)
			if reflect.ValueOf(a).Kind() == reflect.Func {
				t.error(expr, "赋值错误:赋值语句无法作用于全局函数上")
			} else {
				a = rhs
			}
		case *globalIdentExpr:
			s.insertGlobal(&ident{name: expr.ident.name, canset: true, data: rhs})
		default:
			t.error(d.lhs, "赋值错误:赋值语句只能作用在自定义变量上")
			return
		}
	case n > 1:
		t.error(d.rhs, "赋值错误:操作符=右侧表达式的返回值超过1个")
	}
}
func (t *template) execIfStmt(b, r *bool, d *ifStmt, s *scope) {
	cond, n := t.evalExpr(d.cond, s)
	t.checkExprValue(d.cond, &cond, n, 1)
	if t.toBool(d.cond, reflect.ValueOf(cond)) {
		t.execStmts(b, r, d.body.list, s)

	} else {
		if d.branch != nil {
			t.execStmts(b, r, d.branch.(*blockStmt).list, s)
		}
	}
}
func (t *template) execForStmt(b, r *bool, d *forStmt, s *scope) {
	init, n := t.evalExpr(d.init, s)
	t.checkExprValue(d.init, &init, n, 1)
	cond, n := t.evalExpr(d.cond, s)
	t.checkExprValue(d.cond, &cond, n, 1)
	step, n := t.evalExpr(d.step, s)
	t.checkExprValue(d.step, &step, n, 1)
	x := t.toInt(d.init, reflect.ValueOf(init))
	y := t.toInt(d.cond, reflect.ValueOf(cond))
	z := t.toInt(d.step, reflect.ValueOf(step))
	if x < y && z <= 0 {
		t.error(d.step, "语法错误:步长不能为0,为0则代表死循环，模板中不允许")
	}
	t.checkIdent(s, d.key)
	s.insert(d.key)
	for ; x < y; x += z {
		d.key.data = x
		t.execStmts(b, r, d.body.list, s)
	}
	*b = false
}
func (t *template) execForRange(b, r *bool, d *forRange, s *scope) {
	target, n := t.evalExpr(d.target, s)
	t.checkExprValue(d.target, &target, n, 1)
	tv := reflect.Indirect(reflect.ValueOf(target))
	t.checkIdent(s, d.key)
	t.checkIdent(s, d.val)
	switch tv.Kind() {
	case reflect.Slice:
		d.key.data = -1
		d.val.data = nil
		s.insert(d.key, d.val)
		for i := 0; i < tv.Len(); i++ {
			d.key.data = i
			d.val.data = tv.Index(i).Interface()
			t.execStmts(b, r, d.body.list, s)
		}
		*b = false
	case reflect.Map:
		d.key.data = ""
		d.val.data = nil
		s.insert(d.key, d.val)
		for _, v := range tv.MapKeys() {
			d.key.data = v.Interface()
			d.val.data = tv.MapIndex(v).Interface()
			t.execStmts(b, r, d.body.list, s)
		}
		*b = false
	default:
		t.error(d.target, "类型错误:目标不是数组或者map类型无法迭代")
	}
}
func (t *template) execEcho(d *echoStmt, s *scope) {
	for _, v := range d.exprs {
		switch value, n := t.evalExpr(v, s); {
		case n == 0:
			return
		case n == 1:
			t.writer.Write(t.valueToBytes(v, value))
		case n > 1:
			if fv, ok := value.([]reflect.Value); ok {
				for _, v1 := range fv {
					t.writer.Write(t.valueToBytes(v, v1.Interface()))
				}
			} else {
				t.error(v, "函数返回值转换出错")
				return
			}
		}
	}
	t.writer.Write([]byte{'\n'})
	t.size -= 1
}
func (t *template) execInclude(b, r *bool, d *include, s *scope) {
	args := make([]interface{}, 0)
	for _, v := range d.args {
		r, n := t.evalExpr(v, s)
		t.checkExprValue(v, &r, n, 1)
		args = append(args, r)
	}
	news := newIncludeScope(t.context, s.global)
	news.objects["args"] = &ident{name: "args", data: args}
	t.name = d.name
	t.execStmts(b, r, d.tree, news)
}
func (t *template) execReturnStmt(d *returnStmt, s *scope) (i interface{}, outNum int) {
	if d.x == nil {
		return nil, 0
	}
	r := make([]reflect.Value, 0)
	for _, v := range d.x {
		switch e, n := t.evalExpr(v, s); {
		case n == 1:
			r = append(r, reflect.ValueOf(e))
		case n > 1:
			if rv, ok := e.([]reflect.Value); ok {
				r = append(r, rv...)
			} else {
				t.error(v, "返回值处理出错，联系开发人员")
			}
		}
	}
	switch n := len(r); {
	case n == 1:
		return r[0].Interface(), 1
	case n > 1:
		return r, n
	default:
		return nil, 0
	}

}
func (t *template) execDecFuncStmt(d *funDecStmt, s *scope) {
	s.insertGlobal(&ident{name: d.name, data: d, canset: false})
}
func (t *template) evalExprStmt(d *exprStmt, s *scope) (i interface{}, outNum int) {
	return t.evalExpr(d.x, s)
}
func (t *template) valueToBytes(e expr, i interface{}) (b []byte) {
	if t.size <= 0 {
		t.error(e, "页面大小已经超出了正常范围,强制停止")
	}
	value := reflect.ValueOf(i)
	switch value.Kind() {
	case reflect.Map:
		bj, err := json.Marshal(i)
		if err != nil {
			t.error(e, "map转换json格式字符串出错！")
		}
		b = bj
	case reflect.Slice:
		if value.Len() == 1 {
			b = []byte(fmt.Sprint(value.Index(0)))
		} else {
			bj, err := json.Marshal(value.Interface())
			if err != nil {
				t.error(e, "数组转换json格式字符串出错！")
			}
			b = bj
		}
	case reflect.String:
		b = append(b, []byte(value.String())...)
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		b = []byte(fmt.Sprint(value.Int()))
	case reflect.Float32, reflect.Float64:
		b = []byte(fmt.Sprint(value.Float()))
	case reflect.Struct:
		b = []byte(fmt.Sprintf("%#v", i))
	case reflect.Ptr:
		if value.Elem().IsValid() {
			b = t.valueToBytes(e, value.Elem().Interface())
		} else {
			b = []byte("null")
		}
	case reflect.Invalid:
		b = []byte("null")
	default:
		b = []byte(fmt.Sprintf("%#v", i))
	}
	t.size -= len(b)
	return b
}
func (t *template) evalExpr(e expr, s *scope) (i interface{}, outNum int) {
	switch expr := e.(type) {
	case *ident:
		return t.evalIdent(expr, s)
	case *literal:
		return t.evalLiteral(expr, s)
	case *slice:
		return t.evalSliceExpr(expr, s)
	case *object:
		return t.evalObjectExpr(expr, s)
	case *unaryExpr:
		return t.evalUnaryExpr(expr, s)
	case *binaryExpr:
		return t.evalBinaryExpr(expr, s)
	case *memberExpr:
		return t.evalMemberExpr(expr, s)
	case *indexExpr:
		return t.evalIndexExpr(expr, s)
	case *memBerCallExpr:
		return t.evalMemberCallExpr(expr, s)
	case *callExpr:
		return t.evalCallExpr(expr, s)
	case *globalCallExpr:
		return t.evalGlobalCallExpr(expr, s)
	case *globalIdentExpr:
		return t.evalGlobalIdentExpr(expr, s)

	default:
		t.error(expr, "未知的表达式错误,请联系引擎开发人员")
		return nil, 0
	}
}
func (t *template) evalIdent(e *ident, s *scope) (i interface{}, outNum int) {
	id, ok := s.isExist(e.name)
	if !ok {
		t.error(e, "语法错误:使用了未定义的变量 %s", e.name)
	}
	return id.data, 1
}
func (t *template) evalLiteral(expr *literal, s *scope) (i interface{}, outNum int) {
	switch expr.typ {
	case DT_STRING_LIT, DT_RAWSTRING_LIT:
		i = expr.val
	case DT_INT_LIT:
		int, e := strconv.Atoi(expr.val)
		if e != nil {
			t.error(expr, "字面量错误:%s 不是一个int", expr.val)
		}
		i = int
	case DT_FLOAT_LIT:
		float, e := strconv.ParseFloat(expr.val, 64)
		if e != nil {
			t.error(expr, "字面量错误:%s 不是一个float", expr.val)
		}
		i = float
	case DT_BOOL_LIT:
		b, e := strconv.ParseBool(expr.val)
		if e != nil {
			t.error(expr, "字面量错误:%s 不是一个布尔值", expr.val)
		}
		i = b
	case DT_NULL:
		i = nil
	}
	return i, 1
}
func (t *template) checkExprValue(e expr, in interface{}, inNum, expectNum int) {
	if inNum != expectNum {
		t.error(e, "返回值错误:表达式返回值个数不符合预期值")
		return
	}
	if in == nil {
		return
	}
	if inNum > 1 {
		//TODO .....
	}
}
func (t *template) evalSliceExpr(expr *slice, s *scope) (i interface{}, outNum int) {
	if expr.typ > 0 {
		switch expr.typ {
		case DT_STRING:
			slice := make([]string, len(expr.element))
			for k, v := range expr.element {
				value, n := t.evalExpr(v, s)
				t.checkExprValue(v, &value, n, 1)
				slice[k] = t.toString(v, reflect.ValueOf(value))
			}
			return slice, 1
		case DT_BOOL:
			slice := make([]bool, len(expr.element))
			for k, v := range expr.element {
				value, n := t.evalExpr(v, s)
				t.checkExprValue(v, &value, n, 1)
				slice[k] = t.toBool(v, reflect.ValueOf(value))
			}
			return slice, 1
		case DT_INT:
			slice := make([]int, len(expr.element))
			for k, v := range expr.element {
				value, n := t.evalExpr(v, s)
				t.checkExprValue(v, &value, n, 1)
				slice[k] = t.toInt(v, reflect.ValueOf(value))
			}
			return slice, 1
		case DT_FLOAT:
			slice := make([]float64, len(expr.element))
			for k, v := range expr.element {
				value, n := t.evalExpr(v, s)
				t.checkExprValue(v, &value, n, 1)
				slice[k] = t.toFloat(v, reflect.ValueOf(value))
			}
			return slice, 1
		default:
			return nil, 0
		}
	} else {
		slice := make([]interface{}, len(expr.element))
		for k, v := range expr.element {
			value, n := t.evalExpr(v, s)
			t.checkExprValue(v, &value, n, 1)
			slice[k] = value
		}
		return slice, 1
	}
}
func (t *template) evalObjectExpr(expr *object, s *scope) (i interface{}, outNum int) {
	m := make(map[string]interface{})
	for _, v := range expr.properties {
		value, n := t.evalExpr(v.key, s)
		t.checkExprValue(v.key, &value, n, 1)
		key := reflect.ValueOf(value)
		if key.Kind() != reflect.String {
			t.error(&v, "类型错误:map的键必须是string类型")
		}
		value, n = t.evalExpr(v.value, s)
		t.checkExprValue(v.value, &value, n, 1)
		m[key.String()] = value
	}
	return m, 1
}
func (t *template) evalUnaryExpr(expr *unaryExpr, s *scope) (i interface{}, outNum int) {
	value, n := t.evalExpr(expr.e, s)
	t.checkExprValue(expr.e, &value, n, 1)
	switch expr.op {
	case DT_NOT:
		i = !t.toBool(expr, reflect.ValueOf(value))
	case DT_ADD, DT_SUB:
		switch T := value.(type) {
		case int:
			if expr.op == DT_ADD {
				i = +T
			} else {
				i = -T
			}
		case float64:
			if expr.op == DT_ADD {
				i = +T
			} else {
				i = -T
			}
		default:
			t.error(expr.e, "类型错误:操作数不是一个数字类型")
		}
	case DT_INC, DT_DEC:
		switch T := value.(type) {
		case int:
			if expr.op == DT_INC {
				T++
			} else {
				T--
			}
			i = T
		case float64:
			if expr.op == DT_INC {
				T++
			} else {
				T--
			}
			i = T
		default:
			t.error(expr.e, "类型错误:操作数不是一个数字类型")
		}
	default:
		t.error(expr, "语法错误:非法的一元表达式")
	}
	return i, 1
}
func (t *template) evalBinaryExpr(expr *binaryExpr, s *scope) (i interface{}, outNum int) {
	x, n := t.evalExpr(expr.left, s)
	t.checkExprValue(expr.left, &x, n, 1)
	y, n := t.evalExpr(expr.right, s)
	t.checkExprValue(expr.right, &y, n, 1)
	xval := reflect.ValueOf(x)
	yval := reflect.ValueOf(y)
	switch expr.op {
	case DT_LOGIC_AND:
		return t.toBool(expr.left, xval) && t.toBool(expr.right, yval), 1
	case DT_LOGIC_OR:
		if t.toBool(expr.left, xval) {
			return xval.Interface(), 1
		}
		if t.toBool(expr.right, yval) {
			return yval.Interface(), 1
		}
		return nil, 1
	case DT_LSS, DT_LEQ, DT_GTR, DT_GEQ, DT_ADD, DT_SUB, DT_MUL, DT_QUO, DT_REM:
		if xval.Kind() != yval.Kind() {
			t.error(expr.right, "类型错误:%s!=%s二元表达式左右手类型不一致", xval.Type().String(), yval.Type().String())
		}
		switch expr.op {
		case DT_LSS: // <
			return t.evalOpLSS(expr, xval, yval), 1
		case DT_LEQ: // <=
			return t.evalOpLEQ(expr, xval, yval), 1
		case DT_GTR: // >
			return t.evalOpGTR(expr, xval, yval), 1
		case DT_GEQ: // >=
			return t.evalOpGEQ(expr, xval, yval), 1
		case DT_ADD: // +
			return t.evalOpADD(expr, xval, yval), 1
		case DT_SUB: // -
			return t.evalOpSUB(expr, xval, yval), 1
		case DT_MUL: // *
			return t.evalOpMUL(expr, xval, yval), 1
		case DT_QUO: // /
			return t.evalOpQUO(expr, xval, yval), 1
		case DT_REM: // %/
			return t.evalOpREM(expr, xval, yval), 1
		}
	case DT_EQL: //==
		return t.evalOpEQL(expr, xval, yval), 1
	case DT_NEQ: // !=
		return t.evalOpNEQ(expr, xval, yval), 1
	}
	return
}
func (t *template) evalOpEQL(expr *binaryExpr, xval, yval reflect.Value) (b bool) {
	var x, y interface{}
	if xval.IsValid() {
		x = xval.Interface()
	}
	if yval.IsValid() {
		y = yval.Interface()
	}
	return reflect.DeepEqual(x, y)
}
func (t *template) evalOpNEQ(expr *binaryExpr, xval, yval reflect.Value) (b bool) {
	var x, y interface{}
	if xval.IsValid() {
		x = xval.Interface()
	}
	if yval.IsValid() {
		y = yval.Interface()
	}
	return !reflect.DeepEqual(x, y)
}
func (t *template) evalOpLSS(expr *binaryExpr, xval, yval reflect.Value) (b bool) {
	switch xval.Kind() {
	case reflect.Int:
		b = t.toInt(expr.left, xval) < t.toInt(expr.right, yval)
	case reflect.Float64:
		b = t.toFloat(expr.left, xval) < t.toFloat(expr.right, yval)
	default:
		t.error(expr.right, `类型错误:操作符"<"号只支持int||float`)
	}
	return
}
func (t *template) evalOpLEQ(expr *binaryExpr, xval, yval reflect.Value) (b bool) {
	switch xval.Kind() {
	case reflect.Int:
		b = t.toInt(expr.left, xval) <= t.toInt(expr.right, yval)
	case reflect.Float64:
		b = t.toFloat(expr.left, xval) <= t.toFloat(expr.right, yval)
	default:
		t.error(expr.right, `类型错误:操作符"<="号只支持int||float`)
	}
	return
}
func (t *template) evalOpGTR(expr *binaryExpr, xval, yval reflect.Value) (b bool) {
	switch xval.Kind() {
	case reflect.Int:
		b = t.toInt(expr.left, xval) > t.toInt(expr.right, yval)
	case reflect.Float64:
		b = t.toFloat(expr.left, xval) > t.toFloat(expr.right, yval)
	default:
		t.error(expr.right, `类型错误:操作符">"号只支持int||float`)
	}
	return
}
func (t *template) evalOpGEQ(expr *binaryExpr, xval, yval reflect.Value) (b bool) {
	switch xval.Kind() {
	case reflect.Int:
		b = t.toInt(expr.left, xval) >= t.toInt(expr.right, yval)
	case reflect.Float64:
		b = t.toFloat(expr.left, xval) >= t.toFloat(expr.right, yval)
	default:
		t.error(expr.right, `类型错误:操作符">="号只支持int||float`)
	}
	return
}
func (t *template) evalOpADD(expr *binaryExpr, xval, yval reflect.Value) (b interface{}) {
	switch xval.Kind() {
	case reflect.Int:
		b = t.toInt(expr.left, xval) + t.toInt(expr.right, yval)
	case reflect.Float64:
		b = t.toFloat(expr.left, xval) + t.toFloat(expr.right, yval)
	case reflect.String:
		b = t.toString(expr.left, xval) + t.toString(expr.right, yval)
	case reflect.Slice:
		if reflect.Indirect(xval).Type() != reflect.Indirect(yval).Type() {
			t.error(expr.right, "类型错误:%s!=%s带类型的数组相加时请确定操作符左右类型一致", reflect.Indirect(xval).Type(), reflect.Indirect(yval).Type())
		}
		b = reflect.AppendSlice(reflect.Indirect(xval), reflect.Indirect(yval)).Interface()
	default:
		t.error(expr.right, `类型错误:操作符"+"号只支持int||float||string||array`)
	}
	return
}
func (t *template) evalOpSUB(expr *binaryExpr, xval, yval reflect.Value) (b interface{}) {
	switch xval.Kind() {
	case reflect.Int:
		b = t.toInt(expr.left, xval) - t.toInt(expr.right, yval)
	case reflect.Float64:
		b = t.toFloat(expr.left, xval) - t.toFloat(expr.right, yval)
	default:
		t.error(expr.right, `类型错误:操作符"-"号只支持int||float`)
	}
	return
}
func (t *template) evalOpMUL(expr *binaryExpr, xval, yval reflect.Value) (b interface{}) {
	switch xval.Kind() {
	case reflect.Int:
		b = t.toInt(expr.left, xval) * t.toInt(expr.right, yval)
	case reflect.Float64:
		b = t.toFloat(expr.left, xval) * t.toFloat(expr.right, yval)
	default:
		t.error(expr.right, `类型错误:操作符"*"号只支持int||float`)
	}
	return
}
func (t *template) evalOpQUO(expr *binaryExpr, xval, yval reflect.Value) (b interface{}) {
	switch xval.Kind() {
	case reflect.Int:
		b = t.toInt(expr.left, xval) / t.toInt(expr.right, yval)
	case reflect.Float64:
		b = t.toFloat(expr.left, xval) / t.toFloat(expr.right, yval)
	default:
		t.error(expr.right, `类型错误:操作符"/"号只支持int||float`)
	}
	return
}
func (t *template) evalOpREM(expr *binaryExpr, xval, yval reflect.Value) (b interface{}) {
	switch xval.Kind() {
	case reflect.Int:
		b = t.toInt(expr.left, xval) % t.toInt(expr.right, yval)
	default:
		t.error(expr.right, `类型错误:操作符"%"号只支持int`)
	}
	return
}
func (t *template) evalMemberExpr(expr *memberExpr, s *scope) (i interface{}, outNum int) {
	val, _ := t.evalExpr(expr.x, s)
	ident := reflect.Indirect(reflect.ValueOf(val))
	defer func(name string) {
		if err := recover(); err != nil {
			t.error(expr.x, "语法错误:不包含%s这个字段", name)
		}
	}(expr.sel.name)
	return ident.FieldByName(expr.sel.name).Interface(), 1
}
func (t *template) evalIndexExpr(expr *indexExpr, s *scope) (i interface{}, outNum int) {
	x, n := t.evalExpr(expr.x, s)
	t.checkExprValue(expr.x, &x, n, 1)
	xval := reflect.Indirect(reflect.ValueOf(x))
	switch xval.Kind() {
	case reflect.Slice:
		low, n := t.evalExpr(expr.index, s)
		t.checkExprValue(expr.index, &low, n, 1)
		index := t.toInt(expr.index, reflect.ValueOf(low))
		index1 := -1
		if expr.index1 == nil {
			index1 = xval.Len()
		} else {
			high, n := t.evalExpr(expr.index1, s)
			t.checkExprValue(expr.index1, &low, n, 1)
			index1 = t.toInt(expr.index1, reflect.ValueOf(high))
		}
		if index+1 > xval.Len() {
			t.error(expr.index, "语法错误:索引越界")
		}
		if index1 > xval.Len() {
			t.error(expr.index1, "语法错误:索引越界")
		}
		if index1 < index {
			t.error(expr.index1, "语法错误:右侧索引值小于左侧索引值")
		}
		if expr.colon > 0 {
			return xval.Slice(index, index1).Interface(), 1
		} else {
			return xval.Index(index).Interface(), 1
		}
	case reflect.String:
		low, n := t.evalExpr(expr.index, s)
		t.checkExprValue(expr.index, &low, n, 1)
		index := t.toInt(expr.index, reflect.ValueOf(low))
		index1 := -1
		if expr.index1 == nil {
			index1 = xval.Len()
		} else {
			high, n := t.evalExpr(expr.index1, s)
			t.checkExprValue(expr.index1, &low, n, 1)
			index1 = t.toInt(expr.index1, reflect.ValueOf(high))
		}
		if index > xval.Len() {
			t.error(expr.index, "语法错误:索引越界")
		}
		if index1 > xval.Len() {
			t.error(expr.index1, "语法错误:索引越界")
		}
		if index1 < index {
			t.error(expr.index1, "语法错误:右侧索引值小于左侧索引值")
		}
		if expr.colon > 0 {
			return string([]rune(xval.String()[index:index1])), 1
		} else {
			return string([]rune(xval.String())[index]), 1
		}
	case reflect.Map:
		if expr.index1 != nil {
			t.error(expr.index1, "语法错误:map[]只能有一个键值")
		}
		index, _ := t.evalExpr(expr.index, s)
		indexval := reflect.ValueOf(index)
		if indexval.Kind() != reflect.String {
			t.error(expr.index, "类型错误:键必须是string类型")
		}
		if v := xval.MapIndex(indexval); v.IsValid() {
			return v.Interface(), 1
		} else {
			return nil, 1
		}
	default:
		t.error(expr, "类型错误:不是JSon||Array||string类型")
	}
	return
}
func (t *template) evalMemberCallExpr(expr *memBerCallExpr, s *scope) (i interface{}, outNum int) {
	a, b := t.evalExpr(expr.x, s)
	if a == nil {
		t.error(expr.x, "类型错误:%s的调用对象是null,请检查", expr.name)
	}
	t.checkExprValue(expr.x, &a, b, 1)
	obj := reflect.ValueOf(a)
	fnObj := obj.MethodByName(expr.name)
	if fnObj.Kind() != reflect.Func {
		t.error(expr.x, "类型错误:%s不是一个函数 ", expr.name)
	}
	fn := fnObj.Type()
	callValues := make([]reflect.Value, 0)
	callValues = fnObj.Call(t.checkParams(fn, expr, s))
	if fn.NumOut() == 1 {
		return callValues[0].Interface(), 1
	} else {
		return callValues, len(callValues)
	}
}
func (t *template) evalCallExpr(expr *callExpr, s *scope) (i interface{}, outNum int) {
	a, b := t.evalExpr(expr.x, s)
	if a == nil {
		t.error(expr.x, "类型错误:%s是null无法调用", expr.name)
	}
	t.checkExprValue(expr.x, &a, b, 1)
	fnObj := reflect.ValueOf(a)
	if fnObj.Kind() != reflect.Func {
		t.error(expr.x, "类型错误:%s不是一个函数", expr.name)
	}
	fn := fnObj.Type()
	callValues := make([]reflect.Value, 0)
	callValues = fnObj.Call(t.checkParams(fn, expr, s))
	if fn.NumOut() == 1 {
		return callValues[0].Interface(), 1
	} else {
		return callValues, len(callValues)
	}
}
func (t *template) checkParams(fn reflect.Type, e expr, s *scope) []reflect.Value {
	fnName := ""
	exprs := []expr{}
	switch t := e.(type) {
	case *callExpr:
		exprs = t.args
		fnName = t.name
	case *memBerCallExpr:
		exprs = t.args
		fnName = t.name
	}
	args := make([]reflect.Value, 0)
	if fn.IsVariadic() {
		if fn.NumIn()-1 > len(exprs) {
			t.error(e, "参数错误:%s传参个数与函数定义的不一致,声明为:%s", fnName, fn.String())
		}
		switch fn.NumIn() {
		case 1:
			for _, v := range exprs {
				value, n := t.evalExpr(v, s)
				t.checkExprValue(v, &value, n, 1)
				switch fn.In(0).Elem().Kind() {
				case reflect.Interface:
					if value == nil {
						args = append(args, reflect.Zero(reflect.TypeOf((*interface{})(nil)).Elem()))
					} else {
						args = append(args, reflect.ValueOf(value))
					}
				case reflect.Slice:
					if reflect.ValueOf(value).Kind() == reflect.Slice && reflect.ValueOf(value).Type().Elem().Kind() != fn.In(0).Elem().Elem().Kind() {
						t.error(e, "参数错误:%s传参类型与函数定义的不一致,声明为:%s", fnName, fn.String())
					}
					args = append(args, reflect.ValueOf(value))
				default:
					if fn.In(0).Elem().Kind() != reflect.ValueOf(value).Kind() {
						t.error(e, "参数错误:%s传参类型与函数定义的不一致,声明为:%s", fnName, fn.String())
					}
					args = append(args, reflect.ValueOf(value))
				}
			}
		default:
			fnN := fn.NumIn() - 1
			for k := 0; k < fnN; k++ {
				value, n := t.evalExpr(exprs[k], s)
				t.checkExprValue(exprs[k], &value, n, 1)
				switch fn.In(k).Kind() {
				case reflect.Interface:
					if value == nil {
						args = append(args, reflect.Zero(reflect.TypeOf((*interface{})(nil)).Elem()))
					} else {
						args = append(args, reflect.ValueOf(value))
					}
				default:
					if fn.In(k).Kind() != reflect.ValueOf(value).Kind() {
						t.error(e, "参数错误:%s传参类型与函数定义的不一致,声明为:%s", fnName, fn.String())
					}
					if fn.In(k).Kind() == reflect.Slice && fn.In(k).Elem().Kind() != reflect.ValueOf(value).Type().Elem().Kind() {
						t.error(e, "参数错误:%s传参类型与函数定义的不一致,声明为:%s", fnName, fn.String())
					}
					args = append(args, reflect.ValueOf(value))
				}
			}
			if l := len(exprs); l > fnN {
				for i := fnN; i < l; i++ {
					value, n := t.evalExpr(exprs[i], s)
					t.checkExprValue(exprs[i], &value, n, 1)
					switch fn.In(fnN).Elem().Kind() {
					case reflect.Interface:
						if value == nil {
							args = append(args, reflect.Zero(reflect.TypeOf((*interface{})(nil)).Elem()))
						} else {
							args = append(args, reflect.ValueOf(value))
						}
					case reflect.Slice:
						if reflect.ValueOf(value).Kind() == reflect.Slice && reflect.ValueOf(value).Type().Elem().Kind() != fn.In(fnN).Elem().Elem().Kind() {
							t.error(e, "参数错误:%s传参类型与函数定义的不一致,声明为:%s", fnName, fn.String())
						}
						args = append(args, reflect.ValueOf(value))
					default:
						if fn.In(fnN).Elem().Kind() != reflect.ValueOf(value).Kind() {
							t.error(e, "参数错误:%s传参类型与函数定义的不一致,声明为:%s", fnName, fn.String())
						}
						args = append(args, reflect.ValueOf(value))
					}
				}
			}
		}
	} else {
		if fn.NumIn() != len(exprs) {
			t.error(e, "参数错误:%s传参个数与函数定义的不一致,声明为:%s", fnName, fn.String())
		}
		for k, v := range exprs {
			value, n := t.evalExpr(v, s)
			t.checkExprValue(v, &value, n, 1)
			switch fn.In(k).Kind() {
			case reflect.Interface:
				if value == nil {
					args = append(args, reflect.Zero(reflect.TypeOf((*interface{})(nil)).Elem()))
				} else {
					args = append(args, reflect.ValueOf(value))
				}
			default:
				if fn.In(k).Kind() != reflect.ValueOf(value).Kind() {
					t.error(e, "参数错误:%s传参类型与函数定义的不一致,声明为:%s", fnName, fn.String())
				}
				if fn.In(k).Kind() == reflect.Slice && fn.In(k).Elem().Kind() != reflect.ValueOf(value).Type().Elem().Kind() {
					t.error(e, "参数错误:%s传参类型与函数定义的不一致,声明为:%s", fnName, fn.String())
				}
				args = append(args, reflect.ValueOf(value))
			}
		}
	}
	return args
}
func (t *template) evalGlobalIdentExpr(e *globalIdentExpr, s *scope) (i interface{}, outNum int) {
	obj, ok := s.isExistGlobal(e.ident.name)
	if !ok {
		t.error(e, "语法错误:global.%s不存在", e.ident.name)
	}
	return obj.data, 1
}
func (t *template) evalGlobalCallExpr(e *globalCallExpr, s *scope) (i interface{}, outNum int) {
	obj, ok := s.isExistGlobal(e.name)
	if !ok {
		t.error(e, "语法错误:函数%s不存在", e.name)
	}
	if _, ok := t.check[e.name]; ok {
		t.error(e, "语法错误:函数%s存在递归调用", e.name)
	} else {
		t.check[e.name] = true
	}

	fn, ok := (obj.data).(*funDecStmt)
	if !ok {
		t.error(e, "语法错误:%s不是一个函数", e.name)
	}
	if len(fn.args) != len(e.args) {
		t.error(e, "参数错误:函数%s调用时,传参与声明时不一致", e.name)
	}
	for k, v := range e.args {
		value, n := t.evalExpr(v, s)
		t.checkExprValue(v, &value, n, 1)
		check := reflect.ValueOf(value)
		if reflect.ValueOf(value).IsValid() {
			fn.args[k].data = value
		} else {
			t.error(v, "参数错误:传入的参数是无效的数据,%s", check.Kind().String())
		}
	}
	news := createScope(s)
	t.checkIdent(s, fn.args...)
	news.insert(fn.args...)
	r, b := false, false
	i, outNum = t.runFunc(&b, &r, fn.body.list, news)
	closeScope(news)
	t.check = make(map[string]bool)
	if r == false {
		return nil, 1
	}
	return i, outNum
}
func (t *template) runFunc(b, r *bool, n []node, s *scope) (val interface{}, outNum int) {
	for i := 0; i < len(n) && !*b && !*r; i++ {
		switch stmt := n[i].(type) {
		case *varDecStmt:
			t.execDecVarStmt(stmt, s)
		case *assignStmt:
			t.execAssignStmt(stmt, s)
		case *ifStmt:
			news := createScope(s)
			cond, n := t.evalExpr(stmt.cond, news)
			t.checkExprValue(stmt.cond, &cond, n, 1)
			if t.toBool(stmt.cond, reflect.ValueOf(cond)) {
				val, outNum = t.runFunc(b, r, stmt.body.list, news)
			} else {
				if stmt.branch != nil {
					val, outNum = t.runFunc(b, r, stmt.branch.(*blockStmt).list, news)
				}
			}
			closeScope(news)
		case *forStmt:
			news := createScope(s)
			init, n := t.evalExpr(stmt.init, news)
			t.checkExprValue(stmt.init, &init, n, 1)
			cond, n := t.evalExpr(stmt.cond, news)
			t.checkExprValue(stmt.cond, &cond, n, 1)
			step, n := t.evalExpr(stmt.step, news)
			t.checkExprValue(stmt.step, &step, n, 1)
			x := t.toInt(stmt.init, reflect.ValueOf(init))
			y := t.toInt(stmt.cond, reflect.ValueOf(cond))
			z := t.toInt(stmt.step, reflect.ValueOf(step))
			if x < y && z <= 0 {
				t.error(stmt.step, "语法错误:步长不能为0,为0则代表死循环，模板中不允许")
			}
			t.checkIdent(s, stmt.key)
			news.insert(stmt.key)
			for ; x < y; x += z {
				stmt.key.data = x
				val, outNum = t.runFunc(b, r, stmt.body.list, news)
			}
			*b = false
			closeScope(news)
		case *forRange:
			news := createScope(s)
			target, n := t.evalExpr(stmt.target, news)
			t.checkExprValue(stmt.target, &target, n, 1)
			t.checkIdent(s, stmt.key)
			t.checkIdent(s, stmt.val)
			switch T := target.(type) {
			case []interface{}:
				stmt.key.data = -1
				stmt.val.data = nil
				news.insert(stmt.key, stmt.val)
				for k, v := range T {
					stmt.key.data = k
					stmt.val.data = v
					val, outNum = t.runFunc(b, r, stmt.body.list, news)
					if outNum > 0 {
						return
					}
				}
			case map[string]interface{}:
				stmt.key.data = ""
				stmt.val.data = nil
				news.insert(stmt.key, stmt.val)
				for k, v := range T {
					stmt.key.data = k
					stmt.val.data = v
					val, outNum = t.runFunc(b, r, stmt.body.list, news)
					if outNum > 0 {
						return
					}
				}
			default:
				t.error(stmt.target, "类型错误:目标不是数组或者map类型无法迭代")
			}
			*b = false
			closeScope(news)
		case *echoStmt:
			t.execEcho(stmt, s)
		case *breakStmt:
			*b = true
		case *blockStmt:
			val, outNum = t.runFunc(b, r, stmt.list, s)
		case *returnStmt:
			if stmt == nil {
				return nil, 1
			} else {
				val, outNum = t.execReturnStmt(stmt, s)
				*r = true
			}
		case *badStmt:
			t.error(stmt, stmt.text)
		case *globalCallExpr:
			val, outNum = t.evalGlobalCallExpr(stmt, s)
		default:
			t.error(stmt, "语法错误:函数语句体内包含非法语句")
		}
	}
	return
}
func (t *template) toBool(e expr, value reflect.Value) (b bool) {
	switch value.Kind() {
	case reflect.Bool:
		b = value.Bool()
	case reflect.Invalid:
		b = false
	case reflect.String:
		if value.String() != "" {
			b = true
		}
	case reflect.Float64:
		if value.Float() != 0 {
			b = true
		}
	case reflect.Int:
		if value.Int() != 0 {
			b = true
		}
	case reflect.Map:
		b = true
	case reflect.Slice:
		b = true
	case reflect.Ptr:
		if value.Pointer() > 0 {
			b = true
		}
	case reflect.Interface:
		return t.toBool(e, value.Elem())
	default:
		t.error(e, "类型错误:表达式结果无法转换为bool类型")
	}
	return b
}
func (t *template) toInt(e expr, value reflect.Value) (i int) {
	switch value.Kind() {
	case reflect.Int:
		i = value.Interface().(int)
	default:
		t.error(e, "类型错误:期望值是int类型，但表达式结果%s类型", value.Kind().String())
	}
	return i
}
func (t *template) toFloat(e expr, value reflect.Value) (f float64) {
	switch value.Kind() {
	case reflect.Float64:
		f = value.Interface().(float64)
	default:
		t.error(e, "类型错误:期望值是float类型，但表达式结果%s类型", value.Kind().String())
	}
	return f
}
func (t *template) toString(e expr, value reflect.Value) (s string) {
	switch value.Kind() {
	case reflect.String:
		s = value.String()
	default:
		t.error(e, "类型错误:期望值是string类型，但表达式结果%s类型", value.Kind().String())
	}
	return s
}
