// Copyright 2010  The "goheader" Authors
//
// Use of this source code is governed by the Simplified BSD License
// that can be found in the LICENSE file.
//
// This software is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied. See the License
// for more details.

package main

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/printer"
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
func (self *translate) format() os.Error {
	// The output is an abstract syntax tree (AST) representing the Go source.
	ast, err := parser.ParseFile(self.filename, self.raw.Bytes(), PARSER_MODE)
	if err != nil {
		return err
	}

	// Print an AST node to output.
	_, err = (&printer.Config{PRINTER_MODE, TAB_WIDTH, nil}).Fprint(self.fmt, ast)
	if err != nil {
		return err
	}

	self.fmt.WriteByte('\n')

	return nil
}

func (self *translate) write() os.Error {
	output := new(bytes.Buffer)

	if !*debug {
		output = self.fmt
	} else {
		output = self.raw
	}

	if *write {
		filename := strings.Split(path.Base(self.filename), ".h", 2)[0]
		filename = fmt.Sprintf("%s_%s.go", filename, *system)

		outFile, err := os.Open(filename, os.O_CREATE|os.O_WRONLY, 0644)
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
