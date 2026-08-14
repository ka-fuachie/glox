package interpreter

import (
	"fmt"
	gloxError "glox/pkg/error"
	"glox/pkg/token"
)

type LoxInstance struct {
	class  *LoxClass
	fields map[string]any
}

func NewLoxInstance(class *LoxClass) *LoxInstance {
	return &LoxInstance{class: class, fields: make(map[string]any)}
}

func (i *LoxInstance) String() string {
	return fmt.Sprintf("%s instance", i.class.name)
}

func (i *LoxInstance) Get(name token.Token) (any, error) {
	if value, ok := i.fields[name.Lexeme]; ok {
		return value, nil
	}

	method := i.class.FindMethod(name.Lexeme)
	if method != nil {
		return method.Bind(i), nil
	}

	return nil, gloxError.NewRuntimeError(name, fmt.Sprintf("Undefined property '%s'", name.Lexeme))
}

func (i *LoxInstance) Set(name token.Token, value any) {
	i.fields[name.Lexeme] = value
}
