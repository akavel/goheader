package h2go

import (
	"bufio"
	"io"
)

func writeindent(w *bufio.Writer, n int) {
	for i := 0; i < n; i++ {
		w.WriteByte('\t')
	}
}

func Simplify(r *bufio.Reader, w *bufio.Writer) error {
	waswhite := false
	indent := 0
	for {
		c, err := r.ReadByte()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		switch c {
		case '\n', '\r', '\t', ' ':
			if !waswhite {
				w.WriteByte(' ')
			}
			waswhite = true
		case ';':
			w.WriteString(";\n")
			writeindent(w, indent)
			waswhite = true
		case '{':
			w.WriteString("{\n")
			indent++
			writeindent(w, indent)
			waswhite = true
		case '}':
			w.WriteString("\n")
			indent--
			writeindent(w, indent)
			fallthrough
		default:
			w.WriteByte(c)
			waswhite = false
		}
	}
}
