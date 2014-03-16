package h2go

import (
	"bufio"
	"fmt"
	"io"
)

// NOTE: methods called on Lexer may panic!
// TODO(akavel): monitor position in source file (line & column)
type Lexer struct {
	R        *bufio.Reader
	Row, Col int
	ungets   []byte
}

func (l *Lexer) Getc() (c byte, eof bool) {
	var err error
	if len(l.ungets) > 0 {
		i := len(l.ungets) - 1
		c, l.ungets = l.ungets[i], l.ungets[:i]
		//FIXME: handle Row, Col properly
	} else {
		c, err = l.R.ReadByte()
	}
	if err == io.EOF {
		return 0, true
	}
	if err != nil {
		panic(err)
	}
	if c == '\n' {
		l.Row++
		l.Col = 0
	} else {
		l.Col++
	}
	return c, false
}

func (l *Lexer) Ungetc(c byte) {
	l.ungets = append(l.ungets, c)
	if c == '\n' {
		l.Row--
		l.Col = -1
	} else {
		l.Col--
	}
}

func (l *Lexer) Ungets(b []byte) {
	for i := len(b) - 1; i >= 0; i-- {
		l.Ungetc(b[i])
	}
}

func (l *Lexer) SkipBlank() {
	for {
		c, eof := l.Getc()
		if eof {
			return
		}
		if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
			l.Ungetc(c)
			break
		}
	}
}

func (l *Lexer) Expect(s string) {
	for _, exp := range []byte(s) {
		got, _ := l.Getc()
		if got != exp {
			panic(fmt.Errorf("expected '%c', got '%c' or EOF", exp, got))
		}
	}
}

func (l *Lexer) Maybe(s string) bool {
	for i, exp := range []byte(s) {
		got, eof := l.Getc()
		if got != exp {
			if !eof {
				l.Ungetc(got)
			}
			l.Ungets([]byte(s)[:i])
			return false
		}
	}
	return true
}

func startIdent(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'
}

func (l *Lexer) MaybeIdentifier() string {
	c, eof := l.Getc()
	if !startIdent(c) {
		if !eof {
			l.Ungetc(c)
		}
		return ""
	}
	buf := []byte{c}
	for {
		c, eof = l.Getc()
		if !startIdent(c) && !(c >= '0' && c <= '9') {
			if !eof {
				l.Ungetc(c)
			}
			break
		}
		buf = append(buf, c)
	}
	return string(buf)
}

func (l *Lexer) Identifier() string {
	s := l.MaybeIdentifier()
	if s == "" {
		panic("expected identifier")
	}
	return s
}
