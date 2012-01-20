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
	"bytes"
	"fmt"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

// Values by default used in 'gofmt'.
const (
	PARSER_MODE  = parser.ParseComments
	PRINTER_MODE = printer.TabIndent | printer.UseSpaces
	TAB_WIDTH    = 8
)

// Formats the Go source code.
func (self *translate) format() error {
	fset := token.NewFileSet()

	// The output is an abstract syntax tree (AST) representing the Go source.
	ast, err := parser.ParseFile(fset, self.filename, self.raw.Bytes(), PARSER_MODE)
	if err != nil {
		return err
	}

	// Print an AST node to output.
	_, err = (&printer.Config{PRINTER_MODE, TAB_WIDTH}).Fprint(
		self.fmt, fset, ast)
	if err != nil {
		return err
	}

	return nil
}

func (self *translate) write() error {
	output := new(bytes.Buffer)

	if !*debug {
		output = self.fmt
	} else {
		output = self.raw
	}

	if *write {
		/*filename := self.filename

		switch *system {
		case "linux":
			dirBase := "/usr/include/"

			if strings.HasPrefix(filename, dirBase) {
				filename = strings.SplitN(filename, dirBase, 2)[1]
				filename = strings.Replace(filename, "/", "_", -1)
			} else {
				filename = path.Base(filename)
			}
		}
		filename = strings.SplitN(filename, ".h", 2)[0]
		*/
		filename := strings.SplitN(path.Base(self.filename), ".h", 2)[0]
		filename = fmt.Sprintf("h-%s_%s.go", filename, *system)

		outFile, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer outFile.Close()

		if err = ioutil.WriteFile(filename, output.Bytes(), 0); err != nil {
			return err
		}
	} else {
		if _, err := os.Stdout.Write(output.Bytes()); err != nil {
			return err
		}
	}

	return nil
}
