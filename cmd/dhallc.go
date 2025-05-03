package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Victorystick/dhallc"
	"github.com/philandstuff/dhall-golang/v6/parser"
)

var outfile = flag.String("out", "-", "The output file")

func main() {
	flag.Parse()

	filename := flag.Arg(0)
	if filename == "" {
		flag.Usage()
		os.Exit(2)
	}

	term, err := parser.ParseFile(filename)
	check(err)

	extensionless := strings.TrimSuffix(filename, ".dhall")
	str, err := dhallc.GeneratePackage(filepath.Base(extensionless), term)
	check(err)

	if *outfile == "-" {
		os.Stdout.WriteString(str)
	} else {
		os.WriteFile(*outfile, []byte(str), 0666)
	}
}

// Check prints the error and terminates.
func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
