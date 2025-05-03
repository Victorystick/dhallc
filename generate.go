package dhallc

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/philandstuff/dhall-golang/v6/term"
)

// TODO: Use go/printer instead?
type builder struct {
	strings.Builder
	// Path to name.
	imports map[string]string
	indent  string
	types   Scope
	// The current type?
	current string
}

type Scope struct {
	terms  map[string]term.Term
	parent *Scope
}

// Defines a type within the scope.
func (s Scope) Define(name string, t term.Term) {
	// fmt.Println("defining", name, "as", t)
	s.terms[name] = t
}

// Looks up the given name in the scopes.
func (s Scope) LookUp(name string) (term.Term, error) {
	term, ok := s.terms[name]
	if ok {
		return term, nil
	}
	if s.parent != nil {
		return s.parent.LookUp(name)
	}
	return nil, fmt.Errorf("unknown var '%s'", name)
}

// Creates a sub-scope.
func (s *Scope) Push() {
	child := ChildScope(s)
	s = &child
}

func ChildScope(s *Scope) Scope {
	return Scope{terms: make(map[string]term.Term), parent: s}
}

// Removes a sub-scope.
func (s *Scope) Pop() {
	s = s.parent
}

func MakeBuilder() builder {
	return builder{
		imports: make(map[string]string),
		types:   Scope{terms: make(map[string]term.Term)},
	}
}

// Indent increases the indentation.
func (b *builder) Indent() {
	b.indent = "  " + b.indent
	b.WriteString("{")
}

func (b *builder) NewLine() {
	b.WriteByte('\n')
	b.WriteString(b.indent)
}

func (b *builder) Dedent() {
	// Will panic ?
	b.indent = b.indent[2:]
	b.NewLine()
	b.WriteString("}")
}

// Imports the helper library and uses name.
func (b *builder) Lib(name string) {
	b.imports["github.com/Victorystick/dhallc/lib"] = ""
	b.WriteString("lib.")
	b.WriteString(name)
}

func GeneratePackage(filename string, t term.Term) (string, error) {
	p := MakeBuilder()
	p.WriteString("package ")
	p.WriteString(filename)
	p.WriteByte('\n')

	b := MakeBuilder()
	err := generatePackage(&b, t)
	if err != nil {
		fmt.Println(b.String())
		return "", err
	}

	b.WriteByte('\n')

	if len(b.imports) > 0 {
		p.NewLine()
		p.WriteString("import (")
		p.indent = "\t"
		for k := range b.imports {
			p.NewLine()
			p.WriteByte('"')
			p.WriteString(k)
			p.WriteByte('"')
		}
		p.indent = ""
		p.NewLine()
		p.WriteByte(')')
		p.NewLine()
	}

	p.WriteString(b.String())

	return p.String(), nil
}

func generatePackage(b *builder, t term.Term) error {
	switch t := t.(type) {
	case term.Let:
		for _, binding := range t.Bindings {
			err := writeLet(b, binding)
			if err != nil {
				return err
			}
			b.WriteByte('\n')
		}
		return generatePackage(b, t.Body)

	case term.RecordLit:
		// For stable outputs.
		for _, key := range slices.Sorted(maps.Keys(t)) {
			val := t[key]
			if !IsExported(key) {
				return fmt.Errorf("unexported field '%s' of %v", key, t)
			}

			// Special-case the redundant let A = A.
			switch v := val.(type) {
			case term.Var:
				if v.Name == key {
					continue
				}
			}

			b.WriteString("\nvar ")
			b.WriteString(key)
			b.WriteString(" = ")
			err := generate(b, val)
			if err != nil {
				return err
			}
		}

		return nil
	}

	return fmt.Errorf("cannot generate package from term %v: %T", t, t)
}

// IsExported reports whether name starts with an upper-case letter.
func IsExported(name string) bool {
	ch, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(ch)
}

// Generates a Go expression from the given term.
func Generate(t term.Term) (string, error) {
	b := MakeBuilder()

	err := generate(&b, t)
	if err != nil {
		return "", err
	}

	return b.String(), nil
}

