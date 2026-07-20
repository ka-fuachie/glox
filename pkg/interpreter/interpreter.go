package interpreter

import (
	"fmt"
	"glox/pkg/ast"
	gloxError "glox/pkg/error"
	"glox/pkg/token"
	"io"
)

type Interpreter struct {
	out io.Writer
}

func New(out io.Writer) *Interpreter {
	return &Interpreter{out: out}
}

func (i *Interpreter) Interpret(stmts []ast.Stmt) error {
	for _, stmt := range stmts {
		err := i.execute(stmt)
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *Interpreter) evaluate(expr ast.Expr) (any, error) {
	switch e := expr.(type) {
	case *ast.LiteralExpr:
		return e.Value, nil
	case *ast.GroupingExpr:
		return i.evaluate(e.Expression)
	case *ast.UnaryExpr:
		right, err := i.evaluate(e.Right)
		if err != nil {
			return nil, err
		}

		switch e.Operator.Type {
		case token.MINUS:
			right, err := i.checkNumberOperand(e.Operator, right)
			if err != nil {
				return nil, err
			}
			return -right, nil
		case token.BANG:
			return !i.isTruthy(right), nil
		default:
			panic("invalid operator")
		}
	case *ast.BinaryExpr:
		left, err := i.evaluate(e.Left)
		if err != nil {
			return nil, err
		}
		right, err := i.evaluate(e.Right)
		if err != nil {
			return nil, err
		}

		switch e.Operator.Type {
		case token.GREATER:
			left, right, err := i.checkNumberOperands(e.Operator, left, right)
			if err != nil {
				return nil, err
			}
			return left > right, nil
		case token.GREATER_EQUAL:
			left, right, err := i.checkNumberOperands(e.Operator, left, right)
			if err != nil {
				return nil, err
			}
			return left >= right, nil
		case token.LESS:
			left, right, err := i.checkNumberOperands(e.Operator, left, right)
			if err != nil {
				return nil, err
			}
			return left < right, nil
		case token.LESS_EQUAL:
			left, right, err := i.checkNumberOperands(e.Operator, left, right)
			if err != nil {
				return nil, err
			}
			return left <= right, nil
		case token.BANG_EQUAL:
			left, right, err := i.checkNumberOperands(e.Operator, left, right)
			if err != nil {
				return nil, err
			}
			return !i.isEqual(left, right), nil
		case token.EQUAL_EQUAL:
			left, right, err := i.checkNumberOperands(e.Operator, left, right)
			if err != nil {
				return nil, err
			}
			return i.isEqual(left, right), nil
		case token.MINUS:
			left, right, err := i.checkNumberOperands(e.Operator, left, right)
			if err != nil {
				return nil, err
			}
			return left - right, nil
		case token.SLASH:
			left, right, err := i.checkNumberOperands(e.Operator, left, right)
			if err != nil {
				return nil, err
			}
			return left / right, nil
		case token.STAR:
			left, right, err := i.checkNumberOperands(e.Operator, left, right)
			if err != nil {
				return nil, err
			}
			return left * right, nil
		case token.PLUS:
			strL, okStrL := left.(string)
			strR, okStrR := right.(string)
			if okStrL && okStrR {
				return strL + strR, nil
			}

			numL, okNumL := left.(float64)
			numR, okNumR := right.(float64)
			if okNumL && okNumR {
				return numL + numR, nil
			}

			return nil, gloxError.NewRuntimeError(e.Operator, "Operands must be two numbers or two strings.")
		default:
			panic("invalid operator")
		}
	default:
		panic("invalid expression")
	}
}

func (i *Interpreter) execute(expr ast.Stmt) error {
	switch e := expr.(type) {
	case *ast.ExpressionStmt:
		_, err := i.evaluate(e.Expression)
		if err != nil {
			return err
		}

		return nil
	case *ast.PrintStmt:
		value, err := i.evaluate(e.Expression)
		if err != nil {
			return err
		}

		fmt.Fprintln(i.out, stringValue(value))
		return nil
	default:
		panic("invalid statement")
	}
}

func (i *Interpreter) isTruthy(val any) bool {
	if val == nil {
		return false
	}

	if v, ok := val.(bool); ok {
		return v
	}

	return true
}

func (i *Interpreter) isEqual(a, b any) bool {
	return a == b
}

func (i *Interpreter) checkNumberOperand(operator token.Token, operand any) (float64, error) {
	if num, ok := operand.(float64); ok {
		return num, nil
	}

	return 0, gloxError.NewRuntimeError(operator, "Operand must be a number.")
}

func (i *Interpreter) checkNumberOperands(operator token.Token, left any, right any) (float64, float64, error) {
	leftNum, okLeft := left.(float64)
	rightNum, okRight := right.(float64)
	if okLeft && okRight {
		return leftNum, rightNum, nil
	}

	return 0, 0, gloxError.NewRuntimeError(operator, "Operands must be numbers.")
}

func stringValue(result any) string {
	if result == nil {
		return "nil"
	}

	return fmt.Sprintf("%#v", result)
}
