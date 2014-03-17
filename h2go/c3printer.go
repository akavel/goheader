package h2go

import (
	"bufio"
	"fmt"
	"os"
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

	if decor.Enum || decor.Union {
		return fmt.Errorf("enum/union not supported")
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
	if typedef && p.Flatten[ident] == "" {
		p.Flatten[ident] = full
	}

	p.W.WriteString("\t// " + original)
	p.W.WriteString("\n")
	return nil
}

func (p *Printer) Preload(gofile string) error {
	if p.Flatten == nil {
		p.Flatten = make(map[string]string)
	}

	f, err := os.Open(gofile)
	if err != nil {
		return err
	}
	defer f.Close()

	prefix1 := "type "

	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Text()
		if strings.Contains(line, "{") {
			continue // TODO: structs not supported
		}
		if !strings.HasPrefix(line, prefix1) {
			continue
		}
		words := strings.SplitN(line[len(prefix1):], " ", 2)
		if p.Flatten[words[0]] == "" {
			p.Flatten[words[0]] = words[1]
		}
	}
	return s.Err()
}
