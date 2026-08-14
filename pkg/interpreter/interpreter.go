package interpreter

import (
	"fmt"
	"glox/pkg/ast"
	gloxError "glox/pkg/error"
	"glox/pkg/token"
	"io"
	"time"
)

type Interpreter struct {
	out         io.Writer
	globals     *Environment
	environment *Environment
	locals      map[ast.Expr]int
}

func New(out io.Writer) *Interpreter {
	globals := NewEnvironment(nil)
	globals.Define("clock", &Clock{})

	return &Interpreter{out: out, globals: globals, environment: globals, locals: make(map[ast.Expr]int)}
}

type Clock struct{}

func (*Clock) Arity() int { return 0 }
func (*Clock) Call(i *Interpreter, arguments []any) (any, error) {
	now := time.Now()
	return float64(now.UnixMilli()) / 1000, nil
}
func (*Clock) String() string {
	return "<native fn>"
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
	case *ast.AssignExpr:
		value, err := i.evaluate(e.Value)
		if err != nil {
			return nil, err
		}

		distance, ok := i.locals[e]
		if ok {
			i.environment.AssignAt(distance, e.Name.Lexeme, value)
		} else {
			err = i.globals.Assign(e.Name, value)
			if err != nil {
				return nil, err
			}
		}

		return value, nil
	case *ast.LiteralExpr:
		return e.Value, nil
	case *ast.LogicalExpr:
		left, err := i.evaluate(e.Left)
		if err != nil {
			return nil, err
		}

		switch e.Operator.Type {
		case token.AND:
			if !i.isTruthy(left) {
				return left, nil
			}
		case token.OR:
			if i.isTruthy(left) {
				return left, nil
			}
		default:
			panic("invalid operator")
		}

		right, err := i.evaluate(e.Right)
		if err != nil {
			return nil, err
		}

		return right, nil
	case *ast.SetExpr:
		object, err := i.evaluate(e.Object)
		if err != nil {
			return nil, err
		}

		if instance, ok := object.(*LoxInstance); ok {
			value, err := i.evaluate(e.Value)
			if err != nil {
				return nil, err
			}

			instance.Set(e.Name, value)
			return value, nil
		}

		return nil, gloxError.NewRuntimeError(e.Name, "Only instances have fields.")
	case *ast.ThisExpr:
		return i.lookUpVariable(e.Keyword, e)
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
	case *ast.VariableExpression:
		return i.lookUpVariable(e.Name, expr)
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
	case *ast.CallExpr:
		callee, err := i.evaluate(e.Callee)
		if err != nil {
			return nil, err
		}

		var arguments []any
		for _, argument := range e.Arguments {
			arg, err := i.evaluate(argument)
			if err != nil {
				return nil, err
			}
			arguments = append(arguments, arg)
		}

		function, ok := callee.(LoxCallable)
		if !ok {
			return nil, gloxError.NewRuntimeError(e.Paren, "Can only call functions and classes.")
		}

		if len(arguments) != function.Arity() {
			return nil, gloxError.NewRuntimeError(e.Paren,
				fmt.Sprintf("Expected %d arguments but got %d.", function.Arity(), len(arguments)))
		}

		return function.Call(i, arguments)
	case *ast.GetExpr:
		object, err := i.evaluate(e.Object)
		if err != nil {
			return nil, err
		}

		if instance, ok := object.(*LoxInstance); ok {
			return instance.Get(e.Name)
		}

		return nil, gloxError.NewRuntimeError(e.Name, "Only instances have properties.")
	default:
		panic("invalid expression")
	}
}

func (i *Interpreter) lookUpVariable(name token.Token, expr ast.Expr) (any, error) {
	distance, ok := i.locals[expr]
	if ok {
		return i.environment.GetAt(distance, name.Lexeme), nil
	} else {
		return i.globals.Get(name)
	}
}

func (i *Interpreter) execute(stmt ast.Stmt) error {
	switch s := stmt.(type) {
	case *ast.BlockStmt:
		return i.executeBlock(s.Statements, NewEnvironment(i.environment))
	case *ast.ClassStmt:
		i.environment.Define(s.Name.Lexeme, nil)

		methods := make(map[string]*LoxFunction)
		for _, method := range s.Methods {
			function := NewLoxFunction(method, i.environment, method.Name.Lexeme == "init")
			methods[method.Name.Lexeme] = function
		}

		class := NewLoxClass(s.Name.Lexeme, methods)
		i.environment.Assign(s.Name, class)
		return nil
	case *ast.ExpressionStmt:
		_, err := i.evaluate(s.Expression)
		if err != nil {
			return err
		}

		return nil
	case *ast.FunctionStmt:
		function := NewLoxFunction(s, i.environment, false)
		i.environment.Define(s.Name.Lexeme, function)
		return nil
	case *ast.IfStmt:
		condition, err := i.evaluate(s.Condition)
		if err != nil {
			return err
		}

		if i.isTruthy(condition) {
			err := i.execute(s.ThenBranch)
			if err != nil {
				return err
			}
		} else {
			err := i.execute(s.ElseBranch)
			if err != nil {
				return err
			}
		}

		return nil
	case *ast.PrintStmt:
		value, err := i.evaluate(s.Expression)
		if err != nil {
			return err
		}

		fmt.Fprintln(i.out, stringValue(value))
		return nil
	case *ast.ReturnStmt:
		value, err := i.evaluate(s.Value)
		if err != nil {
			return err
		}

		return NewReturn(value)
	case *ast.VarStmt:
		var value any

		if s.Initializer != nil {
			var err error
			value, err = i.evaluate(s.Initializer)
			if err != nil {
				return err
			}
		}

		i.environment.Define(s.Name.Lexeme, value)
		return nil
	case *ast.WhileStmt:
		for {
			condition, err := i.evaluate(s.Condition)
			if err != nil {
				return err
			}

			if !i.isTruthy(condition) {
				break
			}

			err = i.execute(s.Body)
			if err != nil {
				return err
			}
		}

		return nil
	}

	return nil
}

func (i *Interpreter) executeBlock(stmts []ast.Stmt, environment *Environment) error {
	prevEnvironment := i.environment
	i.environment = environment
	defer func() {
		i.environment = prevEnvironment
	}()

	for _, stmt := range stmts {
		err := i.execute(stmt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *Interpreter) Resolve(expr ast.Expr, depth int) {
	i.locals[expr] = depth
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

	return fmt.Sprintf("%v", result)
}
