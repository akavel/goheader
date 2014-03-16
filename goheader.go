// Copyright 2010  The "GoHeader" Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	fmt.Fprintf(os.Stderr, "Usage: goheader < HEADER.h > FILE.go\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func run() error {
	flag.Usage = usage
	flag.Parse()

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
			s := strings.TrimRight(string(line), "\n\r")
			out := bytes.NewBuffer(nil)
			parser.W = bufio.NewWriter(out)
			err := parser.ParseLine(s)
			if err != nil {
				return []byte(fmt.Sprintf("%s %s // %s\n", FAIL, s, err))
			}
			return out.Bytes()
		}),
		//pipe.TaskFunc(func(s *pipe.State) error {
		//	out := bufio.NewWriter(s.Stdout)
		//	defer out.Flush()
		//	return h2go.C(bufio.NewReader(s.Stdin), out)
		//}),
		pipe.Write(os.Stdout),
	)

	_ = h2go.C

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
