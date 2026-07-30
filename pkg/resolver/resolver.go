package resolver

import (
	"errors"
	"glox/pkg/ast"
	gloxError "glox/pkg/error"
	"glox/pkg/interpreter"
	"glox/pkg/token"
	"iter"
	"slices"
)

type FunctionType int

const (
	FunctionTypeNone FunctionType = iota
	FunctionTypeFunction
)

type Resolver struct {
	interpreter     *interpreter.Interpreter
	scopes          []map[string]bool
	currentFunction FunctionType
}

func NewResolver(interpreter *interpreter.Interpreter) *Resolver {
	return &Resolver{
		interpreter:     interpreter,
		scopes:          make([]map[string]bool, 0),
		currentFunction: FunctionTypeNone,
	}
}

func (r *Resolver) Resolve(statements []ast.Stmt) iter.Seq[error] {
	return r.resolveStatements(statements)
}

func (r *Resolver) resolveStatements(statements []ast.Stmt) iter.Seq[error] {
	return func(yield func(error) bool) {
		for _, stmt := range statements {
			err := r.resolveStatement(stmt)
			if err != nil {
				if !yield(err) {
					break
				}
			}
		}
	}
}

func (r *Resolver) resolveStatement(stmt ast.Stmt) error {
	switch s := stmt.(type) {
	case *ast.BlockStmt:
		r.beginScope()
		errs := slices.Collect(r.resolveStatements(s.Statements))
		r.endScope()
		return errors.Join(errs...)

	case *ast.ExpressionStmt:
		return r.resolveExpression(s.Expression)

	case *ast.FunctionStmt:
		var errs []error
		errs = append(errs, r.declare(s.Name))
		r.define(s.Name)
		errs = append(errs, r.resolveFunction(s, FunctionTypeFunction))
		return errors.Join(errs...)

	case *ast.IfStmt:
		var errs []error
		errs = append(errs, r.resolveExpression(s.Condition))
		errs = append(errs, r.resolveStatement(s.ThenBranch))
		if s.ElseBranch != nil {
			errs = append(errs, r.resolveStatement(s.ElseBranch))
		}
		return errors.Join(errs...)

	case *ast.PrintStmt:
		return r.resolveExpression(s.Expression)

	case *ast.ReturnStmt:
		if r.currentFunction == FunctionTypeNone {
			return gloxError.TokenStaticError(s.Keyword, "Can't return from top-level code.")
		}

		if s.Value != nil {
			return r.resolveExpression(s.Value)
		}

	case *ast.VarStmt:
		var errs []error
		errs = append(errs, r.declare(s.Name))
		errs = append(errs, r.resolveExpression(s.Initializer))
		r.define(s.Name)
		return errors.Join(errs...)

	case *ast.WhileStmt:
		var errs []error
		errs = append(errs, r.resolveExpression(s.Condition))
		errs = append(errs, r.resolveStatement(s.Body))
		return errors.Join(errs...)
	}

	return nil
}

func (r *Resolver) resolveExpression(expr ast.Expr) error {
	switch e := expr.(type) {
	case *ast.AssignExpr:
		var err error
		err = r.resolveExpression(e.Value)
		r.resolveLocal(e, e.Name)
		return err

	case *ast.BinaryExpr:
		var errs []error
		errs = append(errs, r.resolveExpression(e.Left))
		errs = append(errs, r.resolveExpression(e.Right))
		return errors.Join(errs...)

	case *ast.CallExpr:
		var errs []error
		errs = append(errs, r.resolveExpression(e.Callee))
		for _, arg := range e.Arguments {
			errs = append(errs, r.resolveExpression(arg))
		}
		return errors.Join(errs...)

	case *ast.GroupingExpr:
		return r.resolveExpression(e.Expression)

	case *ast.LiteralExpr:
		return nil

	case *ast.LogicalExpr:
		var errs []error
		errs = append(errs, r.resolveExpression(e.Left))
		errs = append(errs, r.resolveExpression(e.Right))
		return errors.Join(errs...)

	case *ast.UnaryExpr:
		return r.resolveExpression(e.Right)

	case *ast.VariableExpression:
		var err error
		if len(r.scopes) > 0 {
			if flag, ok := r.scopes[len(r.scopes)-1][e.Name.Lexeme]; ok && flag == false {
				err = gloxError.TokenStaticError(e.Name, "Can't read local variable in it's own initializer.")
			}
		}

		r.resolveLocal(e, e.Name)
		return err
	}

	return nil
}

func (r *Resolver) resolveFunction(function *ast.FunctionStmt, functionType FunctionType) error {
	var errs []error
	enclosingFunction := r.currentFunction
	r.currentFunction = functionType
	r.beginScope()
	for _, param := range function.Params {
		errs = append(errs, r.declare(param))
		r.define(param)
	}
	errs = append(errs, slices.Collect(r.resolveStatements(function.Body))...)
	r.endScope()
	r.currentFunction = enclosingFunction

	return errors.Join(errs...)
}

func (r *Resolver) beginScope() {
	r.scopes = append(r.scopes, make(map[string]bool))
}

func (r *Resolver) endScope() {
	r.scopes = r.scopes[:len(r.scopes)-1]
}

func (r *Resolver) declare(name token.Token) error {
	if len(r.scopes) == 0 {
		return nil
	}

	scope := r.scopes[len(r.scopes)-1]
	if _, ok := scope[name.Lexeme]; ok {
		return gloxError.TokenStaticError(name, "Already a variable with this name in scope.")
	}

	scope[name.Lexeme] = false
	return nil
}

func (r *Resolver) define(name token.Token) {
	if len(r.scopes) == 0 {
		return
	}

	scope := r.scopes[len(r.scopes)-1]
	scope[name.Lexeme] = true
}

func (r *Resolver) resolveLocal(expr ast.Expr, name token.Token) {
	for i, scope := range slices.Backward(r.scopes) {
		if _, ok := scope[name.Lexeme]; ok {
			r.interpreter.Resolve(expr, len(r.scopes)-i-1)
		}
	}
}
