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