// Generates a Go expression from the given term.
func generate(b *builder, val term.Term) error {
	// Values
	switch val {
	case term.True:
		_, err := b.WriteString("true")
		return err
	case term.False:
		_, err := b.WriteString("false")
		return err
	case term.Natural:
		b.WriteString("uint")
		return nil
	case term.Integer:
		b.WriteString("int")
		return nil
	}

	// Types
	switch val := val.(type) {
	case term.Builtin:
		// This is a hack to serialize build-in functions.
		// But I'll try it.
		parts := strings.SplitN(string(val), "/", 2)
		b.Lib(parts[0] + strings.Title(parts[1]))
		return nil

	case term.NaturalLit:
		b.WriteString(strconv.FormatUint(uint64(val), 10))
		return nil

	case term.IntegerLit:
		b.WriteString(strconv.FormatInt(int64(val), 10))
		return nil

	case term.Op:
		return writeOp(b, val)

	case term.Var:
		b.WriteString(val.Name)
		// TODO: Handle shadowing?
		// b.WriteString(strconv.FormatUint(uint64(t.Index), 10))
		return nil

	case term.TextLit:
		if len(val.Chunks) == 0 {
			b.WriteString(strconv.Quote(val.Suffix))
			return nil
		} else {
			for _, chunk := range val.Chunks {
				b.WriteString(strconv.Quote(chunk.Prefix))
				b.WriteString(" + ")
				err := generate(b, chunk.Expr)
				if err != nil {
					return err
				}
				b.WriteString(" + ")
			}
			b.WriteString(strconv.Quote(val.Suffix))
			return nil
		}

	case term.Let:
		for _, binding := range val.Bindings {
			err := writeLet(b, binding)
			if err != nil {
				return err
			}
			b.NewLine()
		}
		return generate(b, val.Body)

	case term.App:
		err := generate(b, val.Fn)
		if err != nil {
			return err
		}
		if isType(val.Arg) {
			// TODO: Do we need this or can Go infer types?
			// b.WriteByte('[')
			// err = generate(b, val.Arg)
			// if err != nil {
			// 	return err
			// }
			// b.WriteByte(']')
			return nil
		}
		b.WriteByte('(')
		err = generate(b, val.Arg)
		if err != nil {
			return err
		}
		b.WriteByte(')')
		return nil

	case term.Lambda:
		b.WriteString("func")
		return writeFn(b, val, nil)

	case term.RecordLit:
		// This should be done another way. :/
		// Should we have a "current type"?
		typ, err := InferType(b.types, val)
		if err != nil {
			return err
		}
		return writeRecordLit(b, val, typ)

	case term.EmptyList:
		return writeListLit(b, nil, val.Type)

	case term.NonEmptyList:
		// This should be done another way. :/
		// Should we have a "current type"?
		typ, err := InferType(b.types, val[0])
		if err != nil {
			return err
		}
		return writeListLit(b, val, typ)
	}

	return fmt.Errorf("unhandled term %v: %T", val, val)
}

func writeRecordLit(b *builder, val term.RecordLit, typ term.Term) error {
	err := writeType(b, typ)
	if err != nil {
		return err
	}

	// For stable outputs.
	b.Indent()
	for _, key := range slices.Sorted(maps.Keys(val)) {
		val := val[key]
		b.NewLine()
		b.WriteString(key)
		b.WriteString(": ")
		err := generate(b, val)
		if err != nil {
			return err
		}
		b.WriteByte(',')
	}
	b.Dedent()
	return nil
}

func writeListLit(b *builder, val []term.Term, typ term.Term) error {
	b.WriteString("[]")
	err := writeType(b, typ)
	if err != nil {
		return err
	}
	b.Indent()
	for _, val := range val {
		b.NewLine()
		err := generate(b, val)
		if err != nil {
			return err
		}
		b.WriteByte(',')
	}
	b.Dedent()
	return nil
}

