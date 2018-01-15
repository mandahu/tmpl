package template

type token int

const (
	DT_EOF     token = 0
	DT_ILLEGAL token = iota
	DT_LEFTDEL
	DT_RIGHTDEL
	DT_HTMLTEXT
	DT_IDENT
	DT_INT_LIT
	DT_FLOAT_LIT
	DT_BOOL_LIT
	DT_STRING_LIT
	DT_RAWSTRING_LIT

	Keywords
	DT_INT
	DT_FLOAT
	DT_BOOL
	DT_STRING
	DT_ECHO
	DT_NULL
	DT_IF
	DT_ELSE
	DT_FOR
	DT_RANGE
	//DT_IN
	DT_INCLUDE
	DT_VAR
	DT_RETURN
	DT_SWITCH
	DT_CASE
	DT_BREAK
	DT_FUNC
	DT_GLOBAL
	Operators
	DT_INC        //  ++
	DT_DEC        //  --
	DT_ADD        //  +
	DT_SUB        //  -
	DT_MUL        //  *
	DT_QUO        //  /
	DT_REM        //  %
	DT_ASSIGN     //  =
	DT_ADD_ASSIGN //  +=
	DT_SUB_ASSIGN //  -=
	DT_MUL_ASSIGN //  *=
	DT_QUO_ASSIGN //  =
	DT_REM_ASSIGN //  %=
	DT_LOGIC_AND  //  &&
	DT_LOGIC_OR   //  ||
	DT_EQL        //  ==
	DT_LSS        //  <
	DT_GTR        //  >
	DT_NOT        //  !
	DT_NEQ        //  !=
	DT_LEQ        //  <=
	DT_GEQ        //  >=
	DT_LPAREN     //  (
	DT_RPAREN     //  )
	DT_LBRACK     //  [
	DT_RBRACK     //  ]
	DT_LBRACE     //  {
	DT_RBRACE     //  }
	DT_COMMA      //  ,
	DT_PERIOD     //  .
	DT_SEMICOLON  //  ;
	DT_COLON      //  :
	OperatorsEnd
)

var tokens = [...]string{
	DT_EOF:      "eof",
	DT_ILLEGAL:  "illegal",
	DT_LEFTDEL:  "leftDel",
	DT_RIGHTDEL: "rightDel",

	DT_HTMLTEXT:      "htmlText",
	DT_IDENT:         "identifier",
	DT_INT_LIT:       "int_lit",
	DT_FLOAT_LIT:     "float_lit",
	DT_BOOL_LIT:      "bool_lit",
	DT_STRING_LIT:    "string_lit",
	DT_RAWSTRING_LIT: "rawString_lit",

	DT_INT:    "int",
	DT_FLOAT:  "float",
	DT_BOOL:   "bool",
	DT_STRING: "string",
	DT_ECHO:   "@",
	DT_NULL:   "null",
	DT_IF:     "if",
	DT_ELSE:   "else",
	DT_FOR:    "for",
	DT_RANGE:  "range",
	//DT_IN:      "in",
	DT_INCLUDE: "#",
	DT_VAR:     "var",
	DT_RETURN:  "return",
	DT_SWITCH:  "switch",
	DT_CASE:    "case",
	DT_BREAK:   "break",
	DT_FUNC:    "func",
	DT_GLOBAL:  "global",

	DT_ADD:        "+",
	DT_SUB:        "-",
	DT_MUL:        "*",
	DT_QUO:        "/",
	DT_REM:        "%",
	DT_ASSIGN:     "=",
	DT_ADD_ASSIGN: "+=",
	DT_SUB_ASSIGN: "-=",
	DT_MUL_ASSIGN: "*=",
	DT_QUO_ASSIGN: "/=",
	DT_REM_ASSIGN: "%=",
	DT_LOGIC_AND:  "&&",
	DT_LOGIC_OR:   "||",
	DT_INC:        "++",
	DT_DEC:        "--",
	DT_EQL:        "==",
	DT_LSS:        "<",
	DT_GTR:        ">",
	DT_NOT:        "!",
	DT_NEQ:        "!=",
	DT_LEQ:        "<=",
	DT_GEQ:        ">=",
	DT_LPAREN:     "(",
	DT_RPAREN:     ")",
	DT_LBRACK:     "[",
	DT_RBRACK:     "]",
	DT_LBRACE:     "{",
	DT_RBRACE:     "}",
	DT_COMMA:      ",",
	DT_PERIOD:     ".",
	DT_SEMICOLON:  ";",
	DT_COLON:      ":",
}
var keywords = map[string]token{}
var operators = map[string]token{}

func (t token) String() string {
	return tokens[t]
}
func init() {
	for i := Keywords + 1; i < Operators; i++ {
		keywords[tokens[i]] = i
	}
	for i := Operators + 1; i < OperatorsEnd; i++ {
		operators[tokens[i]] = i
	}
}
