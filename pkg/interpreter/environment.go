package interpreter

import (
	"fmt"
	gloxError "glox/pkg/error"
	"glox/pkg/token"
)

type Environment struct {
	enclosing *Environment
	values    map[string]any
}

func NewEnvironment(enclosing *Environment) *Environment {
	return &Environment{values: make(map[string]any), enclosing: enclosing}
}

func (env *Environment) Get(name token.Token) (any, error) {
	if value, ok := env.values[name.Lexeme]; ok {
		return value, nil
	}

	if env.enclosing != nil {
		return env.enclosing.Get(name)
	}

	return nil, gloxError.NewRuntimeError(name, fmt.Sprintf("Undefined variable '%s'.", name.Lexeme))
}

func (env *Environment) Assign(name token.Token, value any) error {
	if _, ok := env.values[name.Lexeme]; ok {
		env.values[name.Lexeme] = value
		return nil
	}

	if env.enclosing != nil {
		return env.enclosing.Assign(name, value)
	}

	return gloxError.NewRuntimeError(name, fmt.Sprintf("Undefined variable '%s'.", name.Lexeme))
}

func (env *Environment) Define(name string, value any) {
	env.values[name] = value
}

func (env *Environment) GetAt(distance int, name string) any {
	return env.ancestor(distance).values[name]
}

func (env *Environment) AssignAt(distance int, name string, value any) {
	env.ancestor(distance).values[name] = value
}

func (env *Environment) ancestor(distance int) *Environment {
	environment := env
	for range distance {
		environment = environment.enclosing
	}
	return environment
}
