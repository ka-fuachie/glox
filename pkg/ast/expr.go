package ast

import "glox/pkg/token"

type Expr interface {
	expr()
}

type AssignExpr struct {
	Name  token.Token
	Value Expr
}

func (e *AssignExpr) expr() {}

type BinaryExpr struct {
	Left     Expr
	Operator token.Token
	Right    Expr
}

func (e *BinaryExpr) expr() {}

type CallExpr struct {
	Callee    Expr
	Paren     token.Token
	Arguments []Expr
}

func (e *CallExpr) expr() {}

type GetExpr struct {
	Object Expr
	Name   token.Token
}

func (e *GetExpr) expr() {}

type GroupingExpr struct {
	Expression Expr
}

func (e *GroupingExpr) expr() {}

type LiteralExpr struct {
	Value any
}

func (e *LiteralExpr) expr() {}

type LogicalExpr struct {
	Left     Expr
	Operator token.Token
	Right    Expr
}

func (e *LogicalExpr) expr() {}

type SetExpr struct {
	Object Expr
	Name   token.Token
	Value  Expr
}

func (e *SetExpr) expr() {}

type ThisExpr struct {
	Keyword token.Token
}

func (e *ThisExpr) expr() {}

type UnaryExpr struct {
	Operator token.Token
	Right    Expr
}

func (e *UnaryExpr) expr() {}

type VariableExpression struct {
	Name token.Token
}

func (e *VariableExpression) expr() {}
