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

type DecoratedType struct {
	Struct bool
	Enum   bool
	Union  bool
	Const  bool
}

func (p *SimpleLineParser) ParseSimpleType() (goTypename string, decorated DecoratedType) {
	primitive := ""
	consters := map[string]*bool{"const": &decorated.Const, "volatile": new(bool)}
	qualifiers := map[string]int{"unsigned": 0, "long": 0, "short": 0, "signed": 0}
	primitives := map[string]bool{"int": false, "char": false, "float": false, "double": false}
	composite := map[string]*bool{"struct": &decorated.Struct, "enum": &decorated.Enum, "union": &decorated.Union}
	// FIXME: on first time, bail out if empty
	for {
		p.SkipBlank()
		id := p.MaybeIdentifier()

		if _, ok := consters[id]; ok {
			*consters[id] = true
			continue
		}
		if _, ok := qualifiers[id]; ok {
			qualifiers[id] = qualifiers[id] + 1
			primitive = "int"
			continue
		}
		if _, ok := primitives[id]; ok {
			primitive = id
			continue
		}
		if _, ok := composite[id]; ok {
			*composite[id] = true
			continue
		}

		if decorated.Struct || decorated.Enum || decorated.Union {
			return upcase(id), decorated
		}
		if primitive == "" {
			return upcase(id), decorated
		}
		p.Ungets([]byte(id))
		return translatePrimitive(primitive, qualifiers), decorated
	}
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

	//TODO: `short FOO;` vs `short int FOO;`

	p.Lexer = Lexer{R: bufio.NewReader(strings.NewReader(s))}
	p.SkipBlank()

	d := Decl{}
	d.Typedef = p.Maybe("typedef")
	p.SkipBlank()

	typenameGo, decor := p.ParseSimpleType()
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

	if bcurly1 || ecurly1 {
		return fmt.Errorf("struct/enum/union definitions not supported")
	}

	ptr := p.Maybe("*")
	p.SkipBlank()

	ident := p.Identifier()
	p.SkipBlank()
	p.Expect(";")

	if p.CurlyDepth == 0 && d.Typedef {
		p.W.WriteString("type ")
	} else if p.CurlyDepth == 0 {
		return fmt.Errorf("variables not supported")
	}

	if decor.Enum || decor.Union || decor.Const {
		return fmt.Errorf("enum/union/const not supported")
	}

	writeindent(p.W, p.CurlyDepth)

	if ident != "" {
		p.W.WriteString(upcase(ident) + " ")
	} else {
		return fmt.Errorf("anonymous structs/enums not supported")
	}

	if ptr {
		p.W.WriteString("*")
	}

	if typenameGo != "" {
		p.W.WriteString(typenameGo)
	} else {
		return fmt.Errorf("untyped declarations not supported")
	}

	p.W.WriteString("\t// " + s)
	p.W.WriteString("\n")
	return nil
}

func upcase(s string) string {
	return strings.Title(s)
}

func setif(condition bool, dst *bool) bool {
	if condition {
		*dst = true
	}
	return condition
}

func translatePrimitive(primitive string, q map[string]int) string {
	//FIXME: verify if those are ok
	switch primitive {
	case "int":
		switch {
		//FIXME: handle long long (?)
		case q["long"] > 1:
			panic("long long not supported")
		case q["long"] > 0 && q["unsigned"] > 0:
			return "uint32"
		case q["long"] > 0:
			return "int32"
		case q["unsigned"] > 0: // "short" or not
			return "uint16"
		default:
			return "int16"
		}
	case "char":
		switch {
		case q["unsigned"] > 0:
			return "byte"
		default:
			return "int8"
		}
	case "float":
		return "float32"
	case "double": //FIXME: what about long double? panic?
		return "float64"
	default:
		panic("unknown primitive type")
	}
}
