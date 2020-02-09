package fnlang

import (
	"fmt"
	"log"

	"github.com/xiam/fnlang/context"
	"github.com/xiam/sexpr/ast"
)

var defaultContext = context.New(nil).Name("root").Executable()

func Defn(name string, fn func(ctx *context.Context) error) {
	wrapper := func(ctx *context.Context) error {
		if err := fn(ctx); err != nil {
			ctx.Exit(err)
			return err
		}
		ctx.Exit(nil)
		return nil
	}
	if err := defaultContext.Set(name, context.NewFunctionValue(wrapper)); err != nil {
		log.Fatal("Defn: %w", err)
	}
}

func execFunctionBody(ctx *context.Context, body *context.Value) error {
	switch body.Type() {
	case context.ValueTypeInt:
		ctx.Yield(body)
		return nil
	case context.ValueTypeFunction:
		log.Printf("execFunctionBody: FUNCTION")
		newCtx := context.New(ctx).Name("exec-body")
		fnErr := make(chan error, 1)
		go func() {
			defer newCtx.Exit(nil)
			fnErr <- body.Function().Exec(newCtx)
		}()
		values, err := newCtx.Results()
		if err != nil {
			return err
		}
		ctx.Yield(values.List()...)
		if err := <-fnErr; err != nil {
			return runtimeError(ctx, body.Node(), err)
		}
		return nil
	case context.ValueTypeList:
		newCtx := context.New(ctx).Name("exec-list")
		log.Printf("execFunctionBody: LIST")
		go func() error {
			defer newCtx.Exit(nil)
			for _, item := range body.List() {
				if err := execFunctionBody(newCtx, item); err != nil {
					return err
				}
			}
			return nil
		}()
		values, err := newCtx.Results()
		if err != nil {
			return err
		}
		ctx.Yield(values)
		return nil
	default:
		log.Fatalf("unhandled type: %v", body.Type())
		panic("unhandled")
	}
}

func derefFunc(ctx *context.Context, fn *context.Function) (*context.Value, error) {
	execCtx := context.New(ctx).Name("deref-exec")

	go func() {
		defer execCtx.Exit(nil)

		if err := fn.Exec(execCtx); err != nil {
			log.Fatalf("ERR1: %v", err)
		}
	}()

	values, err := execCtx.Results()
	if err != nil {
		return nil, err
	}

	if len(values.List()) == 1 {
		return values.List()[0], nil
	}

	return nil, fmt.Errorf("unexpected result")
}

func execFunc(ctx *context.Context, fn *context.Function, args []*context.Value) error {
	log.Printf("FN: %v, ARGS: %v", fn, args)
	go func() {
		defer ctx.Close()

		for i := 0; i < len(args) && ctx.Accept(); i++ {
			log.Printf("CTX: %v, ARGS[%d]: %v (%v)", ctx.IsExecutable(), i, args[i], args[i].Type())
			ctx.Push(args[i])
		}
	}()

	return fn.Exec(ctx)
}

func execExpr(ctx *context.Context, expr *context.Value, values []*context.Value) error {
	log.Printf("EXEC-EXPR: %v", expr.Type())

	switch expr.Type() {
	case context.ValueTypeInt:
		ctx.Yield(expr)
		return nil
	case context.ValueTypeString:
		ctx.Yield(expr)
		return nil
	case context.ValueTypeAtom:

		fn, err := ctx.Get(expr.Atom())
		if err == nil {
			if fn.Type() == context.ValueTypeFunction {
				return execFunc(ctx, fn.Function(), values)
			}
			return execExpr(ctx, fn, values)
		}

		if len(values) > 0 {
			return fmt.Errorf("invalid expression: %v", expr)
		}
		ctx.Yield(expr)
		return nil
	case context.ValueTypeList:
		node, err := mapListItem(expr, values)
		if err != nil {
			return err
		}
		ctx.Yield(node)
		return nil
	case context.ValueTypeMap:
		node, err := mapElement(expr, values)
		if err != nil {
			return err
		}
		ctx.Yield(node)
		return nil
	case context.ValueTypeFunction:
		log.Printf("GOT EXPR FUNCTION: %v", expr)
		fn, err := derefFunc(ctx, expr.Function())
		if err != nil {
			return err
		}
		log.Printf("GOT DEREF FUNCTION: %v", fn)
		if fn.Type() == context.ValueTypeFunction {
			return execFunc(ctx, fn.Function(), values)
		}
		return execExpr(ctx, fn, values)
	case context.ValueTypeSymbol:
		fn, err := ctx.Get(expr.Symbol())
		log.Printf("GET NAME: %v, VALUE: %v", expr.Symbol(), fn)
		if err != nil {
			if err == context.ErrUndefinedFunction {
				return fmt.Errorf("undefined function %q", fn.Symbol())
			}
			return err
		}
		if fn.Type() == context.ValueTypeFunction {
			return execFunc(ctx, fn.Function(), values)
		}
		return execExpr(ctx, fn, values)
	}

	panic("reached")

	return fmt.Errorf("invalid expression type: %v", expr.Type())
}

func prepareFunc(values []*context.Value) *context.Value {
	return context.NewFunctionValue(func(ctx *context.Context) error {
		if len(values) < 1 {
			ctx.Yield(context.Nil)
			return nil
		}

		expr := values[0]
		/*
			if expr.Type() == context.ValueTypeFunction {
				var err error
				expr, err = derefFunc(ctx, expr)
				if err != nil {
					return err
				}
			}
		*/

		return execExpr(ctx, expr, values[1:])
	})
}

