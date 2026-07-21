package parser

import (
	"glox/pkg/ast"
	gloxError "glox/pkg/error"
	"glox/pkg/token"
	"slices"
)

type Parser struct {
	tokens   []token.Token
	current  int
	reporter *gloxError.Reporter
}

func New(tokens []token.Token, reporter *gloxError.Reporter) *Parser {
	return &Parser{tokens: tokens, current: 0, reporter: reporter}
}

func (p *Parser) Parse() []ast.Stmt {
	var statements []ast.Stmt

	for !p.isAtEnd() {
		stmt, err := p.declaration()
		if err != nil {
			p.reporter.Report(err)
			p.synchronize()
			continue
		}
		statements = append(statements, stmt)
	}

	return statements
}

func (p *Parser) match(tokenTypes ...token.TokenType) bool {
	if slices.ContainsFunc(tokenTypes, p.check) {
		p.advance()
		return true
	}

	return false
}

func (p *Parser) check(tokenType token.TokenType) bool {
	if p.isAtEnd() {
		return false
	}

	return p.peek().Type == tokenType
}

func (p *Parser) advance() token.Token {
	if !p.isAtEnd() {
		p.current++
	}

	return p.previous()
}

func (p *Parser) isAtEnd() bool {
	return p.peek().Type == token.EOF
}

func (p *Parser) peek() token.Token {
	return p.tokens[p.current]
}

func (p *Parser) previous() token.Token {
	return p.tokens[p.current-1]
}

func (p *Parser) consume(tokenType token.TokenType, message string) (token.Token, error) {
	if p.check(tokenType) {
		return p.advance(), nil
	}

	return token.Token{}, gloxError.TokenStaticError(p.peek(), message)
}

func (p *Parser) synchronize() {
	p.advance()

	for !p.isAtEnd() {
		if p.previous().Type == token.SEMICOLON {
			return
		}

		switch p.peek().Type {
		case token.CLASS, token.FUN, token.VAR, token.FOR, token.IF, token.WHILE, token.PRINT, token.RETURN:
			return
		}

		p.advance()
	}
}

func (p *Parser) expression() (ast.Expr, error) {
	return p.assignment()
}

func (p *Parser) declaration() (ast.Stmt, error) {
	if p.match(token.VAR) {
		return p.varDeclaration()
	}

	return p.statement()
}

func (p *Parser) statement() (ast.Stmt, error) {
	if p.match(token.PRINT) {
		return p.printStatement()
	}

	if p.match(token.LEFT_BRACE) {
		statements, err := p.block()
		if err != nil {
			return nil, err
		}
		return &ast.BlockStmt{Statements: statements}, nil
	}

	return p.expressionStatement()
}

func (p *Parser) expressionStatement() (ast.Stmt, error) {
	expr, err := p.expression()
	if err != nil {
		return nil, err
	}

	_, err = p.consume(token.SEMICOLON, "Expect ';' after expression.")
	if err != nil {
		return nil, err
	}

	return &ast.ExpressionStmt{Expression: expr}, nil
}

func (p *Parser) block() ([]ast.Stmt, error) {
	var statements []ast.Stmt
	for !p.check(token.RIGHT_BRACE) {
		stmt, err := p.declaration()
		if err != nil {
			return nil, err
		}

		statements = append(statements, stmt)
	}

	_, err := p.consume(token.RIGHT_BRACE, "Expect '}' after block.")
	if err != nil {
		return nil, err
	}

	return statements, nil
}

func (p *Parser) printStatement() (ast.Stmt, error) {
	value, err := p.expression()
	if err != nil {
		return nil, err
	}

	_, err = p.consume(token.SEMICOLON, "Expect ';' after value.")
	if err != nil {
		return nil, err
	}

	return &ast.PrintStmt{Expression: value}, nil
}

