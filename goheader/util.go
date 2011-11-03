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
	"os"
	"strings"
	"path/filepath"
)

func isHeader(f *os.FileInfo) bool {
	return f.IsRegular() && !strings.HasPrefix(f.Name, ".") &&
		strings.HasSuffix(f.Name, ".h")
}

//
// === Walk into a directory

func walkDir(path string) {
	errors := make(chan error)
	done := make(chan bool)

	// Error handler
	go func() {
		for err := range errors {
			if err != nil {
				reportError(err)
			}
		}
		done <- true
	}()

	filepath.Walk(path, walkFn(errors)) // Walk the tree.
	close(errors)                       // Terminate error handler loop.
	<-done                              // Wait for all errors to be reported.
}

// Implements "filepath.WalkFunc".
func walkFn(errors chan error) filepath.WalkFunc {
	return func(path string, info *os.FileInfo, err error) error {
		if err != nil {
			errors <- err
			return nil
		}

		if isHeader(info) {
			if err := processFile(path); err != nil {
				errors <- err
			}
			return nil
		}

		return nil
	}
}
