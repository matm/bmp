//go:build openbsd || darwin

package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/matm/bmp/pkg/types"
)

type basicPrompt struct {
	r *bufio.Reader
}

func (p *basicPrompt) Input() string {
	fmt.Printf("> ")
	ch, _, _ := p.r.ReadLine()
	return string(ch)
}

func newPrompt() types.Prompter {
	r := bufio.NewReader(os.Stdin)
	return &basicPrompt{r: r}
}
