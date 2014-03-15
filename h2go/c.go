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
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Represents the header file to translate. It has the Go output in both raw and
// formatted code.
type Translate struct {
	Filename string        // The header file
	Raw, Fmt *bytes.Buffer // The Go output
}

const COMMENT_LINE = "//!!! "
const NoLastEnumValue = -1000

var GoBase = `// {cmd}
// MACHINE GENERATED; DO NOT EDIT
// ===

package {pkg}

`

// Regular expressions for C header.
var (
	reSkip = regexp.MustCompile(`^(\n|//)`) // Empty lines and comments.

	reTypedef     = regexp.MustCompile(`^(typedef)[ \t]+(.+)[ \t]+(.+)[;](.+)?`)
	reTypedefOnly = regexp.MustCompile(`^(typedef)[ \t]+`)

	reStruct          = regexp.MustCompile(`^(struct)[ \t]+(.+)[ \t]*{`)
	reStructField     = regexp.MustCompile(`^(.+)[ \t]+(.+)[;](.+)?`)
	reStructFieldName = regexp.MustCompile(`^([^_]*_)?(.+)`)

	reEnum      = regexp.MustCompile(`^(enum)[ \t]+(.+)[ \t]*{`)
	reEnumValue = regexp.MustCompile(`^(.+)[ \t]*=[ \t]*([^,]+)`)
	reEnumIota  = regexp.MustCompile(`^([^,]+)[,]?\n`)
	reEnumEnd   = regexp.MustCompile(`^};`)

	reDefine      = regexp.MustCompile(`^#[ \t]*(define|DEFINE)[ \t]+([^ \t]+)[ \t]+(.+)`)
	reDefineOnly  = regexp.MustCompile(`^#[ \t]*(define|DEFINE)[ \t]+`)
	reDefineMacro = regexp.MustCompile(`^.*(\(.*\))`)

	reSingleComment         = regexp.MustCompile(`^(.+)?/\*[ \t]*(.+)[ \t]*\*/`)
	reStartMultipleComment  = regexp.MustCompile(`^/\*(.+)?`)
	reMiddleMultipleComment = regexp.MustCompile(`^[ \t*]*(.+)`)
	reEndMultipleComment    = regexp.MustCompile(`^(.+)?\*/`)
)

