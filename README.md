# fn (WIP)

*fn* ([IPA][1]: /ɛfɛn/) is a functional programming language, a dialect of Lisp
built on top of the [Go](https://golang.org) programming language.

*fn* is still in design phase.

## Early testing

Install the CLI interpreter and feed it with some examples:

```sh
go install github.com/xiam/fnlang/cmd/fn

fn < _examples/001-hello-world.fn
# Hello world!
# [[:nil]]

cat _examples/003-square.fn
# (defn square [x] (* x x))
# (square 10)
# (square 100)
fn < _examples/003-square.fn
# [[:true 100 10000 1000000]]
```

### Examples

#### Fibonacci numbers

```lisp
(defn fib [n]
	(when
		(= n 0) 0
		(= n 1) 1
		:true (+ (fib (- n 1)) (fib (- n 2)))
	)
)

(fib 0)
(fib 1)
(fib 2)
(fib 3)
(fib 4)
(fib 5)
(fib 6)
```

```
[[:true 0 1 1 2 3 5 8]]
```

## Built-in types

### Atom

### Numeric

#### Integer

#### Float

### String

### List

### Expression

### Map

## License

## Functions

## Error handling

## Concurrency

This project is licensed under the terms of the **MIT License**.

> Copyright (c) 2019-present. José Nieto. All rights reserved.
>
> Permission is hereby granted, free of charge, to any person obtaining
> a copy of this software and associated documentation files (the
> "Software"), to deal in the Software without restriction, including
> without limitation the rights to use, copy, modify, merge, publish,
> distribute, sublicense, and/or sell copies of the Software, and to
> permit persons to whom the Software is furnished to do so, subject to
> the following conditions:
>
> The above copyright notice and this permission notice shall be
> included in all copies or substantial portions of the Software.
>
> THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
> EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
> MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
> NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
> LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
> OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
> WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

[1]: https://en.wiktionary.org/wiki/Wiktionary:International_Phonetic_Alphabet
