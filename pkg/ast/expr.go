package ast

import "glox/pkg/token"

type Expr interface {
	expr()
}

type BinaryExpr struct {
	Left     Expr
	Operator token.Token
	Right    Expr
}

func (e *BinaryExpr) expr() {}

type GroupingExpr struct {
	Expression Expr
}

func (e *GroupingExpr) expr() {}

type LiteralExpr struct {
	Value any
}

func (e *LiteralExpr) expr() {}

type UnaryExpr struct {
	Operator token.Token
	Right    Expr
}

func (e *UnaryExpr) expr() {}

type VariableExpression struct {
	Name token.Token
}

func (e *VariableExpression) expr() {}
