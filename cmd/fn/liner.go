package main

import (
	"github.com/peterh/liner"
)

const (
	prompt = "fn> "
)

type readLiner struct {
	*liner.State
	prompt string
}

func newReadLiner() *readLiner {
	state := liner.NewLiner()
	state.SetCtrlCAborts(true)
	return &readLiner{State: state, prompt: prompt}
}

func (rl *readLiner) Prompt() (string, error) {
	return rl.State.Prompt(rl.prompt)
}

func (rl *readLiner) SetPrompt(prompt string) {
	rl.prompt = prompt
}
