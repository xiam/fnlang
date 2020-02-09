package context

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

var ctxID = uint64(0)

type Context struct {
	id   uint64
	name string

	Parent *Context

	executable bool

	ticket chan struct{}

	inMu  sync.Mutex
	mu    sync.Mutex
	argMu sync.Mutex

	in       chan *Value
	inClosed bool

	doneAccept chan struct{}

	out       chan *Value
	outClosed bool

	accept chan struct{}

	exitStatus   error
	lastArgument *Value

	st *symbolTable
}

func (ctx *Context) Closed() bool {
	return ctx.outClosed
}

func (ctx *Context) Name(name string) *Context {
	ctx.name = name
	return ctx
}

func (ctx *Context) IsExecutable() bool {
	return ctx.executable
}

func (ctx *Context) NonExecutable() *Context {
	ctx.executable = false
	return ctx
}

func (ctx *Context) Executable() *Context {
	ctx.executable = true
	return ctx
}

func (ctx *Context) closeIn() {
	if ctx.inClosed {
		return
	}

	close(ctx.accept)
	close(ctx.in)

	ctx.inClosed = true
}

func (ctx *Context) exit(err error) error {
	if err != nil {
		ctx.exitStatus = err
	}
	return nil
}

func (ctx *Context) Next() bool {
	ctx.mu.Lock()
	if ctx.inClosed {
		ctx.mu.Unlock()
		return false
	}
	ctx.accept <- struct{}{}
	ctx.mu.Unlock()

	var ok bool
	ctx.lastArgument, ok = <-ctx.in
	if !ok {
		return false
	}
	return true
}

func (ctx *Context) Arguments() ([]*Value, error) {
	if ctx.inClosed {
		return nil, ErrStreamClosed
	}
	args := []*Value{}
	for ctx.Next() {
		arg, err := ctx.Argument()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
	}
	return args, nil
}

func (ctx *Context) Argument() (*Value, error) {
	if ctx.executable {
		return ExecArgument(ctx, ctx.lastArgument)
	}
	return ctx.lastArgument, nil
}

func (ctx *Context) Exit(err error) {
	if ctx.outClosed {
		return
	}

	close(ctx.out)
	ctx.outClosed = true
	ctx.Close()
}

func (ctx *Context) Close() {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	if ctx.inClosed {
		return
	}
	ctx.inClosed = true
	close(ctx.doneAccept)
	close(ctx.accept)
	close(ctx.in)
}

func (ctx *Context) Push(value *Value) error {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	if ctx.inClosed {
		return errors.New("channel is closed")
	}
	ctx.in <- value
	return nil
}

func (ctx *Context) Return(values ...*Value) error {
	if err := ctx.Yield(values...); err != nil {
		ctx.Exit(err)
		return err
	}

	ctx.Exit(nil)
	return nil
}

func (ctx *Context) Accept() bool {
	if ctx.inClosed {
		return false
	}
	select {
	case <-ctx.doneAccept:
	case <-ctx.accept:
		return true
	}
	return false
}

func (ctx *Context) Yield(values ...*Value) error {
	for i := range values {
		if err := ctx.yield(values[i]); err != nil {
			return err
		}
	}
	return nil
}

func (ctx *Context) yield(value *Value) error {
	if ctx.outClosed {
		return nil
	}
	if value == nil {
		panic("can't yield nil value")
	}
	ctx.out <- value
	return nil
}

func (ctx *Context) Output() (*Value, error) {
	out, ok := <-ctx.out
	if !ok {
		return nil, ErrClosedChannel
	}
	return out, nil
}

func (ctx *Context) Results() (*Value, error) {
	values, err := ctx.Collect()
	if err != nil {
		return nil, err
	}
	return NewArrayValue(values), nil
}

func (ctx *Context) Result() (*Value, error) {
	values, err := ctx.Collect()
	if err != nil {
		return nil, err
	}
	if len(values) > 1 {
		return nil, errors.New("expecting one result")
	}
	if len(values) < 1 {
		return Nil, nil
	}
	return values[0], nil
}

func (ctx *Context) Collect() ([]*Value, error) {
	values := []*Value{}

	for {
		value, err := ctx.Output()
		if err == ErrClosedChannel {
			break
		}
		values = append(values, value)
	}

	return values, nil
}

func (ctx *Context) String() string {
	return fmt.Sprintf("[%v]: %q (%p)", ctx.id, ctx.name, ctx)
}

func (ctx *Context) Set(name string, value *Value) error {
	if !ctx.executable {
		return errors.New("cannot set on a non-executable context")
	}
	return ctx.st.Set(name, value)
}

func (ctx *Context) Get(name string) (*Value, error) {
	value, err := ctx.st.Get(name)
	if err != nil {
		if ctx.Parent != nil {
			return ctx.Parent.Get(name)
		}
		return nil, err
	}
	return value, nil
}

func NewClosure(parent *Context) *Context {
	ctx := New(parent)
	if parent != nil {
		ctx.st = parent.st
	}
	return ctx
}

func New(parent *Context) *Context {
	ctx := &Context{
		id:     atomic.AddUint64(&ctxID, 1),
		ticket: make(chan struct{}),

		accept:     make(chan struct{}, 1),
		doneAccept: make(chan struct{}),

		in:  make(chan *Value),
		out: make(chan *Value),
	}
	if parent == nil {
		ctx.st = newSymbolTable(nil)
		ctx.executable = true
	} else {
		ctx.Parent = parent
		ctx.executable = parent.executable
		ctx.st = newSymbolTable(parent.st)
	}
	return ctx
}

func ExecArgument(ctx *Context, value *Value) (*Value, error) {
	switch value.Type() {
	case ValueTypeInt:
		return value, nil
	case ValueTypeList:
		var err error
		m := value.List()
		for k := range m {
			m[k], err = ExecArgument(ctx, m[k])
			if err != nil {
				return nil, err
			}
		}
		return value, nil
	case ValueTypeMap:
		var err error
		m := value.Map()
		for k := range m {
			m[k], err = ExecArgument(ctx, m[k])
			if err != nil {
				return nil, err
			}
		}
		return value, nil
	case ValueTypeSymbol:
		v, err := ctx.Get(value.String())
		if err != nil {
			return nil, err
		}
		return v, nil
	case ValueTypeFunction:
		newCtx := New(ctx).Name("argument")
		go func() {
			defer newCtx.Exit(nil)
			if err := value.Function().Exec(newCtx); err != nil {
				panic(err.Error())
			}
		}()
		col, err := newCtx.Results()
		return col.List()[0], err
	}
	return value, nil
}