func evalContextList(ctx *context.Context, nodes []*ast.Node) error {
	for i := range nodes {
		if err := evalContext(ctx, nodes[i]); err != nil {
			return err
		}
	}
	return nil
}

func newErrorMap(err error) *context.Value {
	k := context.NewAtomValue(":error")
	v := context.NewSymbolValue(err.Error())
	return context.NewMapValue(map[context.Value]*context.Value{*k: v})
}

func runtimeError(ctx *context.Context, n *ast.Node, err error) error {
	if n == nil {
		log.Printf("runtime error: %v", err)
		ctx.Yield(newErrorMap(err))
		ctx.Exit(err)
		return nil
	}
	tok := n.Token()
	line, col := tok.Pos()
	log.Printf("runtime error: %v (near %v, line: %v, col: %v)",
		err,
		fmt.Sprintf("[%s]", string(ast.Encode(n.Parent()))),
		line,
		col)
	ctx.Yield(newErrorMap(err))
	ctx.Exit(err)
	return nil
}

func evalContext(ctx *context.Context, n *ast.Node) error {

	if n.IsValue() {
		value, err := context.NewValue(n)
		if err != nil {
			return err
		}
		return ctx.Yield(value)
	}

	switch n.Type() {

	case ast.NodeTypeList:
		newCtx := context.New(ctx).Name("list")

		fnErr := make(chan error, 1)
		go func() {
			defer newCtx.Exit(nil)
			fnErr <- evalContextList(newCtx, n.List())
		}()

		value, err := newCtx.Results()
		if err != nil {
			return err
		}
		ctx.Yield(value)

		if err := <-fnErr; err != nil {
			return runtimeError(ctx, n, err)
		}

		return nil

		return ctx.Yield(value)

	case ast.NodeTypeMap:

		newCtx := context.NewClosure(ctx).Name("map")
		go func() error {
			defer newCtx.Exit(nil)

			err := evalContextList(newCtx, n.List())
			if err != nil {
				return err
			}
			return nil
		}()

		result := map[context.Value]*context.Value{}
		var key *context.Value
		for {
			value, err := newCtx.Output()
			if err != nil {
				if err == context.ErrClosedChannel {
					value := context.NewMapValue(result)
					return ctx.Yield(value)
				}
				return err
			}
			if key == nil {
				key = value
				result[*key] = context.Nil
			} else {
				result[*key] = value
				key = nil
			}
		}

		panic("unreachable")

	case ast.NodeTypeExpression:

		newCtx := context.NewClosure(ctx).Name("expr-eval").NonExecutable()

		fnErr := make(chan error, 1)
		go func() {
			defer newCtx.Exit(nil)
			fnErr <- evalContextList(newCtx, n.List())
		}()

		values, err := newCtx.Results()
		if err != nil {
			return err
		}

		if err := <-fnErr; err != nil {
			return runtimeError(ctx, n, err)
		}

		fn := prepareFunc(values.List())

		if ctx.IsExecutable() {
			execCtx := context.New(ctx).Name("expr-exec")

			fnErr := make(chan error, 1)
			go func() {
				defer execCtx.Exit(nil)
				fnErr <- fn.Function().Exec(execCtx)
			}()

			values, err := execCtx.Results()
			if err != nil {
				return runtimeError(ctx, n, err)
			}

			if err := <-fnErr; err != nil {
				return runtimeError(ctx, n, err)
			}

			if len(values.List()) == 1 {
				return ctx.Yield(values.List()[0])
			}

			return ctx.Yield(values)
		}

		return ctx.Yield(fn)
	}

	panic("unreachable")
}

func eval(node *ast.Node) (*context.Context, []*context.Value, error) {
	newCtx := context.New(defaultContext).Name("eval")

	fnErr := make(chan error, 1)
	go func() {
		defer newCtx.Exit(nil)
		fnErr <- evalContext(newCtx, node)
	}()

	values, err := newCtx.Collect()
	if err != nil {
		return nil, nil, err
	}

	if err := <-fnErr; err != nil {
		log.Printf("values: %v", values)
		return nil, nil, err
	}

	if len(values) == 0 {
		return newCtx, nil, nil
	}

	//if len(values) == 1 {
	//	return newCtx, values[0], nil
	//}

	return newCtx, values, nil
}

func mapElement(value *context.Value, path []*context.Value) (*context.Value, error) {
	for i := range path {
		k := *path[i]
		if value.Type() == context.ValueTypeMap {
			_, ok := value.Map()[k]
			if !ok {
				return context.Nil, nil
			}
			value = value.Map()[k]
		} else {
			return context.Nil, nil
		}
	}

	return value, nil
}

func mapListItem(value *context.Value, path []*context.Value) (*context.Value, error) {
	for i := range path {
		k := *path[i]
		if k.Type() != context.ValueTypeInt {
			return context.Nil, nil
		}
		if value.Type() == context.ValueTypeList {
			list := value.List()
			if k.Int() >= int64(len(value.List())) {
				return context.Nil, nil
			}
			value = list[k.Int()]
		} else {
			return context.Nil, nil
		}
	}

	return value, nil
}
