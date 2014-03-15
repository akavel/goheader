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

package h2go

import (
	"go/parser"
	"go/printer"
	"go/token"
)

// Values by default used in 'gofmt'.
const (
	PARSER_MODE  = parser.ParseComments
	PRINTER_MODE = printer.TabIndent | printer.UseSpaces
	TAB_WIDTH    = 8
)

// Formats the Go source code.
func (self *Translate) Format() error {
	fset := token.NewFileSet()

	// The output is an abstract syntax tree (AST) representing the Go source.
	ast, err := parser.ParseFile(fset, self.Filename, self.Raw.Bytes(), PARSER_MODE)
	if err != nil {
		return err
	}

	// Print an AST node to output.
	err = (&printer.Config{Mode: PRINTER_MODE, Tabwidth: TAB_WIDTH}).Fprint(self.Fmt, fset, ast)
	if err != nil {
		return err
	}

	return nil
}
