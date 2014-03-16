package h2go

import (
	"bufio"
	"fmt"
	"strings"
)

type SimpleLineParser struct {
	Lexer
	W          *bufio.Writer
	CurlyDepth int
	//	Typedef bool
}

type Decl struct {
	Typedef bool
}

type SimpleType struct {
	Struct bool
	Enum   bool
	Const  bool
}

func (p *SimpleLineParser) ParseLine(s string) (err error) {
	defer p.W.Flush()
	defer func() {
		e := recover()
		if err == nil && e != nil {
			err = fmt.Errorf("%d: %s", p.Col, e)
		}
	}()

	//TODO: commas
	//TODO: structs, enums
	//TODO: skip unions
	//TODO: typedefs
	//TODO: const

	p.Lexer = Lexer{R: bufio.NewReader(strings.NewReader(s))}
	p.SkipBlank()

	d := Decl{}
	d.Typedef = p.Maybe("typedef")
	p.SkipBlank()

	t := SimpleType{}
	t.Const = p.Maybe("const")
	p.SkipBlank()
	t.Struct = p.Maybe("struct")
	p.SkipBlank()
	t.Enum = p.Maybe("enum")
	p.SkipBlank()

	typename1 := p.MaybeIdentifier()
	p.SkipBlank()
	bcurly1 := p.Maybe("{")
	p.SkipBlank()
	ecurly1 := p.Maybe("}")
	p.SkipBlank()

	if bcurly1 {
		p.CurlyDepth++
	}
	if ecurly1 {
		p.CurlyDepth--
	}

	ident2 := p.Identifier()
	p.SkipBlank()
	p.Expect(";")

	if bcurly1 || ecurly1 {
		return fmt.Errorf("stuct/enum/union definitions not supported")
	}

	if p.CurlyDepth == 0 && d.Typedef {
		p.W.WriteString("type ")
	} else if p.CurlyDepth == 0 {
		return fmt.Errorf("variables not supported")
	}

	if ident2 != "" {
		p.W.WriteString(upcase(ident2) + " ")
	} else {
		return fmt.Errorf("anonymous structs/enums not supported")
	}

	if typename1 != "" {
		p.W.WriteString(upcase(typename1))
	} else {
		return fmt.Errorf("untyped declarations not supported")
	}

	p.W.WriteString("\t// " + s)
	p.W.WriteString("\n")
	return nil

	//p.CurlyDepth += strings.Count(s, "{") - strings.Count(s, "}")
	//return fmt.Errorf("unrecognized declaration")
}

func upcase(s string) string {
	return strings.Title(s)
}
