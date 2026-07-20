package interpreter

import (
	"fmt"
	gloxError "glox/pkg/error"
	"glox/pkg/token"
)

type Environment struct {
	values map[string]any
}

func NewEnvironment() *Environment {
	return &Environment{values: make(map[string]any)}
}

func (env *Environment) Get(name token.Token) (any, error) {
	if value, ok := env.values[name.Lexeme]; ok {
		return value, nil
	}

	return nil, gloxError.NewRuntimeError(name, fmt.Sprintf("Undefined variable '%s'.", name.Lexeme))
}

func (env *Environment) Assign(name token.Token, value any) error {
	if _, ok := env.values[name.Lexeme]; ok {
		env.values[name.Lexeme] = value
		return nil
	}

	return gloxError.NewRuntimeError(name, fmt.Sprintf("Undefined variable '%s'.", name.Lexeme))
}

func (env *Environment) Define(name string, value any) {
	env.values[name] = value
}
