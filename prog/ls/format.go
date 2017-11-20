package ls

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	structform "github.com/urso/go-structform"
)

type formatCtx struct {
	err error
	out io.Writer

	indent         string
	currentIndent  string
	indentRequired bool
}

type paramPrinter struct {
	ctx *formatCtx

	first   boolStack
	inArray boolStack
}

type boolStack struct {
	stack   []bool
	stack0  [32]bool
	current bool
}

func (c *formatCtx) withIndent(fn func(*formatCtx) error) error {
	old := c.currentIndent
	c.currentIndent = old + c.indent
	defer func() { c.currentIndent = old }()
	return fn(c)
}

func (c *formatCtx) increaseIndent() {
	c.currentIndent += c.indent
}

func (c *formatCtx) decreaseIndent() {
	if len(c.currentIndent) > 0 {
		end := len(c.currentIndent) - len(c.indent)
		c.currentIndent = c.currentIndent[:end]
	}
}

func (c *formatCtx) Indent() string {
	return c.currentIndent
}

func (c *formatCtx) Err() error {
	return c.err
}

func (c *formatCtx) Printf(format string, values ...interface{}) error {
	return c.Write(fmt.Sprintf(format, values...))
}

func (c *formatCtx) Println(s string) error {
	return c.Write(s + "\n")
}

func (c *formatCtx) Write(s string) error {
	if c.err != nil {
		return c.err
	}

	for len(s) > 0 {
		idx := strings.IndexRune(s, '\n')
		if idx < 0 {
			c.doWriteString(s)
			if c.indentRequired {
				c.indentRequired = false
			}
			return c.Err()
		}

		line := s[:idx] + "\n"
		s = s[idx+1:]

		if err := c.doWriteString(line); err != nil {
			return err
		}

		c.indentRequired = true
	}

	return nil
}

func (c *formatCtx) doWriteString(s string) error {
	if c.indentRequired && s != "\n" {
		c.doWriteString_(c.Indent())
	}
	return c.doWriteString_(s)
}

func (c *formatCtx) doWriteString_(s string) error {
	for len(s) > 0 {
		n, err := io.WriteString(c.out, s)
		s = s[n:]
		if err != nil {
			c.err = err
			return err
		}
	}
	return nil
}

func newParamPrinter(ctx *formatCtx) *paramPrinter {
	p := &paramPrinter{ctx: ctx}
	p.first.init()
	p.inArray.init()
	return p
}

func (p *paramPrinter) tryElemNext() error {
	if !p.inArray.current {
		return nil
	}

	if p.first.current {
		p.first.current = false
		return nil
	}

	return p.ctx.doWriteString(",\n")
}

func (p *paramPrinter) enter(isArray bool) {
	p.first.push(true)
	p.inArray.push(isArray)
	p.ctx.increaseIndent()
}

func (p *paramPrinter) exit() {
	p.ctx.decreaseIndent()
	p.first.pop()
	p.inArray.pop()
}

func (p *paramPrinter) OnObjectStart(len int, baseType structform.BaseType) error {
	if err := p.tryElemNext(); err != nil {
		return err
	}

	p.ctx.Printf("{")
	p.enter(false)
	return p.ctx.Err()
}

func (p *paramPrinter) OnObjectFinished() error {
	p.exit()
	return p.ctx.Println("\n}")
}

func (p *paramPrinter) OnKey(s string) error {
	if len(p.inArray.stack) > 1 || s[0] == '[' && s[len(s)-1] == ']' {
		s = strconv.Quote(s)
	}
	return p.ctx.Printf("\n%v => ", s)
}

func (p *paramPrinter) OnArrayStart(len int, baseType structform.BaseType) error {
	if err := p.tryElemNext(); err != nil {
		return err
	}

	p.ctx.Printf("[\n")
	p.enter(true)
	return p.ctx.Err()
}

func (p *paramPrinter) OnArrayFinished() error {
	p.exit()
	return p.ctx.Printf("\n]")
}

func (p *paramPrinter) OnNil() error {
	return p.onValue(`""`)
}

func (p *paramPrinter) OnBool(b bool) error {
	return p.onValue(b)
}

func (p *paramPrinter) OnString(s string) error {
	return p.onValue(strconv.Quote(s))
}

func (p *paramPrinter) OnInt8(i int8) error {
	return p.onValue(i)
}

func (p *paramPrinter) OnInt16(i int16) error {
	return p.onValue(i)
}

func (p *paramPrinter) OnInt32(i int32) error {
	return p.onValue(i)
}

func (p *paramPrinter) OnInt64(i int64) error {
	return p.onValue(i)
}

func (p *paramPrinter) OnInt(i int) error {
	return p.onValue(i)
}

func (p *paramPrinter) OnByte(b byte) error {
	return p.onValue(b)
}

func (p *paramPrinter) OnUint8(u uint8) error {
	return p.onValue(u)
}

func (p *paramPrinter) OnUint16(u uint16) error {
	return p.onValue(u)
}

func (p *paramPrinter) OnUint32(u uint32) error {
	return p.onValue(u)
}

func (p *paramPrinter) OnUint64(u uint64) error {
	return p.onValue(u)
}

func (p *paramPrinter) OnUint(u uint) error {
	return p.onValue(u)
}

func (p *paramPrinter) OnFloat32(f float32) error {
	return p.onValue(f)
}

func (p *paramPrinter) OnFloat64(f float64) error {
	return p.onValue(f)
}

func (p *paramPrinter) onValue(value interface{}) error {
	if err := p.tryElemNext(); err != nil {
		return err
	}
	return p.ctx.Printf("%v", value)
}

func (s *boolStack) init() {
	s.stack = s.stack0[:0]
}

func (s *boolStack) push(b bool) {
	s.stack = append(s.stack, s.current)
	s.current = b
}

func (s *boolStack) pop() {
	if len(s.stack) == 0 {
		panic("pop from empty stack")
	}

	last := len(s.stack) - 1
	s.current = s.stack[last]
	s.stack = s.stack[:last]
}
