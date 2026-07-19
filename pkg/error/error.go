package error

import (
	"fmt"
	gloxToken "glox/pkg/token"
)

type EvalError struct {
	Line    int
	Where   string
	Message string
}

func (e *EvalError) Error() string {
	return fmt.Sprintf("[line: %d] Error%s: %s", e.Line, e.Where, e.Message)
}

func LineError(line int, message string) *EvalError {
	return &EvalError{Line: line, Message: message}
}

func TokenError(token gloxToken.Token, message string) *EvalError {
	var where string
	if token.Type == gloxToken.EOF {
		where = " at end"
	} else {
		where = fmt.Sprintf(" at '%s'", token.Lexeme)
	}

	return &EvalError{Line: token.Line, Where: where, Message: message}
}
