package main

import (
	"bytes"
	"fmt"
	"github.com/xiam/fnlang"
	_ "github.com/xiam/fnlang/stdlib"
	"github.com/xiam/sexpr/parser"
	"io"
	"log"
	"os"
)

func main() {
	buf := bytes.NewBuffer(nil)
	_, err := io.Copy(buf, os.Stdin)
	if err != nil {
		log.Fatal("io.Copy: ", err)
	}
	root, err := parser.Parse(buf.Bytes())
	if err != nil {
		log.Fatal("parser.Parse: ", err)
	}
	_, result, err := fnlang.Eval(root)
	if err != nil {
		log.Fatal("fnlang.Eval: ", err)
	}
	fmt.Printf("%s\n", result)
}
