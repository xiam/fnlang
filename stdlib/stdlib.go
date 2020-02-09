package stdlib

import (
	"errors"
	"fmt"
	"log"

	"github.com/xiam/fnlang"
	"github.com/xiam/fnlang/context"
)

func execFunctionBody(ctx *context.Context, body *context.Value) error {
	switch body.Type() {
	case context.ValueTypeInt:
		ctx.Yield(body)
		return nil
	case context.ValueTypeFunction:
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
			return fnlang.RuntimeError(ctx, body.Node(), err)
		}
		return nil
	case context.ValueTypeList:
		newCtx := context.New(ctx).Name("exec-list")
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

func init() {

	fnlang.Defn("when", func(ctx *context.Context) error {
		for {
			if !ctx.Next() {
				break
			}
			cond, err := ctx.Argument()
			if err != nil {
				return err
			}
			if ctx.Next() {
				if context.Eq(cond, context.True) {
					value, err := ctx.Argument()
					if err != nil {
						return err
					}
					ctx.Yield(value)
					return nil
				}
			} else {
				ctx.Yield(cond)
				return nil
			}
		}

		ctx.Yield(context.Nil)
		return nil
	})

	fnlang.Defn("push", func(ctx *context.Context) error {
		var name *context.Value
		var err error

		if ctx.NonExecutable().Next() {
			name, err = ctx.Argument()
			if err != nil {
				return err
			}
		}
		if name == nil {
			return errors.New("required symbol")
		}

		value, err := ctx.Get(name.Symbol())
		if err != nil {
			ctx.Yield(context.Nil)
			return nil
		}

		list := value.List()
		for ctx.Executable().Next() {
			value, err := ctx.Argument()
			if err != nil {
				return err
			}
			list = append(list, value)
			ctx.Parent.Set(name.Symbol(), context.NewListValue(list))
		}

		ctx.Yield(context.True)
		return nil
	})

	fnlang.Defn("-", func(ctx *context.Context) error {
		result := (interface{})(int64(0))
		for i := 0; ctx.Next(); i++ {
			value, err := ctx.Argument()
			if err != nil {
				return err
			}
			if i < 1 {
				result = value.Numeric()
				continue
			}

			if value.Type() == context.ValueTypeFloat {
				if _, ok := result.(float64); !ok {
					result = float64(result.(int64))
				}
			}
			switch result.(type) {
			case float64:
				result = result.(float64) - value.Float()
			case int64:
				result = result.(int64) - value.Int()
			}
		}
		ctx.Yield(context.NewNumericValue(result))
		return nil
	})

	fnlang.Defn("+", func(ctx *context.Context) error {
		result := (interface{})(int64(0))
		for ctx.Next() {
			value, err := ctx.Argument()
			if err != nil {
				return err
			}
			if value.Type() == context.ValueTypeFloat {
				if _, ok := result.(float64); !ok {
					result = float64(result.(int64))
				}
			}
			switch result.(type) {
			case float64:
				result = result.(float64) + value.Float()
			case int64:
				result = result.(int64) + value.Int()
			}
		}
		ctx.Yield(context.NewNumericValue(result))
		return nil
	})

	fnlang.Defn("/", func(ctx *context.Context) error {
		result := (interface{})(nil)
		for ctx.Next() {
			value, err := ctx.Argument()
			if err != nil {
				return err
			}
			if result == nil {
				result = value.Numeric()
				continue
			}
			if value.IsFloat() {
				if _, ok := result.(float64); !ok {
					result = float64(result.(int64))
				}
			}
			switch result.(type) {
			case float64:
				result = result.(float64) / value.Float()
			case int64:
				result = result.(int64) / value.Int()
			}
		}
		if result == nil {
			return errors.New("wrong number of arguments")
		}
		ctx.Yield(context.NewNumericValue(result))
		return nil
	})

	fnlang.Defn("*", func(ctx *context.Context) error {
		result := (interface{})(int64(1))
		for ctx.Next() {
			value, err := ctx.Argument()
			if err != nil {
				return err
			}
			if value.Type() == context.ValueTypeFloat {
				if _, ok := result.(float64); !ok {
					result = float64(result.(int64))
				}
			}
			switch result.(type) {
			case float64:
				result = result.(float64) * value.Float()
			case int64:
				result = result.(int64) * value.Int()
			}
		}
		ctx.Yield(context.NewNumericValue(result))
		return nil
	})

	fnlang.Defn(":false", func(ctx *context.Context) error {
		ctx.Yield(context.False)
		return nil
	})

	fnlang.Defn(":true", func(ctx *context.Context) error {
		for ctx.Next() {
			_, err := ctx.Argument()
			if err != nil {
				return err
			}
		}

		ctx.Yield(context.True)
		return nil
	})

	fnlang.Defn("echo", func(ctx *context.Context) error {
		for ctx.Next() {
			value, err := ctx.Argument()
			if err != nil {
				return err
			}
			ctx.Yield(value)
		}

		return nil
	})

	fnlang.Defn("=", func(ctx *context.Context) error {
		var first *context.Value
		for ctx.Next() {
			value, err := ctx.Argument()
			if err != nil {
				return err
			}

			if first == nil {
				first = value
				continue
			}

			if !context.Eq(first, value) {
				ctx.Yield(context.False)
				return nil
			}
		}

		ctx.Yield(context.True)
		return nil
	})

	fnlang.Defn("nop", func(ctx *context.Context) error {
		ctx.Yield(context.Nil)

		return nil
	})

	fnlang.Defn("fn", func(ctx *context.Context) error {
		var params, body *context.Value

		ctx = ctx.NonExecutable()
		for i := 0; ctx.Next(); i++ {
			arg, err := ctx.Argument()
			if err != nil {
				return err
			}

			switch i {
			case 0:
				params = arg
			default:
				body = arg
			}
		}

		paramsList := params.List()

		wrapperFn := context.NewFunctionValue(func(ctx *context.Context) error {
			for i := 0; ctx.Next() && i < len(paramsList); i++ {
				arg, err := ctx.Argument()
				if err != nil {
					return err
				}
				ctx.Set(paramsList[i].Symbol(), arg)
			}

			return execFunctionBody(ctx, body)
		})

		ctx.Yield(wrapperFn)

		return nil
	})

	fnlang.Defn("defn", func(ctx *context.Context) error {
		var name, params, body *context.Value

		ctx = ctx.NonExecutable()
		for i := 0; ctx.Next(); i++ {
			arg, err := ctx.Argument()
			if err != nil {
				return err
			}

			switch i {
			case 0:
				name = arg
			case 1:
				params = arg
			default:
				body = arg
			}
		}

		if params == nil {
			return errors.New("missing parameters list")
		}
		paramsList := params.List()

		wrapperFn := context.NewFunctionValue(func(ctx *context.Context) error {

			for i := 0; ctx.Next() && i < len(paramsList); i++ {
				arg, err := ctx.Argument()
				if err != nil {
					return err
				}
				ctx.Set(paramsList[i].Symbol(), arg)
			}

			return execFunctionBody(ctx, body)
		})
		wrapperFn.SetNode(body.Node())

		if err := ctx.Parent.Set(name.Symbol(), wrapperFn); err != nil {
			return err
		}

		ctx.Yield(context.True)
		return nil
	})

	fnlang.Defn("assert", func(ctx *context.Context) error {
		for ctx.Next() {
			var v1, v2 *context.Value
			var err error

			v1, err = ctx.Argument()
			if err != nil {
				return nil
			}

			v2 = context.True
			if ctx.Next() {
				v2, err = ctx.Argument()
				if err != nil {
					return err
				}
			}

			if context.Eq(v1, v2) {
				ctx.Yield(context.True)
				return nil
			} else {
				ctx.Yield(context.False)
				return nil
			}
		}

		ctx.Yield(context.Nil)
		return nil
	})

	fnlang.Defn("print", func(ctx *context.Context) error {
		for ctx.Next() {
			value, err := ctx.Argument()
			if err != nil {
				return err
			}
			fmt.Printf("%s", value.Symbol())
		}

		ctx.Yield(context.Nil)
		return nil
	})

	fnlang.Defn("get", func(ctx *context.Context) error {
		var name *context.Value
		ctx = ctx.NonExecutable()
		for i := 0; ctx.Next(); i++ {
			argument, err := ctx.Argument()
			if err != nil {
				return err
			}
			switch i {
			case 0:
				name = argument
			default:
				return errors.New("expecting one argument")
			}
		}

		value, err := ctx.Get(name.Symbol())
		if err != nil {
			ctx.Yield(context.Nil)
			return nil
		}

		ctx.Yield(value)
		return nil
	})

	fnlang.Defn("set", func(ctx *context.Context) error {
		var name, value *context.Value
		ctx = ctx.NonExecutable()
		for i := 0; ctx.Next(); i++ {
			if i > 0 {
				ctx = ctx.Executable()
			}
			argument, err := ctx.Argument()
			if err != nil {
				return err
			}
			switch i {
			case 0:
				name = argument
			case 1:
				value = argument
			default:
				return errors.New("expecting two arguments")
			}
		}
		if name == nil {
			return errors.New("fn set requires an argument")
		}
		if value == nil {
			value = context.Nil
		}

		ctx.Parent.Set(name.Symbol(), value)
		ctx.Yield(context.True)

		return nil
	})

	fnlang.Defn(":error", func(ctx *context.Context) error {
		for ctx.Next() {
			value, err := ctx.Argument()
			if err != nil {
				return err
			}
			return errors.New(value.String())
		}
		return nil
	})

}
