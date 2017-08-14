package walnut

import (
	"fmt"
	"strings"
)

type stateFn func(p *parser) stateFn

type parser struct {
	items chan item
	tape  []byte
	ptr   int
}

const (
	eof       = byte(0)
	clearLoop = "[-]"
)

func parse(tape []byte) <-chan item {
	p := parser{
		items: make(chan item, 1),
		tape:  tape,
		ptr:   0,
	}
	go p.run()
	return p.items
}

func (p *parser) run() {
	for state := dispatcher; state != nil; state = state(p) {
	}
	close(p.items)
}

func (p *parser) incr(n int) {
	p.ptr += n
}

func (p *parser) curr() byte {
	if p.ptr >= len(p.tape) {
		return eof
	}
	return p.tape[p.ptr]
}

func (p *parser) next() byte {
	p.incr(1)
	return p.curr()
}

func (p *parser) backup() {
	p.ptr--
}

func (p *parser) peek() byte {
	r := p.next()
	p.backup()
	return r
}

func (p *parser) emit(i item) {
	p.items <- i
}

func (p *parser) errorf(msg string) stateFn {
	p.emit(parseError{fmt.Sprintf("parse error: %s", msg)})
	return nil
}

func dispatcher(p *parser) stateFn {
	var n int
	for {
		n = 1
		switch r := p.curr(); r {
		case '[':
			if strings.HasPrefix(string(p.tape[p.ptr:]), clearLoop) {
				i := clear{}
				p.emit(i)
				n = len(clearLoop)
			} else {
				return parseLoop
			}
		case '+':
			return parseAdd
		case '-':
			return parseSub
		case '>':
			return parseNext
		case '<':
			return parsePrev
		case '.':
			return parseWrite
		case ',':
			return parseRead
		case '#':
			return ignoreComment
		case eof:
			return nil
		}
		p.incr(n)
	}
}

func parseLoop(p *parser) stateFn {
	begin := p.ptr
	c := p.next()
	var l, r int
	for {
		if c == '[' {
			l++
		} else if c == ']' {
			r++
		}
		if l-r == -1 {
			break
		}
		if p.peek() == eof {
			return p.errorf("unclosed loop")
		}
		c = p.next()
	}
	end := p.ptr
	p.emit(loopStart{})
	for i := range parse(p.tape[begin+1 : end]) {
		p.emit(i)
	}
	p.emit(loopEnd{})
	p.next()
	return dispatcher
}

func parseWrite(p *parser) stateFn {
	p.emit(write{})
	p.next()
	return dispatcher
}

func parseRead(p *parser) stateFn {
	p.emit(read{})
	p.next()
	return dispatcher
}

func parsePrev(p *parser) stateFn {
	n := 1
	for r := p.next(); r == '<'; r = p.next() {
		n++
	}
	p.emit(prev{n})
	return dispatcher
}

func parseNext(p *parser) stateFn {
	n := 1
	for r := p.next(); r == '>'; r = p.next() {
		n++
	}
	p.emit(next{n})
	return dispatcher
}

func parseSub(p *parser) stateFn {
	n := 1
	for r := p.next(); r == '-'; r = p.next() {
		n++
	}
	p.emit(sub{n})
	return dispatcher
}

func parseAdd(p *parser) stateFn {
	n := 1
	for r := p.next(); r == '+'; r = p.next() {
		n++
	}
	p.emit(add{n})
	return dispatcher
}

func ignoreComment(p *parser) stateFn {
	for r := p.next(); r != '\n' && r != eof; r = p.next() {
	}
	p.next()
	return dispatcher
}
