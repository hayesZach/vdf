package vdf

import "fmt"

type SyntaxError struct {
	Line    int
	Column  int
	Message string
}

func (e *SyntaxError) Error() string {
	return fmt.Sprintf("%d:%d: syntax error: %s", e.Line, e.Column, e.Message)
}