func writeLet(b *builder, bd term.Binding) (err error) {
	typ := bd.Annotation
	val := bd.Value
	if typ == nil || typ == term.Type {
		typ, err = InferType(b.types, val)
		if err != nil {
			return err
		}
	}

	b.types.Define(bd.Variable, typ)

	// If type, not value.
	if bd.Annotation == term.Type {
		b.NewLine()
		b.WriteString("type ")
		b.WriteString(bd.Variable)
		b.WriteByte(' ')
		return writeType(b, typ)
	}

	if fn, ok := val.(term.Lambda); ok {
		if pi, ok := typ.(term.Pi); ok {
			b.NewLine()
			b.WriteString("func ")
			b.WriteString(bd.Variable)
			return writeFn(b, fn, pi.Body)
		} else {
			return fmt.Errorf("bad fn def %v: %T", fn, fn)
		}
	}

	b.NewLine()
	b.WriteString("var ")
	b.WriteString(bd.Variable)
	b.WriteString(" = ")
	return writeTyped(b, val, typ)
}

// Defines a
func writeFn(b *builder, l term.Lambda, ret term.Term) (err error) {
	b.WriteByte('(')
	b.WriteString(l.Label)
	b.WriteByte(' ')
	err = writeType(b, l.Type)
	if err != nil {
		return err
	}
	b.WriteString(") ")

	if ret == nil {
		ret, err = InferType(b.types, l.Body)
		if err != nil {
			return err
		}
	}

	// Define a new scope for the body.
	b.types.Push()
	defer b.types.Pop()

	err = writeType(b, ret)
	if err != nil {
		return err
	}

	b.WriteByte(' ')
	b.Indent()
	err = writeFnBody(b, l.Body, ret)
	if err != nil {
		return err
	}
	b.Dedent()
	return nil
}

func writeFnBody(b *builder, t term.Term, ret term.Term) error {
	if let, ok := t.(term.Let); ok {
		for _, binding := range let.Bindings {
			b.NewLine()
			err := writeLet(b, binding)
			if err != nil {
				return err
			}
		}

		b.NewLine()
		b.WriteString("return ")
		return writeTyped(b, let.Body, ret)
	}

	b.NewLine()
	b.WriteString("return ")
	return writeTyped(b, t, ret)
}

func writeTyped(b *builder, val term.Term, typ term.Term) error {
	switch val := val.(type) {
	case term.RecordLit:
		return writeRecordLit(b, val, typ)
	case term.NonEmptyList:
		return writeListLit(b, val, typ)
	}
	return generate(b, val)
}

func writeOp(b *builder, op term.Op) error {
	switch op.OpCode {
	case term.PlusOp:
		return writeInfixOp(b, " + ", op)
	case term.TimesOp:
		return writeInfixOp(b, " * ", op)
	case term.TextAppendOp:
		err := generate(b, op.L)
		if err != nil {
			return err
		}
		b.WriteString(" + ")
		return generate(b, op.R)
	case term.ListAppendOp:
		b.Lib("ListConcat(")
		err := generate(b, op.L)
		if err != nil {
			return err
		}
		b.WriteString(", ")
		err = generate(b, op.R)
		if err != nil {
			return err
		}
		b.WriteByte(')')
		return nil
	}

	return fmt.Errorf("unhandled operation %v", op.OpCode)
}

func writeInfixOp(b *builder, infix string, op term.Op) error {
	err := generate(b, op.L)
	if err != nil {
		return err
	}
	b.WriteString(infix)
	return generate(b, op.R)
}

