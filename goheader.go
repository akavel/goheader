package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"

	"labix.org/v2/pipe"

	"github.com/akavel/goheader/h2go"
)

var FAIL = "//!!! "

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: goheader [-p GO_PACKAGE] [-l PRELOAD.go] [TYPE_NAME]... < HEADER.h >> FILE.go\n")
	flag.PrintDefaults()
	os.Exit(2)
}

var (
	pkg     = flag.String("p", "", "If used, a `package GO_PACKAGE` line is printed before anything else.")
	preload = flag.String("l", "", "If used, preloads type definitions from PRELOAD.go")
)

func run() error {
	flag.Usage = usage
	flag.Parse()

	if *pkg != "" {
		fmt.Println("package", *pkg)
	}

	parser := h2go.SimpleLineParser{}

	if *preload != "" {
		err := parser.Preload(*preload)
		if err != nil {
			return err
		}
	}

	p := pipe.Line(
		pipe.Read(os.Stdin),

		// Strip lines starting with '# 2342' -- i.e. compiler directive for marking line number
		pipe.Filter(func(line []byte) bool {
			line = bytes.TrimSpace(line)
			if len(line) == 0 || line[0] != '#' {
				return true
			}
			line = bytes.TrimSpace(line[1:])
			return len(line) == 0 || line[0] < '0' || line[0] > '9'
		}),

		BufferedFunc(h2go.Simplify),

		ReplaceAll([][2]string{
			{`__extension__`, ` `},
			{`__attribute__ ((`, `__attribute__((`},
			{`__attribute__((__cdecl__))`, ` `},
			{`__attribute__((__stdcall__))`, ` `},
			{`__attribute__((__nothrow__))`, ` `},
			{`__attribute__((__pure__))`, ` `},
			{`__attribute__((__const__))`, ` `},
			{`__attribute__((__deprecated__))`, ` `},
			{`__attribute__((__malloc__))`, ` `},
			{`__attribute__((__malloc__))`, ` `},
			{`__attribute__((noreturn))`, ` `},
			{`__attribute__((__noreturn__))`, ` `},
			{`__attribute__((__dllimport__))`, ` `},
			{`__attribute__((packed))`, ` /* PACKED!!! */ `},
			{`# pragma `, `#pragma `},
			{`#pragma pack (`, `#pragma pack(`},
			{`#pragma pack( `, `#pragma pack(`},
		}),

		BufferedFunc(h2go.Simplify),

		// Add warnings in case of packing pragmas
		WarnPackingPragmas(),

		// Main parsing & translation
		pipe.Replace(func(line []byte) []byte {
			s := strings.Trim(string(line), "\n\r\t ")
			if s == "" {
				return []byte{}
			}
			out := bytes.NewBuffer(nil)
			parser.W = bufio.NewWriter(out)
			err := parser.ParseLine(s)
			if err != nil {
				return []byte(fmt.Sprintf("%s %s // %s\n", FAIL, s, err))
			}
			return out.Bytes()
		}),

		// Optional filtering to extract just one type
		KeepTypenames(flag.Args()),

		pipe.Write(os.Stdout),
	)

	err := pipe.Run(p)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func BufferedFunc(f func(r *bufio.Reader, w *bufio.Writer) error) pipe.Pipe {
	return pipe.TaskFunc(func(s *pipe.State) error {
		out := bufio.NewWriter(s.Stdout)
		defer out.Flush()
		return f(bufio.NewReader(s.Stdin), out)
	})
}

func ReplaceAll(p [][2]string) pipe.Pipe {
	return pipe.Replace(func(line []byte) []byte {
		for _, r := range p {
			line = bytes.Replace(line, []byte(r[0]), []byte(r[1]), -1)
		}
		return line
	})
}

func KeepTypenames(t []string) pipe.Pipe {
	m := map[string]struct{}{}
	for _, s := range t {
		m[s] = struct{}{}
	}

	in := false
	prefix1 := []byte("type ") // + t + " ")
	prefix2 := []byte("struct {")
	return pipe.Filter(func(line []byte) bool {
		if len(t) == 0 {
			return true
		}

		if bytes.HasPrefix(line, prefix1) {
			words := bytes.SplitN(line[len(prefix1):], []byte(" "), 2)
			if _, ok := m[string(words[0])]; !ok {
				return in // could 'return false', but here we will get more lines in some bad cases, helping to quickly debug
			}
			if bytes.HasPrefix(words[1], prefix2) {
				in = true
			}
			return true
		}
		if in && len(line) > 0 {
			if line[0] == '}' {
				in = false
			}
			return true
		}
		return false
	})
}

func WarnPackingPragmas() pipe.Pipe {
	stack := []string{}
	push := []byte(`#pragma pack(push`)
	pop := []byte(`#pragma pack(pop`)
	return pipe.Replace(func(line []byte) []byte {
		buf := bytes.TrimSpace(line)
		switch {
		case bytes.HasPrefix(buf, push):
			stack = append(stack, " //WARNING: "+string(buf)+"\n")
		case bytes.HasPrefix(buf, pop):
			stack = stack[:len(stack)-1]
		case len(stack) > 0 && len(buf) > 0:
			return append(buf, []byte(stack[len(stack)-1])...)
		}
		return line
	})
}
