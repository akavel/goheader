package h2go

import (
	"bufio"
	"fmt"
)

type Printer struct {
	W          *bufio.Writer
	CurlyDepth int
}

func (p *Printer) emit(d Decl, decor DecoratedType, ident, ornaments, typenameGo, original string) error {
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

	p.W.WriteString(ornaments)

	if typenameGo != "" {
		p.W.WriteString(typenameGo)
	} else {
		return fmt.Errorf("untyped declarations not supported")
	}

	p.W.WriteString("\t// " + original)
	p.W.WriteString("\n")
	return nil
}