func writeType(b *builder, t term.Term) error {
	switch t {
	case term.Double:
		b.WriteString("uint")
		return nil
	case term.Text:
		b.WriteString("string")
		return nil
	case term.Bool:
		b.WriteString("bool")
		return nil
	case term.Natural:
		b.WriteString("uint")
		return nil
	case term.Integer:
		b.WriteString("int")
		return nil

	// Should we make a generic class for this?
	case term.List:
		b.WriteString("List")
		return nil
	}

	switch t := t.(type) {
	case term.Var:
		b.WriteString(t.Name)
		return nil

	case term.Pi:
		b.WriteString("func(")
		err := writeType(b, t.Type)
		if err != nil {
			return err
		}
		b.WriteString(") ")
		return writeType(b, t.Body)

	case term.RecordType:
		b.WriteString("struct ")
		b.Indent()
		// For stable outputs.
		for _, key := range slices.Sorted(maps.Keys(t)) {
			val := t[key]
			b.NewLine()
			b.WriteString(key)
			b.WriteByte(' ')
			err := writeType(b, val)
			if err != nil {
				return err
			}
		}
		b.Dedent()

		return nil

	case term.NonEmptyList:
		b.WriteString("[]")
		return writeType(b, t[0])

	case term.App:
		if t.Fn == term.List {
			return writeType(b, t.Arg)
		}

		// Generics?
		err := writeType(b, t.Fn)
		if err != nil {
			return err
		}
		b.WriteByte('[')
		err = writeType(b, t.Arg)
		if err != nil {
			return err
		}
		b.WriteByte(']')
		return nil

	case term.Op:
		switch t.OpCode {
		case term.PlusOp:
			return writeType(b, term.Natural)
		case term.TimesOp:
			return writeType(b, term.Natural)
		case term.TextAppendOp:
			return writeType(b, term.Text)
		}
	}

	return fmt.Errorf("unhandled type %v: %T", t, t)
}

func isType(t term.Term) bool {
	if t == nil {
		return false
	}

	switch t {
	case term.Natural:
		return true
	case term.Integer:
		return true
	}

	switch t.(type) {
	case term.RecordType:
		return true
	case term.Pi:
		return true
	}

	return false
}

var (
	ListLength      = term.NewAnonPi(term.Type, term.NewAnonPi(term.List, term.Natural))
	NaturalSubtract = term.NewAnonPi(term.Natural, term.NewAnonPi(term.Natural, term.Natural))
)

func InferType(scope Scope, t term.Term) (term.Term, error) {
	switch t {
	case term.ListLength:
		return ListLength, nil
	case term.NaturalSubtract:
		return NaturalSubtract, nil
	}

	switch t := t.(type) {
	// We shouldn't return self. :/
	case term.Builtin:
		return t, nil
	case term.RecordType:
		return t, nil

	case term.EmptyList:
		return t.Type, nil
	case term.NonEmptyList:
		return InferType(scope, t[0])

	case term.Let:
		child := ChildScope(&scope)
		for _, binding := range t.Bindings {
			t := binding.Annotation
			if !isType(t) {
				var err error
				t, err = InferType(child, binding.Value)
				if err != nil {
					return nil, err
				}
			}
			child.Define(binding.Variable, t)
		}
		return InferType(child, t.Body)

		// Literals
	case term.DoubleLit:
		return term.Double, nil
	case term.TextLit:
		return term.Text, nil
	case term.BoolLit:
		return term.Bool, nil
	case term.NaturalLit:
		return term.Natural, nil
	case term.IntegerLit:
		return term.Integer, nil

	case term.Op:
		switch t.OpCode {
		case term.PlusOp:
			return term.Natural, nil
		case term.TimesOp:
			return term.Natural, nil
		case term.TextAppendOp:
			return term.Text, nil
		case term.ListAppendOp:
			return InferType(scope, t.L)
		}

	case term.Lambda:
		arg, err := InferType(scope, t.Type)
		if err != nil {
			return nil, err
		}
		res, err := InferType(scope, t.Body)
		if err != nil {
			return nil, err
		}
		return term.Pi{Label: t.Label, Type: arg, Body: res}, nil

	case term.App:
		fn, err := InferType(scope, t.Fn)
		if err != nil {
			return nil, err
		}
		if pi, ok := fn.(term.Pi); ok {
			return pi.Body, nil
		}
		return nil, fmt.Errorf("cannot call non-lambda %v: %T", fn, fn)

	case term.Var:
		return scope.LookUp(t.Name)

	case term.RecordLit:
		res := term.RecordType{}
		for key, val := range t {
			typ, err := InferType(scope, val)
			if err != nil {
				return nil, err
			}
			res[key] = typ
		}
		return res, nil
	}

	return nil, fmt.Errorf("uninferred type %v: %T", t, t)
}
