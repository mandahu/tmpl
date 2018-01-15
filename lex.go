package template

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

type lex struct {
	input []byte
	left  []byte
	right []byte
	start int
	pos   int
	width int
	line  int
	state fn
	item  chan item
}
type item struct {
	t token
	l int
	p int
	v string
}
type fn func(l *lex) fn

const operator = "+-*/%&|!<=>(){}[]:,.;"

func (l *lex) next() rune {
	if l.pos >= len(l.input) {
		return -1
	}
	r, w := utf8.DecodeRune(l.input[l.pos:])
	l.pos += w
	l.width = w
	if r == '\n' {
		l.line++
	}
	return r
}
func (l *lex) backup() {
	l.pos -= l.width
	if l.input[l.pos] == '\n' {
		l.line--
	}
}
func (l *lex) peek() rune {
	r := l.next()
	l.backup()
	return r
}
func (l *lex) ignore() {
	l.start = l.pos
}
func (l *lex) emit(t token) {
	l.item <- item{t, l.line, l.start, string(l.input[l.start:l.pos])}
	if t == DT_HTMLTEXT {
		l.line += bytes.Count(l.input[l.start:l.pos], []byte("\n"))
	}
	l.ignore()
}
func (l *lex) errorf(format string, args ...interface{}) fn {
	val := fmt.Sprintf(format, args...)
	l.item <- item{DT_ILLEGAL, l.line, l.start, val}
	return nil
}
func lexer(input []byte, left, rigth string) (l *lex) {
	l = &lex{
		left:  []byte(left),
		right: []byte(rigth),
		input: input,
		line:  1,
		item:  make(chan item, 100),
	}
	//tokens[DT_LEFTDEL] = string(l.left)
	//tokens[DT_RIGHTDEL] = string(l.right)
	go l.run()
	return l
}
func (l *lex) run() {
	for l.state = scanText; l.state != nil; {
		l.state = l.state(l)
	}
	close(l.item)
}
func scanText(l *lex) fn {
	if x := bytes.Index(l.input[l.pos:], l.left); x >= 0 {
		l.pos += x
		if x > 0 {
			l.emit(DT_HTMLTEXT)
		}
		return scanLeftDel
	}
	l.pos = len(l.input)
	return scanHTML
}
func scanLeftDel(l *lex) fn {
	l.pos += len(l.left)
	l.emit(DT_LEFTDEL)
	return scanScript
}
func scanHTML(l *lex) fn {
	if l.pos > l.start {
		l.emit(DT_HTMLTEXT)
		return scanText
	}
	l.emit(DT_EOF)
	return nil
}
func scanScript(l *lex) fn {
	if bytes.HasPrefix(l.input[l.pos:], l.right) {
		return scanRightDel
	}
	switch r := l.next(); {
	case isNewLine(r):
		l.ignore()
		return scanScript
	case isSpace(r):
		return scanSpace
	case r == '/' && l.peek() == '/':
		return scanComment
	case isAlpha(r):
		return scanIdentOrKeywords
	case r == '@':
		l.emit(DT_ECHO)
		return scanScript
	case r == '#':
		l.emit(DT_INCLUDE)
		return scanScript
	case strings.IndexRune(operator, r) >= 0:
		return scanOp
	case unicode.IsDigit(r):
		return scanNumber
	case r == '"':
		return scanString
	case r == '`':
		return scanRawString
	default:
		return l.errorf("syntax error:%s", string(r))
	}
}
func scanRightDel(l *lex) fn {
	l.pos += len(l.right)
	l.emit(DT_RIGHTDEL)
	return scanText
}
func isNewLine(r rune) bool {
	return r == '\n' || r == '\r'
}
func isAlphaNumeric(r rune) bool {
	return isAlpha(r) || (r >= '0' && r <= '9')
}
func isAlpha(r rune) bool {
	return (r >= 'a' && r <= 'z') || r == '_' || (r >= 'A' && r <= 'Z')
}
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}
func scanSpace(l *lex) fn {
	for isSpace(l.peek()) {
		l.next()
	}
	l.ignore()
	return scanScript
}
func scanComment(l *lex) fn {
	if bytes.HasPrefix(l.input[l.pos:], l.right) {
		return scanRightDel
	}
	for {
		if r := l.peek(); isNewLine(r) {
			return scanScript
		}
		l.next()
		return scanComment
	}
}
func scanIdentOrKeywords(l *lex) fn {
loop:
	words := string(l.input[l.start:l.pos])
	if t, ok := keywords[words]; ok {
		if isSpace(l.peek()) {
			l.emit(t)
			return scanScript
		} else {
			if isAlphaNumeric(l.peek()) {
				l.next()
				goto loop
			} else {
				l.emit(t)
				return scanScript
			}
		}
	} else {
		if isAlphaNumeric(l.peek()) {
			l.next()
			goto loop
		} else {
			if words == "true" || words == "false" {
				l.emit(DT_BOOL_LIT)
			} else {
				l.emit(DT_IDENT)
			}
			return scanScript
		}
	}
}
func scanOp(l *lex) fn {
	if k, ok := operators[string(l.input[l.start:l.pos])]; ok {
		l.next()
		if k1, ok := operators[string(l.input[l.start:l.pos])]; ok {
			l.emit(k1)
		} else {
			l.backup()
			l.emit(k)
		}
	} else {
		l.next()
		if k1, ok := operators[string(l.input[l.start:l.pos])]; ok {
			l.emit(k1)
		} else {
			l.backup()
			l.errorf("illegal operators: %s", string(l.input[l.start:l.pos]))
		}
	}
	return scanScript
}
func scanNumber(l *lex) fn {
loop:
	for {
		switch r := l.next(); {
		case unicode.IsDigit(r) || r == '.':
			goto loop
		default:
			l.backup()
			val := l.input[l.start:l.pos]
			if ok, _ := regexp.Match(`^[\d]+([\.]?\d+)?$`, val); ok {
				if ok, _ := regexp.Match(`^\d+$`, val); ok {
					l.emit(DT_INT_LIT)
				} else {
					l.emit(DT_FLOAT_LIT)
				}
				return scanScript
			}
			return l.errorf("illegal numeric literal %s", string(l.input)[l.start:l.pos])
		}
	}
}
func scanString(l *lex) fn {
loop:
	for {
		switch r := l.next(); {
		case r != '"':
			goto loop
		default:
			l.emit(DT_STRING_LIT)
			return scanScript
		}
	}
}
func scanRawString(l *lex) fn {
loop:
	for {
		switch r := l.next(); {
		case r != '`':
			goto loop
		default:
			l.emit(DT_RAWSTRING_LIT)
			return scanScript
		}
	}
}
