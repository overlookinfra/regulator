package rgerror

import (
	"fmt"
)

type RGerrorType int

const (
	ShellError RGerrorType = iota
	ExecError
	CompletedError
	InvalidInput
	RemoteExecError
)

func (ar RGerrorType) String() string {
	return []string{"Shell command failed:", "Execution failed:", "Already done:", "Invalid input:", "Remote execution failed:"}[ar]
}

// RGerror is a custom error type that provides a
// Kind field for parsing different error types.
type RGerror struct {
	Kind    RGerrorType
	Message string
	Origin  error
}

func (e *RGerror) Error() string {
	if e.Origin != nil {
		return fmt.Sprintf("%s\n%s\n\nTrace:\n%s\n", e.Kind, e.Message, e.Origin)
	} else {
		return fmt.Sprintf("%s\n%s\n", e.Kind, e.Message)
	}
}