func (p *Parser) varDeclaration() (ast.Stmt, error) {
	name, err := p.consume(token.IDENTIFIER, "Expect variable name.")
	if err != nil {
		return nil, err
	}

	var initializer ast.Expr
	if p.match(token.EQUAL) {
		initializer, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	_, err = p.consume(token.SEMICOLON, "Expect ';' after variable declaration.")

	return &ast.VarStmt{Name: name, Initializer: initializer}, nil
}

func (p *Parser) assignment() (ast.Expr, error) {
	expr, err := p.equality()
	if err != nil {
		return nil, err
	}

	if p.match(token.EQUAL) {
		equals := p.previous()
		value, err := p.assignment()
		if err != nil {
			return nil, err
		}

		if e, ok := expr.(*ast.VariableExpression); ok {
			name := e.Name
			return &ast.AssignExpr{Name: name, Expression: value}, nil
		}

		return nil, gloxError.TokenStaticError(equals, "Invalid assignment target.")
	}

	return expr, err
}

func (p *Parser) equality() (ast.Expr, error) {
	expr, err := p.comparison()
	if err != nil {
		return nil, err
	}

	for p.match(token.BANG_EQUAL, token.EQUAL_EQUAL) {
		operator := p.previous()
		right, err := p.comparison()
		if err != nil {
			return nil, err
		}
		expr = &ast.BinaryExpr{Left: expr, Operator: operator, Right: right}
	}

	return expr, nil
}

func (p *Parser) comparison() (ast.Expr, error) {
	expr, err := p.term()
	if err != nil {
		return nil, err
	}

	for p.match(token.GREATER, token.GREATER_EQUAL, token.LESS, token.LESS_EQUAL) {
		operator := p.previous()
		right, err := p.term()
		if err != nil {
			return nil, err
		}
		expr = &ast.BinaryExpr{Left: expr, Operator: operator, Right: right}
	}

	return expr, nil
}

func (p *Parser) term() (ast.Expr, error) {
	expr, err := p.factor()
	if err != nil {
		return nil, err
	}

	for p.match(token.MINUS, token.PLUS) {
		operator := p.previous()
		right, err := p.factor()
		if err != nil {
			return nil, err
		}
		expr = &ast.BinaryExpr{Left: expr, Operator: operator, Right: right}
	}

	return expr, nil
}

func (p *Parser) factor() (ast.Expr, error) {
	expr, err := p.unary()
	if err != nil {
		return nil, err
	}

	for p.match(token.SLASH, token.STAR) {
		operator := p.previous()
		right, err := p.unary()
		if err != nil {
			return nil, err
		}
		expr = &ast.BinaryExpr{Left: expr, Operator: operator, Right: right}
	}

	return expr, nil
}

func (p *Parser) unary() (ast.Expr, error) {
	if p.match(token.BANG, token.MINUS) {
		operator := p.previous()
		right, err := p.unary()
		if err != nil {
			return nil, err
		}
		return &ast.UnaryExpr{Operator: operator, Right: right}, nil
	}

	return p.primary()
}

func (p *Parser) primary() (ast.Expr, error) {
	if p.match(token.TRUE) {
		return &ast.LiteralExpr{Value: true}, nil
	}
	if p.match(token.FALSE) {
		return &ast.LiteralExpr{Value: false}, nil
	}
	if p.match(token.NIL) {
		return &ast.LiteralExpr{Value: nil}, nil
	}

	if p.match(token.NUMBER, token.STRING) {
		return &ast.LiteralExpr{Value: p.previous().Literal}, nil
	}

	if p.match(token.IDENTIFIER) {
		return &ast.VariableExpression{Name: p.previous()}, nil
	}

	if p.match(token.LEFT_PAREN) {
		expr, err := p.expression()
		if err != nil {
			return nil, err
		}
		_, err = p.consume(token.RIGHT_PAREN, "Expect ')' after expression.")
		if err != nil {
			return nil, err
		}
		return &ast.GroupingExpr{Expression: expr}, nil
	}

	return nil, gloxError.TokenStaticError(p.peek(), "Expect expression.")
}