// Translates C type declaration into Go type declaration.
//
// NOTE: the regular expression for single comments (reSingleComment) returns
// spaces before of "*/".
// The issue is that Go's regexp lib. doesn't support non greedy matches.
func (self *Translate) C(file *os.File) error {
	var isMultipleComment, isTypeBlock, isConstBlock, isStruct, isEnum bool
	lastEnumValue := -1
	extraTypedef := make([]string, 0, 0) // Types defined in the header file.

	self.Raw.WriteString(GoBase)

	// === File to read
	fileBuf := bufio.NewReader(file)

	for {
		line, err := fileBuf.ReadString('\n')
		if err == io.EOF {
			break
		}
		line = strings.TrimSpace(line) + "\n"
		isSingleComment := false

		// === Translate comment of single line.
		if !isMultipleComment {
			if sub := reSingleComment.FindStringSubmatch(line); sub != nil {
				isSingleComment = true
				line = "// " + strings.TrimRight(sub[2], " ") + "\n"

				if sub[1] != "" {
					line = sub[1] + line
				}
			}
		}
		if !isSingleComment && !isMultipleComment && strings.HasPrefix(line, "/*") {
			isMultipleComment = true
		}

		// === Translate comments of multiple line.
		if isMultipleComment {
			if sub := reStartMultipleComment.FindStringSubmatch(line); sub != nil {
				if sub[1] != "\n" {
					self.Raw.WriteString("// " + sub[1])
				}
				continue
			}
			if sub := reEndMultipleComment.FindStringSubmatch(line); sub != nil {
				if sub[1] != "" {
					self.Raw.WriteString("// " + sub[1] + "\n")
				}
				isMultipleComment = false
				continue
			}
			if sub := reMiddleMultipleComment.FindStringSubmatch(line); sub != nil {
				if sub[1] != "\n" {
					self.Raw.WriteString("// " + sub[1])
				} else {
					self.Raw.WriteString("//" + sub[1])
				}
				continue
			}
		}

		// === Translate type definitions.
		if sub := reTypedef.FindStringSubmatch(line); sub != nil {
			// Add the new type.
			extraTypedef = append(extraTypedef, sub[3])

			gotype, ok := ctypeTogo(sub[2], extraTypedef)
			line = fmt.Sprintf("%s %s", sub[3], gotype)

			if sub[4] != "\n" {
				line += sub[4]
			} else {
				line += "\n"
			}

			if !isTypeBlock {
				// Get characters of next line.
				startNextLine, err := fileBuf.Peek(10)
				if err != nil {
					return err
				}

				// In single line.
				if !reTypedefOnly.Match(startNextLine) {
					line = "type " + line
				} else {
					isTypeBlock = true
					line = "type (\n" + line
				}
			}

			if !ok {
				line = COMMENT_LINE + line
			}

			self.Raw.WriteString(line)
			continue
		}

		if isTypeBlock && line == "\n" {
			self.Raw.WriteString(")\n\n")
			isTypeBlock = false
			continue
		}

		// === Translate 'define' to 'const'.
		if sub := reDefine.FindStringSubmatch(line); sub != nil {
			line = fmt.Sprintf("%s = %s", sub[2], sub[3])

			if !isConstBlock {
				// Get characters of next line.
				startNextLine, err := fileBuf.Peek(10)
				if err != nil {
					return err
				}

				// Constant in single line.
				if !reDefineOnly.Match(startNextLine) {
					line = "const " + line
				} else {
					isConstBlock = true
					line = "const (\n" + line
				}
			}

			// Removes comment (if any) to ckeck if it is a macro.
			lastField := strings.Split(sub[3], "//")[0]
			if reDefineMacro.MatchString(lastField) {
				line = COMMENT_LINE + line
			}

			self.Raw.WriteString(line)
			continue
		}

		if isConstBlock && line == "\n" {
			self.Raw.WriteString(")\n\n")
			isConstBlock = false
			continue
		}

		// === Translate enums
		if !isEnum {
			if sub := reEnum.FindStringSubmatch(line); sub != nil {
				isEnum = true
				lastEnumValue = -1
				if !isConstBlock {
					self.Raw.WriteString("const (\n")
					isConstBlock = true
				}
				self.Raw.WriteString(fmt.Sprintf("// enum %s\n",
					strings.Title(sub[2])))
				continue
			}
		} else {
			if sub := reEnumEnd.FindStringSubmatch(line); sub != nil {
				self.Raw.WriteString("\n")
				isEnum = false
				continue
			}
			if sub := reEnumValue.FindStringSubmatch(line); sub != nil {
				self.Raw.WriteString(fmt.Sprintf("%s = %s\n",
					strings.Title(sub[1]), sub[2]))
				if v, err := strconv.Atoi(sub[2]); err == nil {
					lastEnumValue = v
				} else {
					lastEnumValue = NoLastEnumValue
				}
				continue
			}
			if sub := reEnumIota.FindStringSubmatch(line); sub != nil {
				if lastEnumValue != NoLastEnumValue {
					lastEnumValue++
					self.Raw.WriteString(fmt.Sprintf("%s = %d\n",
						strings.Title(sub[1]), lastEnumValue))
					continue
				}
			}
			self.Raw.WriteString(fmt.Sprintf("%s%s", COMMENT_LINE, line))
			continue
		}

		// === Translate structs.
		if !isStruct {
			if sub := reStruct.FindStringSubmatch(line); sub != nil {
				isStruct = true

				if isConstBlock {
					self.Raw.WriteString(")\n")
					isConstBlock = false
				}

				self.Raw.WriteString(fmt.Sprintf(
					"type %s struct {\n", strings.Title(sub[2])))
				continue
			}
		} else {
			if sub := reStructField.FindStringSubmatch(line); sub != nil {
				// Translate the field type.
				gotype, ok := ctypeTogo(sub[1], extraTypedef)

				// === Translate the field name.
				fieldName := reStructFieldName.FindStringSubmatch(sub[2])
				_fieldName := ""

				if fieldName[1] != "" {
					_fieldName = fieldName[2]
				} else {
					_fieldName = fieldName[0]
				}
				// ===

				line = fmt.Sprintf("%s %s %s",
					strings.Title(_fieldName), gotype, sub[3])

				// C type not found.
				if !ok {
					line = COMMENT_LINE + line
				}

				self.Raw.WriteString(line)
				continue
			}
			if strings.HasPrefix(line, "}") {
				self.Raw.WriteString(strings.Replace(line, ";", "", 1))
				isStruct = false
				continue
			}
		}

		// Comment another C lines.
		//if line != "\n" && !strings.HasPrefix(line, "//") {
		if !reSkip.MatchString(line) {
			line = COMMENT_LINE + line
		}

		self.Raw.WriteString(line)
	}

	return nil
}

// Translates a C type definition into Go definition. The C header could have
// defined new types so they're checked in the firs place.
func ctypeTogo(ctype string, extraCtype []string) (gotype string, ok bool) {
	for _, v := range extraCtype {
		if v == ctype {
			return ctype, true
		}
	}

	switch ctype {
	case "char", "signed char":
		return "int8", true
	case "unsigned char":
		return "uint8", true
	case "short", "signed short", "short int", "signed short int":
		return "int16", true
	case "unsigned short", "unsigned short int":
		return "uint16", true
	case "int", "signed int", "signed":
		return "int16", true
	case "unsigned int", "unsigned":
		return "uint16", true
	case "long", "signed long", "long int", "signed long int":
		return "int32", true
	case "unsigned long", "unsigned long int":
		return "uint32", true
	case "size_t":
		return "int", true
	case "float":
		return "float32", true
	case "double", "long double":
		return "float64", true
	}
	return ctype, false
}
