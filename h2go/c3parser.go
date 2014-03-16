package h2go

import (
	"bufio"
	"fmt"
	"strings"
)

type SimpleLineParser struct {
	Lexer
	Printer

	compositeStruct
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

type compositeStruct struct {
	decl       Decl
	typenameGo string
	decor      DecoratedType
}

func (p *SimpleLineParser) ParseSimpleType() (goTypename string, decorated DecoratedType) {
	primitive := ""
	consters := map[string]*bool{"const": &decorated.Const, "volatile": new(bool)}
	qualifiers := map[string]int{"unsigned": 0, "long": 0, "short": 0, "signed": 0}
	primitives := map[string]bool{"int": false, "char": false, "float": false, "double": false, "wchar_t": false}
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

	//TODO: enums
	//TODO: skip unions
	//TODO: const

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

	pluscurly := strings.Count(s, "{")
	minuscurly := strings.Count(s, "}")
	p.CurlyDepth += pluscurly - minuscurly
	if pluscurly > 0 && minuscurly > 0 {
		return fmt.Errorf("more than one { or } per line not supported")
	}

	if bcurly1 && decor.Struct { // && !ecurly1, because input from Simplify
		if p.CurlyDepth > 1 {
			return fmt.Errorf("nested structs not supported; be careful with alignment")
		}
		if typenameGo == "" || decor.Const {
			p.compositeStruct = compositeStruct{}
			return fmt.Errorf("unnamed/const struct definitions not supported")
		}
		p.decl = d
		p.typenameGo = typenameGo
		p.decor = decor

		p.W.WriteString("type " + typenameGo + " struct {\t// " + s + "\n")
		return nil
	} else if ecurly1 && p.CurlyDepth == 0 && p.decor.Struct {
		p.W.WriteString("}\n")
		typenameGo = p.typenameGo
		d = p.decl
	} else if bcurly1 || ecurly1 {
		return fmt.Errorf("enum/union/nested-struct definitions not supported")
	}

	for {
		ident, ornaments := p.ParseOrnamentedIdent()

		fin := p.Maybe(";")
		if !p.Maybe(",") && !fin {
			panic("expected , or ;")
		}
		p.SkipBlank()

		err = p.emit(d, decor, ident, ornaments, typenameGo, s)
		if err != nil {
			return err
		}
		if fin {
			break
		}
	}
	return nil
}

func (p *SimpleLineParser) ParseOrnamentedIdent() (ident, ornaments string) {
	ptr := p.Maybe("*")
	p.SkipBlank()

	ident = p.Identifier()
	p.SkipBlank()

	for {
		if !p.Maybe("[") {
			break
		}
		p.SkipBlank()

		n := p.ExpectNumber()
		p.SkipBlank()

		p.Expect("]")
		p.SkipBlank()

		ornaments += "[" + n + "]"
	}
	if ptr {
		ornaments += "*"
	}
	return
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

/*
want as described in MSDN:
DWORD = unsigned long = uint32
BOOL = int = int32
BYTE = unsigned char = byte
CHAR = char = int8

*/

func translatePrimitive(primitive string, q map[string]int) string {
	//FIXME: verify if those are ok
	switch primitive {
	case "int":
		switch {
		case q["short"] > 0 && q["unsigned"] > 0:
			return "uint16"
		case q["short"] > 0:
			return "int16"

		//FIXME: handle long long (?)
		case q["long"] > 1:
			panic("long long not supported")

		case q["long"] > 0 && q["unsigned"] > 0:
			return "uint32"
		case q["long"] > 0:
			return "int32"

		case q["unsigned"] > 0:
			return "uint32"
		default:
			return "int32"
		}
	case "char":
		switch {
		case q["unsigned"] > 0:
			return "byte"
		default:
			return "int8"
		}
	case "wchar_t":
		//switch {
		//case q["unsigned"] > 0:
		return "uint16"
		//default:
		//	return "int16"
		//}
	case "float":
		return "float32"
	case "double": //FIXME: what about long double? panic?
		return "float64"
	default:
		panic("unknown primitive type")
	}
}
