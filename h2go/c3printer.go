package h2go

import (
	"bufio"
	"fmt"
	"strings"
)

type Printer struct {
	W          *bufio.Writer
	CurlyDepth int
	Flatten    map[string]string
}

func (p *Printer) emit(d Decl, decor DecoratedType, ident, ornaments, typenameGo, original string) error {
	if p.Flatten == nil {
		p.Flatten = make(map[string]string)
	}

	typedef := p.CurlyDepth == 0 && d.Typedef
	if typedef {
		p.W.WriteString("type ")
	} else if p.CurlyDepth == 0 {
		return fmt.Errorf("variables not supported")
	}

	if decor.Enum || decor.Union || decor.Const {
		return fmt.Errorf("enum/union/const not supported")
	}

	writeindent(p.W, p.CurlyDepth)

	if ident == "" {
		return fmt.Errorf("anonymous structs/enums not supported")
	}
	p.W.WriteString(upcase(ident) + " ")

	if typenameGo == "" {
		return fmt.Errorf("untyped declarations not supported")
	}
	if v := p.Flatten[typenameGo]; v != "" {
		typenameGo = v
	}

	full := ornaments + typenameGo
	if strings.HasSuffix(full, "*Void") {
		full = strings.TrimSuffix(full, "*Void") + "uintptr"
	}
	p.W.WriteString(full)
	if typedef {
		p.Flatten[ident] = full
	}

	p.W.WriteString("\t// " + original)
	p.W.WriteString("\n")
	return nil
}
