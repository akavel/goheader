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
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"labix.org/v2/pipe"

	"github.com/akavel/goheader/h2go"
)

// Flags
var (
	debug = flag.Bool("d", false, "If set, outputs the source code translated not formatted.")
	gcc   = flag.String("gcc", "", "Path to GNU C compiler executable")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: goheader [-d] -gcc PATH_TO_GCC PATH_TO_HEADER.h\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func run() error {
	flag.Usage = usage
	flag.Parse()
	if *gcc == "" {
		flag.Usage()
	}
	if len(flag.Args()) == 0 {
		flag.Usage()
	}
	header := flag.Args()[0]

	if runtime.GOOS == "windows" {
		os.Setenv("PATH", filepath.Dir(*gcc)+";"+os.Getenv("PATH"))
	}

	p := pipe.Line(
		pipe.Exec(*gcc, "-E", header),
		pipe.Filter(func(line []byte) bool { return !bytes.HasPrefix(line, []byte{'#'}) }), // strip line-no marks
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

/*

// Flags
var (
	system      = flag.String("s", "", "The operating system.")
	pkgName     = flag.String("p", "", "The name of the package.")
	listSystems = flag.Bool("l", false, "List of valid systems.")
	write       = flag.Bool("w", false, "If set, write its output to file.")
	debug       = flag.Bool("d", false,
		"If set, it outputs the source code translated but without be formatted.")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: goheader -s system -p package [-d] [path ...]\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func processFile(filename string) error {
	file, err := os.OpenFile(filename, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer file.Close()

	_translate := &h2go.Translate{filename, &bytes.Buffer{}, &bytes.Buffer{}}

	if err := _translate.C(file); err != nil {
		return err
	}

	err = _translate.Format()
	if !*debug && err != nil {
		return err
	}

	if err := Write(_translate); err != nil {
		return err
	}

	return nil
}

//
// === Main

func main() {
	validSystems := []string{"linux", "freebsd", "openbsd", "darwin", "plan9"}
	var isSystem bool

	// === Parse the flags
	flag.Usage = usage
	flag.Parse()

	if *listSystems {
		fmt.Print("  = Systems\n\n  ")
		fmt.Println(validSystems)
		os.Exit(0)
	}
	if len(os.Args) == 1 || *system == "" || *pkgName == "" {
		usage()
	}

	*system = strings.ToLower(*system)

	for _, v := range validSystems {
		if v == *system {
			isSystem = true
			break
		}
	}
	if !isSystem {
		fmt.Fprintf(os.Stderr, "ERROR: System passed in flag '-s' is invalid\n")
		os.Exit(2)
	}

	// === Update Go base
	cmd := strings.Join(os.Args, " ")
	h2go.GoBase = strings.Replace(h2go.GoBase, "{cmd}", cmd, 1)
	h2go.GoBase = strings.Replace(h2go.GoBase, "{pkg}", *pkgName, 1)

	// === Translate all headers passed in command line.
	for _, path := range flag.Args() {
		switch info, err := os.Stat(path); {
		case err != nil:
			reportError(err)
		case !info.IsDir():
			if err := processFile(path); err != nil {
				reportError(err)
			}
		case info.IsDir():
			walkDir(path)
		}
	}

	os.Exit(exitCode)
}

func Write(self *h2go.Translate) error {
	output := new(bytes.Buffer)

	if !*debug {
		output = self.Fmt
	} else {
		output = self.Raw
	}

	if *write {
		filename := strings.SplitN(path.Base(self.Filename), ".h", 2)[0]
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
*/
