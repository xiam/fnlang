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
	case context.ValueTypeFunction:
		newCtx := context.NewClosure(ctx).Name("exec-body")
		go func() error {
			defer newCtx.Exit(nil)
			return body.Function().Exec(newCtx)
		}()
		values, err := newCtx.Results()
		if err != nil {
			return err
		}
		ctx.Yield(values.List()...)
		return nil
	case context.ValueTypeList:
		for _, item := range body.List() {
			if err := execFunctionBody(ctx, item); err != nil {
				return err
			}
		}
		return nil
	default:
		panic("unhandled")
	}
}

func prepareFunc(values []*context.Value) *context.Value {
	return context.NewFunctionValue(func(ctx *context.Context) error {

		fn := values[0]

		if len(values) == 1 {
			switch fn.Type() {
			case context.ValueTypeInt, context.ValueTypeAtom, context.ValueTypeList, context.ValueTypeString, context.ValueTypeMap:
				ctx.Yield(fn)
				return nil
			}
		}

		if fn.Type() == context.ValueTypeFunction {
			var err error
			fn, err = context.ExecArgument(ctx, fn)
			if err != nil {
				return err
			}
		}

		switch fn.Type() {
		case context.ValueTypeFunction, context.ValueTypeList, context.ValueTypeAtom, context.ValueTypeInt:
			ctx.Yield(fn)
			return nil
		}

		fnName := fn.Symbol()
		fn, err := ctx.Get(fnName)
		if err != nil {
			if err == context.ErrUndefinedFunction {
				log.Fatalf("undefined function %q", fnName)
				return fmt.Errorf("undefined function %q", fnName)
			}
			return err
		}

		//fnCtx := context.NewClosure(ctx).Executable()

		go func() {
			defer ctx.Close()

			for i := 1; i < len(values) && ctx.Accept(); i++ {
				//fnCtx.
				ctx.Push(values[i])
			}
		}()

		/*
			go func() {
				//defer fnCtx.Exit(nil)

			}()
		*/
		return fn.Function().Exec(ctx)

		/*
			result, err := fnCtx.Result()
			if err != nil {
				log.Printf("err.res: %v", err)
				return err
			}
			for i := 0; i < len(result.List()); i++ {
				ctx.Yield(result.List()[i])
			}

			return nil
		*/

	})
}

func evalContextList(ctx *context.Context, nodes []*ast.Node) error {
	for i := range nodes {
		err := evalContext(ctx, nodes[i])
		if err != nil {
			//ctx.outErr <- err
			return err
		}
	}

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
		go func() {
			defer newCtx.Exit(nil)
			err := evalContextList(newCtx, n.List())
			if err != nil {
				return
			}
		}()

		value, err := newCtx.Results()
		if err != nil {
			return err
		}

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
		go func() error {
			defer newCtx.Exit(nil)

			err := evalContextList(newCtx, n.List())
			if err != nil {
				return err
			}

			return nil
		}()

		values, err := newCtx.Results()
		if err != nil {
			return err
		}

		fn := prepareFunc(values.List())

		if ctx.IsExecutable() {
			execCtx := context.New(ctx).Name("expr-exec")

			go func() {
				defer execCtx.Exit(nil)

				if err := fn.Function().Exec(execCtx); err != nil {
					log.Fatalf("ERR: %v", err)
				}
			}()

			values, err := execCtx.Results()
			if err != nil {
				return err
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

	go func() {
		defer newCtx.Exit(nil)

		if err := evalContext(newCtx, node); err != nil {
			log.Fatalf("EVAL.CONTEXT: %v", err)
			return
		}
	}()

	values, err := newCtx.Collect()
	if err != nil {
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
