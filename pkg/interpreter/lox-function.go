package interpreter

import (
	"errors"
	"fmt"
	"glox/pkg/ast"
)

type LoxFunction struct {
	Declaration *ast.FunctionStmt
	Closure     *Environment
}

func NewLoxFunction(declaration *ast.FunctionStmt, closure *Environment) *LoxFunction {
	return &LoxFunction{Declaration: declaration, Closure: closure}
}

func (f *LoxFunction) Arity() int {
	return len(f.Declaration.Params)
}

func (f *LoxFunction) Call(i *Interpreter, arguments []any) (any, error) {
	environment := NewEnvironment(f.Closure)
	for i, param := range f.Declaration.Params {
		environment.Define(param.Lexeme, arguments[i])
	}

	err := i.executeBlock(f.Declaration.Body, environment)
	if err != nil {
		if e, ok := errors.AsType[*Return](err); ok {
			return e.Value, nil
		}

		return nil, err
	}

	return nil, nil
}

func (f *LoxFunction) String() string {
	return fmt.Sprintf("<fn %s>", f.Declaration.Name.Lexeme)
}
