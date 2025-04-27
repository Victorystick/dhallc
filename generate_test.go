package dhallc

import (
	"strings"
	"testing"

	"github.com/philandstuff/dhall-golang/v6/parser"
)

func expect(t *testing.T, dhall string, expected string) {
	term, err := parser.Parse("test.dhall", []byte(dhall))
	if err != nil {
		t.Error("error:", err)
		return
	}
	result, err := Generate(term)
	if err != nil {
		t.Error("error:", err)
		return
	}
	if result != expected {
		t.Errorf("'%s' != '%s'", result, expected)
	}
}

// func TestWhatIs(t *testing.T) {
// 	term, err := parser.Parse("test.dhall", []byte("Text -> Text"))
// 	if err != nil {
// 		t.Error("error:", err)
// 		return
// 	}
// 	t.Errorf("%v : %T", term, term)
// }

func TestConstants(t *testing.T) {
	expect(t, `True`, `true`)
	expect(t, `False`, `false`)
	expect(t, `1`, `1`)
	expect(t, `"!"`, `"!"`)

	expect(t, `let a = True in a`, "\nvar a = true\na")

	// Shadowing
	expect(t, `let x = 1 let x = x + 2 in x`, "\nvar x = 1\n\nvar x = x + 2\nx")

	expect(t, `let incr = \(x : Natural) -> x + 1 in 2`, `
func incr(x uint) uint {
  return x + 1
}
2`)

	expect(t, `foo 1 2`, "foo(1)(2)")

	expect(t, `let add = \(x : Natural) -> \(y : Natural) -> x + y in add`, `
func add(x uint) func(uint) uint {
  return func(y uint) uint {
  return x + y
}
}
add`)
}

func expectPackage(t *testing.T, dhall string, expected string) {
	term, err := parser.Parse("test.dhall", []byte(dhall))
	if err != nil {
		t.Error("error:", err)
		return
	}
	result, err := GeneratePackage("example", term)
	if err != nil {
		t.Error("error:", err)
		return
	}
	preamble := "package example\n\n"
	if !strings.HasPrefix(result, preamble) {
		t.Log(dhall)
		t.Errorf("Expected to start with %s", preamble)
		return
	}
	result = result[len(preamble):]
	if result != expected {
		t.Errorf("'%s' != '%s'", result, expected)
	}
}

func TestPackage(t *testing.T) {
	expectPackage(t, `{ A = True }`, "var A = true\n")
	expectPackage(t, `let C : Type = { a : Text } let c : C = { a = "Hi" } in { A = c }`, `type C struct {
  a string
}

var c = C{
  a: "Hi",
}

var A = c
`)
}

func TestPackageStripsAliases(t *testing.T) {
	expectPackage(t, `{ B = B }`, "")
}
