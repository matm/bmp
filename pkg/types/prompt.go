package types

// Prompter shows a prompt and reads a line of input.
type Prompter interface {
	Input() string
}
