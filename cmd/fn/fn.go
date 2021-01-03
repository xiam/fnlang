package main

import (
	"fmt"
	"github.com/xiam/fnlang"
	_ "github.com/xiam/fnlang/stdlib"
	//"github.com/xiam/sexpr/ast"
	"github.com/xiam/sexpr/parser"
	"io"
	"log"
)

func main() {
	rl := newReadLiner()
	defer rl.Close()

	quitErr := make(chan error)

	r, w := io.Pipe()

	go func() {
		root, err := parser.NewStreamer(r)
		//log.Printf("ROOT: %v, err: %v", root, err)
		//ast.Print(root)
		if err != nil {
			quitErr <- err
			return
		}
		_, res, err := fnlang.Eval(root)
		if err != nil {
			quitErr <- err
			return
		}
		fmt.Println(res[0].String())
		quitErr <- nil
	}()

	for {
		s, err := rl.Prompt()
		if err != nil {
			w.CloseWithError(err)
			break
		}
		_, err = w.Write([]byte(s + "\n"))
		if err != nil {
			w.CloseWithError(err)
			break
		}
	}

	err := <-quitErr
	if err != nil {
		log.Fatalf("fn: %v", err)
	}
}
