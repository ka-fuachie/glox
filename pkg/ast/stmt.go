package ast

import "glox/pkg/token"

type Stmt interface {
	stmt()
}

type BlockStmt struct {
	Statements []Stmt
}

func (s *BlockStmt) stmt() {}

type ExpressionStmt struct {
	Expression Expr
}

func (s *ExpressionStmt) stmt() {}

type PrintStmt struct {
	Expression Expr
}

func (s *PrintStmt) stmt() {}

type VarStmt struct {
	Name        token.Token
	Initializer Expr
}

func (s *VarStmt) stmt() {}
