package main

import (
	"errors"
	"fmt"
	"github.com/xiam/fnlang"
	_ "github.com/xiam/fnlang/stdlib"
	"github.com/xiam/sexpr/parser"
	"log"
	"strings"
)

func indentation(line string) string {
	var indent string
	for _, c := range line {
		if c == ' ' || c == '\t' {
			indent += string(c)
		}
	}
	return indent
}

func main() {
	rl := newReadLiner()
	defer rl.Close()

	s := fnlang.New()

	var buf string // TODO: use strings.Builder
	for {

		line, err := rl.Prompt()
		if err != nil {
			log.Printf("prompt: %v", err)
			break
		}

		buf += line + "\n"
		values, err := s.Eval(strings.NewReader(buf))
		if err != nil {
			if errors.Is(err, parser.ErrUnexpectedEOF) {
				// TODO: add indentation suggestion
				rl.SetPrompt("... ")
				continue
			}
			fmt.Println("!!! %v", err)
			continue
		}
		fmt.Printf("--> %v\n", values[0])

		buf = ""
		rl.SetPrompt(prompt)
	}

	log.Printf("VALUES: %v", s.Values())
}
