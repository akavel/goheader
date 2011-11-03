// Copyright 2010  The "GoHeader" Authors
//
// Use of this source code is governed by the BSD-2 Clause license
// that can be found in the LICENSE file.
//
// This software is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied. See the License
// for more details.

package main

import (
	"fmt"
	"os"
)

var exitCode = 0

func reportError(err error) {
	fmt.Fprintf(os.Stderr, "%s\n", err.Error())
	exitCode = 2
}
