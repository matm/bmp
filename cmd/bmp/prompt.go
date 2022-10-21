//go:build !(openbsd || darwin)

package main

import (
	"github.com/c-bata/go-prompt"
	"github.com/matm/bmp/pkg/types"
)

func completer(d prompt.Document) []prompt.Suggest {
	return nil
}

func executor(cmd string) {
	return
}

type advancedPrompt struct {
	pr *prompt.Prompt
}

func (p *advancedPrompt) Input() string {
	return p.pr.Input()
}

func newPrompt() types.Prompter {
	p := prompt.New(executor, completer, prompt.OptionHistory([]string{}))
	return &advancedPrompt{pr: p}
}
