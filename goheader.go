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
	fmt.Fprintf(os.Stderr, "Usage: goheader [TYPE_NAME] < HEADER.h >> FILE.go\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func run() error {
	flag.Usage = usage
	flag.Parse()

	var filter string
	switch len(flag.Args()) {
	case 0:
	case 1:
		filter = flag.Args()[0]
	default:
		usage()
	}

	parser := h2go.SimpleLineParser{}
	p := pipe.Line(
		pipe.Read(os.Stdin),
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
		}),
		BufferedFunc(h2go.Simplify),
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
		KeepTypename(filter),
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

func KeepTypename(t string) pipe.Pipe {
	in := false
	prefix := []byte("type " + t + " ")
	prefix2 := []byte("struct {")
	return pipe.Filter(func(line []byte) bool {
		if t == "" {
			return true
		}

		if bytes.HasPrefix(line, prefix) {
			if bytes.HasPrefix(line[len(prefix):], prefix2) {
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
