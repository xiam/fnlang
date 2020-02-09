package fnlang_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xiam/fnlang"
	_ "github.com/xiam/fnlang/stdlib"
	"github.com/xiam/sexpr/ast"
	"github.com/xiam/sexpr/parser"
)

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
			In:  `([1])`,
			Out: `[[1]]`,
		},
		{
			In:  `((1))`,
			Out: `[1]`,
		},
		{
			In:  `(((1)))`,
			Out: `[1]`,
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
											(echo :hello)
											(set x 1)
											(get x)
											(x)
										]
								)
								(foo)
								`,
			Out: `[:true [:hello :true 1 1]]`,
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
									 (= n 0) 0
									 (= n 1) 1
									 :true (
										 +
										 (fib (- n 1))
										 (fib (- n 2))
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
		{
			In: `
							 (defn fib [n]
								 (when
									 (= n 0) 0
									 (= n 1) 1
									 (
										 +
										 (fib (- n 1))
										 (fib (- n 2))
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
		{
			In: `
							(defn square [x] (* x x))
							(square 20)
						`,
			Out: `[:true 400]`,
		},
		{
			In: `
							(defn factorial [n]
								(when
									(= n 0) 1
									(
										* n (factorial (- n 1))
									)
								)
							)
							(factorial 5)
						`,
			Out: `[:true 120]`,
		},
		{
			In: `
							(defn F [n]
								(when
									(= n 0) 1
								)
							)
							(F 5)
						`,
			Out: `[:true :nil]`,
		},
		{
			In: `
							(set x {:a 1})
							(get x)
						`,
			Out: `[:true {:a 1}]`,
		},
		{
			In: `
							(set x [1		 2 3 4 [	5	 6]])
							(get x)
						`,
			Out: `[:true [1 2 3 4 [5 6]]]`,
		},
		{
			In: `
							(set x {:a 1 :b 1.23})
							(x)
						`,
			Out: `[:true {:a 1 :b 1.23}]`,
		},
		{
			In: `
							(set x (echo :hello))
							(x)
						`,
			Out: `[:true :hello]`,
		},
		{
			In: `
							(set x {
								:a 1
								:b {:a 2}
								:c 3
								:f [ 1 2	[4 5]]
							})
							(x)
							(x :a)
							(x :b :a)
							(x :c)
							(x :d)
							(x :a :b :c)
							(x :a :b)
							(x :f)
						`,
			Out: `[:true {:a 1 :b {:a 2} :c 3 :f [1 2 [4 5]]} 1 2 3 :nil :nil :nil [1 2 [4 5]]]`,
		},
		{
			In: `
							(set fib [
								0
								1
								1
								2
								3
								5
								8
							])
							(fib)
							(fib 0)
							(fib 3)
							(fib 5)
						`,
			Out: `[:true [0 1 1 2 3 5 8] 0 2 5]`,
		},
		{
			In: `
							(set x [
								0
								[ 5 23 [ 7 4 ] ]
								[23 5 [45] [22] 33 45]
							])
							(x)
							(x 0)
							(x 9)
							(x 1 0)
							(x 1 1)
							(x 1 2)
							(x 1 2 0)
							(x 2 2)
							(x 2 2 0)
						`,
			Out: `[:true [0 [5 23 [7 4]] [23 5 [45] [22] 33 45]] 0 :nil 5 23 [7 4] 7 [45] 45]`,
		},
		{
			In: `
							(set x {
								:a 1
								:b {:a 2}
								:f [ 1 2 [4 5]]
							})
							(x :a)
							(x :b :a)
							(x :f 1)
							((x :f) 2)
							((x :f) 2 1)
							((x :f) 22)
						`,
			Out: `[:true 1 2 :nil [4 5] 5 :nil]`,
		},
		{
			In: `
							(set obj {
								:a "Hello world!"
								:b (echo :hi)
								:c (fn [x] (* x x))
								:d (fn [x] (echo x))
							})
							(obj :a)
							(obj :b)
							((obj :c) 100)
							((obj :c) 10)
							((obj :d) :hullo)
						`,
			Out: `[:true "Hello world!" :hi 10000 100 :hullo]`,
		},
		{
			In: `
						(set arr [
							(echo "Hello world!")
							(echo :hi)
							(fn [x] (* x x))
							(fn [x] (echo x))
						])
						(arr 0)
						(arr 1)
						((arr 2) 100)
						((arr 2) 10)
						((arr 3) 10)
					`,
			Out: `[:true "Hello world!" :hi 10000 100 10]`,
		},
		{
			In: `
							(set square (fn [x] (* x x)))
						`,
			Out: `[:true]`,
		},
		{
			In: `
							(set hello (echo :hello))
							(hello)
						`,
			Out: `[:true :hello]`,
		},
		{
			In: `
							((fn [x] (* x x)) 100)
						`,
			Out: `[10000]`,
		},
		{
			In: `
							(echo :hello)
							(set square (fn [x] (* x x)))
							(square 100)
						`,
			Out: `[:hello :true 10000]`,
		},
		{
			In: `
							(defn square [x] (* x x))
							(square 100)
						`,
			Out: `[:true 10000]`,
		},
		{
			In: `
							(set square (fn [x] (* x x)))
							(set squareA (fn [y] (square y)))
							(square 10)
							(squareA 100)
						`,
			Out: `[:true :true 100 10000]`,
		},
		{
			In: `
						# A function that returns another function.
						(defn squarefn [] (fn [y] (* y y)))
						# And gets executed.
						((squarefn) 100)
						`,
			Out: `[:true 10000]`,
		},
		{
			In: `
						(set square (fn [y] (* y y)))
						(square 100)
						`,
			Out: `[:true 10000]`,
		},
		{
			In: `
						(set squarefn
							(fn []
								(fn [y]
									(* y y)
								)
							)
						)
						((squarefn) 100)
						`,
			Out: `[:true 10000]`,
		},
		{
			In: `
					(defn square [x] (* x x))
					(assert (square 100) 10000)
					(assert (square 10) 10000)
					(assert :true)
					(assert :false)
					(assert)
					`,
			Out: `[:true :true :false :true :false :nil]`,
		},
		{
			In: `
					(set arr [])
					(push arr 1)
					(push arr 2)
					(push arr 3 4 5)
					(arr)
					`,
			Out: `[:true :true :true :true [1 2 3 4 5]]`,
		},
		{
			In: `
				(+ 3 2 4 5)
				(- 1 2 4 5)
				(+ 0.2 0.3 0.41 4 0.1 5)
				(- 0.2 0.3 0.41 4 0.1 5)
				(* 2 3 4)
				(* -2.2 3 6 -3 -4.6)
				(/ 6 3)
				(/ 6.6 3.3)
				(/ 6 33)
				(/ 6.0 33)
				(/ 6 33.0)
				`,
			Out: `[14 -10 10.01 -9.61 24 -546.48 2 2 0 0.18181818181818182 0.18181818181818182]`,
		},
	}
	for i := range testCases {
		root, err := parser.Parse([]byte(testCases[i].In))
		assert.NoError(t, err)
		assert.NotNil(t, root)

		ast.Print(root)

		_, result, err := fnlang.Eval(root)
		assert.NoError(t, err)

		assert.Equal(t, testCases[i].Out, result[0].String())
	}
}
