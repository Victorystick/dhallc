package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestGoldens(t *testing.T) {
	testdata, err := os.ReadDir("./testdata")

	if err != nil {
		t.Error(err)
		return
	}

	for _, test := range testdata {
		// https://medium.com/swlh/unit-testing-cli-programs-in-go-6275c85af2e7
		os.Args = []string{"./dhallc", "./testdata/" + test.Name()}

		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		main()

		outC := make(chan string)
		go func() {
			var buf bytes.Buffer
			io.Copy(&buf, r)
			outC <- buf.String()
		}()

		w.Close()
		os.Stdout = old
		out := <-outC

		// os.WriteFile("./goldens/"+strings.TrimSuffix(test.Name(), ".dhall")+".go", []byte(out), 0666)
		contents, err := os.ReadFile("./goldens/" + strings.TrimSuffix(test.Name(), ".dhall") + ".go")

		if err != nil {
			t.Error(err)
			return
		}

		expected := string(contents)
		if expected != out {
			t.Errorf("Failed %s - expected:\n\n'%s'\ngot:\n\n'%s'", test.Name(), expected, out)
		}
	}
}

func M() struct{ text string } {
	return struct{ text string }{text: "fo"}
}
