package walnut

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"sync"
	"text/template"
)

// PtrPos represents the data pointer's initial position
type PtrPos int

const (
	// PtrCenter represents a data pointer at the center of program memory
	// The center is defined as (memsize / 2) - 1 for even memsizes and int(memsize / 2)
	// for odd memsizes.
	PtrCenter PtrPos = iota
	// PtrStart represents a data pointer at the beginning of program memory
	PtrStart
	// PtrEnd represents a data pointer at the end of program memory
	PtrEnd
)

var tmpl = template.Must(template.New("compiled").Parse(
	`package main

import "fmt"

func main() {
	var (
		mem [{{.Memsize}}]int
		ptr int = {{.Ptr}}
		err error
		n   int
	)
{{.Code}}
	// Ensure go doesn't complain about unused imports or variables
	func(a ...interface{}) {
		_ = fmt.Sprint()
	}(n, err)
}
`))

type templateOpts struct {
	Memsize int
	Ptr     int
	Code    string
}

// A Compiler accepts Brainfuck source code and compiles it to formatted Go source code.
type Compiler struct {
	w io.Writer
}

var bufPool sync.Pool

func newBuf() *bytes.Buffer {
	if v := bufPool.Get(); v != nil {
		b := v.(*bytes.Buffer)
		b.Reset()
		return b
	}
	return new(bytes.Buffer)
}

// NewCompiler returns a new compiler writing to w
func NewCompiler(w io.Writer) *Compiler {
	return &Compiler{w: w}
}

// Compile compiles the given Brainfuck program to formatted Go source code, returning
// the first error that occurs.
func (c *Compiler) Compile(r io.Reader, memsize int, pos PtrPos) error {
	tape, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	buf := newBuf()
	defer bufPool.Put(buf)

	level := 1
	var tmp string
	var wasLoop bool
	for item := range parse(tape) {
		buf.WriteByte('\n')
		switch i := item.(type) {
		case add:
			tmp = fmt.Sprintf("mem[ptr] += %d", i.N())
		case sub:
			tmp = fmt.Sprintf("mem[ptr] -= %d", i.N())
		case next:
			tmp = fmt.Sprintf("ptr += %d", i.N())
		case prev:
			tmp = fmt.Sprintf("ptr -= %d", i.N())
		case write:
			tmp = `n, err = fmt.Printf("%c", mem[ptr])` + "\n" +
				indent("if n == 0 || err != nil {\n", level) +
				indent("mem[ptr] = 0\n", level+1) +
				indent("}", level)
		case read:
			tmp = `n, err = fmt.Scanf("%c", &mem[ptr])` + "\n" +
				indent("if n == 0 || err != nil {\n", level) +
				indent("mem[ptr] = 0\n", level+1) +
				indent("}", level)
		case clear:
			tmp = "mem[ptr] = 0"
		case loopStart:
			if !wasLoop {
				buf.WriteByte('\n')
			}
			wasLoop = true
			buf.Write(tabs(level))
			buf.WriteString("for mem[ptr] != 0 {")
			buf.WriteByte('\n')
			level++
			continue
		case loopEnd:
			level--
			tmp = "}\n"
		case parseError:
			return i
		}
		wasLoop = false
		buf.Write(tabs(level))
		buf.WriteString(tmp)
	}

	code := buf.String()
	buf.Reset()
	ptr := getAbsPos(memsize, pos)
	err = tmpl.Execute(buf, &templateOpts{memsize, ptr, code})
	if err != nil {
		return err
	}

	_, err = c.w.Write(buf.Bytes())
	return err
}

func tabs(n int) []byte {
	return bytes.Repeat([]byte{'\t'}, n)
}

func indent(s string, n int) string {
	return string(tabs(n)) + s
}

func getAbsPos(memsize int, pos PtrPos) int {
	var ptr int
	switch pos {
	case PtrCenter:
		ptr = (memsize / 2)
		if memsize%2 == 0 {
			ptr--
		}
	case PtrEnd:
		ptr = memsize - 1
	}
	// do nothing on PtrStart, because ints are initialized to 0
	return ptr
}
