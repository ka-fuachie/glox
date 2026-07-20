package ast

import (
	"fmt"
	"strings"
)

func Print(expr Expr) string {
	switch e := expr.(type) {
	case *BinaryExpr:
		return parenthesize(e.Operator.Lexeme, e.Left, e.Right)
	case *GroupingExpr:
		return parenthesize("group", e.Expression)
	case *LiteralExpr:
		if e.Value == nil {
			return "nil"
		}
		return fmt.Sprintf("%v", e.Value)
	case *UnaryExpr:
		return parenthesize(e.Operator.Lexeme, e.Right)
	default:
		return ""
	}
}

func parenthesize(name string, exprs ...Expr) string {
	sb := strings.Builder{}
	sb.WriteRune('(')
	sb.WriteString(name)
	for _, e := range exprs {
		fmt.Fprintf(&sb, " %s", Print(e))
	}
	sb.WriteRune(')')
	return sb.String()
}

func PrintRPN(expr Expr) string {
	switch e := expr.(type) {
	case *BinaryExpr:
		return fmt.Sprintf("%s %s %s", PrintRPN(e.Left), PrintRPN(e.Right), e.Operator.Lexeme)
	case *GroupingExpr:
		return PrintRPN(e.Expression)
	case *LiteralExpr:
		if e.Value == nil {
			return "nil"
		}
		return fmt.Sprintf("%v", e.Value)
	case *UnaryExpr:
		return fmt.Sprintf("%s %s", PrintRPN(e.Right), e.Operator.Lexeme)
	default:
		return ""
	}
}
