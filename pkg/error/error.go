package error

import "fmt"

type EvalError struct {
	Line    int
	Where   string
	Message string
}

func (e *EvalError) Error() string {
	return fmt.Sprintf("[line: %d] Error%s: %s", e.Line, e.Where, e.Message)
}
