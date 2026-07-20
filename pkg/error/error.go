package error

import (
	"fmt"
	"glox/pkg/token"
	gloxToken "glox/pkg/token"
)

type StaticError struct {
	Line    int
	Where   string
	Message string
}

func (e *StaticError) Error() string {
	return fmt.Sprintf("[line: %d] Error%s: %s", e.Line, e.Where, e.Message)
}

func LineStaticError(line int, message string) *StaticError {
	return &StaticError{Line: line, Message: message}
}

func TokenStaticError(token gloxToken.Token, message string) *StaticError {
	var where string
	if token.Type == gloxToken.EOF {
		where = " at end"
	} else {
		where = fmt.Sprintf(" at '%s'", token.Lexeme)
	}

	return &StaticError{Line: token.Line, Where: where, Message: message}
}

type RuntimeError struct {
	Token   token.Token
	Message string
}

func (e *RuntimeError) Error() string {
	return fmt.Sprintf("%s\n[line %d]", e.Message, e.Token.Line)
}

func NewRuntimeError(token token.Token, message string) *RuntimeError {
	return &RuntimeError{Token: token, Message: message}
}
