package fnlang

import (
	"fmt"
	//	"log"

	//"strings"
	"errors"
	"testing"

	//"time"

	"github.com/stretchr/testify/assert"
	"github.com/xiam/fnlang/context"
	"github.com/xiam/sexpr/ast"
	"github.com/xiam/sexpr/parser"
)

func compileValue(in interface{}) ([]byte, error) {
	var buf string
	switch v := in.(type) {
	case *context.Value:
		s := v.String()
		return []byte(s), nil
	case []*context.Value:
		buf = buf + "["
		for i := range v {
			if i > 0 {
				buf = buf + ", "
			}
			chunk, err := compileValue(v[i])
			if err != nil {
				return nil, err
			}
			buf = buf + string(chunk)
		}
		buf = buf + "]"
	default:
		return nil, fmt.Errorf("unknown type: %T", in)
	}
	return []byte(buf), nil
}

func TestParserEvaluate(t *testing.T) {
	testCases := []struct {
		In  string
		Out string
	}{
		{
			In:  `1`,
			Out: `[1]`,
		},
		{
			In:  `1 2 3`,
			Out: `[1 2 3]`,
		},
		{
			In:  `[]`,
			Out: `[[]]`,
		},
		{
			In:  `[1]`,
			Out: `[[1]]`,
		},
		{
			In: `[ 3 2	1 ]`,
			Out: `[[3 2 1]]`,
		},
		{
			In: `[	1			 2 [ 4 5 [6 7 8]] 3]`,
			Out: `[[1 2 [4 5 [6 7 8]] 3]]`,
		},
		{
			In:  `{}`,
			Out: `[{}]`,
		},
		{
			In:  `{:a}`,
			Out: `[{:a :nil}]`,
		},
		{
			In: `{ :a 1		 }`,
			Out: `[{:a 1}]`,
		},
		{
			In:  `{:a 1 :b 2 :c 3 :e [1 2 3]}`,
			Out: `[{:a 1 :b 2 :c 3 :e [1 2 3]}]`,
		},
		{
			In:  `[{:a 1 :b 2 :c 3 :e [1 2 3]} [1 2 3] 4 :foo]`,
			Out: `[[{:a 1 :b 2 :c 3 :e [1 2 3]} [1 2 3] 4 :foo]]`,
		},
		{
			In:  `(1)`,
			Out: `[1]`,
		},
		{
			In:  `(((1)))`,
			Out: `[1]`,
		},
		{
			In:  `([1])`,
			Out: `[[1]]`,
		},
		{
			In:  `([[1]])`,
			Out: `[[[1]]]`,
		},
		{
			In:  `[([1])]`,
			Out: `[[[1]]]`,
		},
		{
			In: `( [1	2	3 ] )`,
			Out: `[[1 2 3]]`,
		},
		{
			In:  `(:nil)`,
			Out: `[:nil]`,
		},
		{
			In:  `(:hello)`,
			Out: `[:hello]`,
		},
		{
			In:  `(([1 2 3 {:a 4}]))`,
			Out: `[[1 2 3 {:a 4}]]`,
		},
		{
			In:  `[(nop [ [ (echo :hello) ]])]`,
			Out: `[[:nil]]`,
		},
		{
			In:  `[(print "hello " "world!")]`,
			Out: `[[:nil]]`,
		},
		{
			In:  `(echo "foo" "bar")`,
			Out: `[["foo" "bar"]]`,
		},
		{
			In:  `(["foo" "bar"])`,
			Out: `[["foo" "bar"]]`,
		},
		{
			In:  `([["foo" "bar"]])`,
			Out: `[[["foo" "bar"]]]`,
		},
		{
			In:  `((([["foo" "bar"]])))`,
			Out: `[[["foo" "bar"]]]`,
		},
		{
			In:  `(print "hello world!" " beautiful world!")`,
			Out: `[:nil]`,
		},
		{
			In: `(echo "hello world!" "beautiful world!"	1		2 )`,
			Out: `[["hello world!" "beautiful world!" 1 2]]`,
		},
		{
			In:  `(10)`,
			Out: `[10]`,
		},
		{
			In:  `(+ 1 2 3 4)`,
			Out: `[10]`,
		},
		{
			In:  `(+ (+ 1 2 3 4))`,
			Out: `[10]`,
		},
		{
			In:  `(+ (+ 1 2 3 4) 10)`,
			Out: `[20]`,
		},
		{
			In:  `(= 2 3)`,
			Out: `[:false]`,
		},
		{
			In:  `(= 1 1)`,
			Out: `[:true]`,
		},
		{
			In:  `(= 1 1 1 1 1 1 1)`,
			Out: `[:true]`,
		},
		{
			In:  `(= 1 1 1 1 1 2 14)`,
			Out: `[:false]`,
		},
		{
			In:  `(set foo 1)`,
			Out: `[:true]`,
		},
		{
			In:  `(get foo)`,
			Out: `[:nil]`,
		},
		{
			In:  `(get foo) (set foo 3) (get foo) (get foo)`,
			Out: `[:nil :true 3 3]`,
		},
		{
			In:  `(echo (set foo 1) (get foo))`,
			Out: `[[:true 1]]`,
		},
		{
			In:  `(echo "hello" "world!")`,
			Out: `[["hello" "world!"]]`,
		},
		{
			In:  `(echo "hello" (echo "world!"))`,
			Out: `[["hello" "world!"]]`,
		},
		{
			In:  `(echo "hello" (echo (echo (echo "world!"))))`,
			Out: `[["hello" "world!"]]`,
		},
		{
			In:  `(:true)`,
			Out: `[:true]`,
		},
		{
			In: `(	123 )`,
			Out: `[123]`,
		},
		{
			In:  `(:true :true)`,
			Out: `[:true]`,
		},
		{
			In:  `(:true :false :true :true :false)`,
			Out: `[:true]`,
		},
		{
			In:  `(:false)`,
			Out: `[:false]`,
		},
		{
			In:  `(:false :true :true)`,
			Out: `[:false]`,
		},
		{
			In:  `(:true "hello")`,
			Out: `[:true]`,
		},
		{
			In:  `(:true (echo "hello" (echo "world")))`,
			Out: `[:true]`,
		},
		{
			In:  `(:false (echo "hello" (echo "world")))`,
			Out: `[:false]`,
		},
		{
			In:  `(:true (echo "hello" "world!"))`,
			Out: `[:true]`,
		},
		{
			In:  `(echo "hello" (echo (echo (echo "world!"))))`,
			Out: `[["hello" "world!"]]`,
		},
		{
			In:  `(:true (echo "hello" (echo (echo (echo "world!")))))`,
			Out: `[:true]`,
		},
		{
			In:  `(:false (echo "hello" "world!"))`,
			Out: `[:false]`,
		},
		{
			In:  `(defn foo [word] (echo (get word)))`,
			Out: `[:true]`,
		},
		{
			In:  `(defn foo [word] (echo (get word))) (foo "HEY")`,
			Out: `[:true "HEY"]`,
		},
		{
			In:  `((= 1 2) 6 7 8 9)`,
			Out: `[:false]`,
		},
		{
			In:  `((= 1 1) 6 7 8 9)`,
			Out: `[:true]`,
		},
		{
			In:  `(when :false 6)`,
			Out: `[:nil]`,
		},
		{
			In:  `(when :true 6)`,
			Out: `[6]`,
		},
		{
			In: `
				 (when
					 :false
						 5
					 :false
						 3
					 :true
						 6
					 :false
						 4
					 :true
						 8
				 )`,
			Out: `[6]`,
		},
		{
			In: `
				 (when
					 (= 1 2)
						 5
					 :false
						 3
					 (:false)
						 3
					 (= 3 3)
						 6
					 (:false)
						 1
				 )`,
			Out: `[6]`,
		},
		{
			In: `
					 (defn F [n]
						 (when
							 (= (get n) 0) 0
							 (= (get n) 1) 1
							 :true 99
						 )
					 )
					 (F 0)
					 (F 1)
					 (F 2)
					 (F 3)
					 (F 4)
					 (F 5)
					 (F "a")
					 `,
			Out: `[:true 0 1 99 99 99 99 99]`,
		},
		{
			In: `
					 (defn F [n]
						 (when
							 (= (get n) 0) 0
							 (= (get n) 1) 1
							 (= (get n) 2) 3
							 :true (F 2)
						 )
					 )
					 `,
			Out: `[:true]`,
		},
		{
			In: `
					 (defn F [n]
						 (when
							 (= (get n) 0) 0
							 (= (get n) 1) 1
							 (= (get n) 2) 3
							 :true (F 2)
						 )
					 )
					 (F 0)
					 (F 1)
					 (F 2)
					 (F 3)
					 (F 4)
					 (F 5)
					 `,
			Out: `[:true 0 1 3 3 3 3]`,
		},
		{
			In: `
					(set x 1)
					(get x)
					[
						(get x)
						(set x 2)
						(get x)
					]
					(get x)
					(set x 6)
					(get x)
					[
						(get x)
						(set x 9)
						[(get x) (set x 10) (get x)]
						(get x)
					]
					(get x)
				`,
			Out: `[:true 1 [1 :true 2] 1 :true 6 [6 :true [9 :true 10] 9] 6]`,
		},
		{
			In: `
				(
					defn foo []
						[
							(
								echo :hello
							)
							(
								set x 1
							)
							(
								get x
							)
						]
				)
				(foo)
				`,
			Out: `[:true [:hello :true 1]]`,
		},
		{
			In: `
			(
				defn foo [] [
					(set x 1)
					(get x)
				]
			)
			(foo)
			`,
			Out: `[:true [:true 1]]`,
		},
		{
			In: `
			(set x 6)
			(
				defn foo [] [
					(set x 1)
					(get x)
				]
			)
			(get x)
			(foo)
			(get x)
			`,
			Out: `[:true :true 6 [:true 1] 6]`,
		},
		{
			In: `
			(set x 1)
			(get x)
			`,
			Out: `[:true 1]`,
		},
		{
			In: `
			 (defn F [n]
				 (when
					 (= (get n) 0) 0
					 (= (get n) 1) 1
					 :true 2
				 )
			 )
			 (F 0)
			 (F 1)
			 (F 2)
			 (F 3)
			 (F 4)
			 (F 5)
			 (F 6)`,
			Out: `[:true 0 1 2 2 2 2 2]`,
		},
		{
			In: `
			 (defn F [n]
				 (when
					 (= (get n) 0) 0
					 (= (get n) 1) 1
					 :true (F 1)
				 )
			 )
			 (F 0)
			 (F 1)
			 (F 2)
			 (F 3)
			 (F 4)
			 (F 5)
			 (F 6)`,
			Out: `[:true 0 1 1 1 1 1 1]`,
		},
		{
			In: `
			 (defn F [n]
				 (when
					 (= (get n) 0) 0
					 (= (get n) 1) 1
					 :true (+ (F 1) 1)
				 )
			 )
			 (F 0)
			 (F 1)
			 (F 2)
			 (F 3)
			 (F 4)
			 (F 5)
			 (F 6)`,
			Out: `[:true 0 1 2 2 2 2 2]`,
		},
		{
			In: `
			 (defn F [n]
				 (when
					 (= (get n) 0) 0
					 (= (get n) 1) 1
					 :true (+ (F 1) (F 1))
				 )
			 )
			 (F 0)
			 (F 1)
			 (F 2)
			 (F 3)
			 (F 4)
			 (F 5)
			 (F 6)`,
			Out: `[:true 0 1 2 2 2 2 2]`,
		},
		{
			In:  `(- 1 2)`,
			Out: `[-1]`,
		},
		{
			In:  `(- 1 1)`,
			Out: `[0]`,
		},
		{
			In:  `(- 10 1 1 1)`,
			Out: `[7]`,
		},
		{
			In: `
			 (defn F [n]
				 (when
					 (= (get n) 0) 0
					 (= (get n) 1) 1
					 :true (- (F 1) 1)
				 )
			 )
			 (F 0)
			 (F 1)
			 (F 2)
			 (F 3)
			 (F 4)
			 (F 5)
			 (F 6)`,
			Out: `[:true 0 1 0 0 0 0 0]`,
		},
		{
			In: `
			 (defn fib [n]
				 (when
					 (= (get n) 0) 0
					 (= (get n) 1) 1
					 :true (
						 +
						 (fib (- (get n) 1))
						 (fib (- (get n) 2))
					 )
				 )
			 )
			 (fib 0)
			 (fib 1)
			 (fib 2)
			 (fib 3)
			 (fib 4)
			 (fib 5)
			 (fib 6)
			 `,
			Out: `[:true 0 1 1 2 3 5 8]`,
		},
	}

	Defn("when", func(ctx *context.Context) error {
		for {
			if !ctx.Next() {
				break
			}
			cond, err := ctx.Argument()
			if err != nil {
				return err
			}
			if !ctx.Next() {
				return errors.New("missing value after condition")
			}

			if context.Eq(cond, context.True) {
				value, err := ctx.Argument()
				if err != nil {
					return err
				}
				ctx.Yield(value)
				return nil
			}
		}

		ctx.Yield(context.Nil)
		return nil
	})

	Defn("-", func(ctx *context.Context) error {
		result := int64(0)
		for i := 0; ctx.Next(); i++ {
			value, err := ctx.Argument()
			if err != nil {
				return err
			}
			if i < 1 {
				result = value.Int()
				continue
			}
			result -= value.Int()
		}
		ctx.Yield(context.NewIntValue(result))
		return nil
	})

	Defn("+", func(ctx *context.Context) error {
		result := int64(0)
		for ctx.Next() {
			value, err := ctx.Argument()
			if err != nil {
				return err
			}
			result += value.Int()
		}
		ctx.Yield(context.NewIntValue(result))
		return nil
	})

	Defn(":false", func(ctx *context.Context) error {
		ctx.Yield(context.False)
		return nil
	})

	Defn(":true", func(ctx *context.Context) error {
		for ctx.Next() {
			_, err := ctx.Argument()
			if err != nil {
				return err
			}
		}

		ctx.Yield(context.True)
		return nil
	})

	Defn("echo", func(ctx *context.Context) error {
		for ctx.Next() {
			value, err := ctx.Argument()
			if err != nil {
				return err
			}
			ctx.Yield(value)
		}

		return nil
	})

	Defn("=", func(ctx *context.Context) error {
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

			if (*first).Symbol() != value.Symbol() {
				ctx.Yield(context.False)
				return nil
			}
		}

		ctx.Yield(context.True)
		return nil
	})

	Defn("nop", func(ctx *context.Context) error {
		ctx.Yield(context.Nil)

		return nil
	})

	Defn("defn", func(ctx *context.Context) error {
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

		if err := ctx.Parent.Set(name.Symbol(), wrapperFn); err != nil {
			return err
		}

		ctx.Yield(context.True)
		return nil
	})

	Defn("print", func(ctx *context.Context) error {
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

	Defn("get", func(ctx *context.Context) error {
		var name *context.Value
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

	Defn("set", func(ctx *context.Context) error {
		var name, value *context.Value
		for i := 0; ctx.Next(); i++ {
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

	for i := range testCases {
		root, err := parser.Parse([]byte(testCases[i].In))
		assert.NoError(t, err)
		assert.NotNil(t, root)

		ast.Print(root)

		_, result, err := eval(root)
		assert.NoError(t, err)

		s, err := compileValue(result)
		assert.NoError(t, err)

		assert.Equal(t, testCases[i].Out, string(s))
	}
}
