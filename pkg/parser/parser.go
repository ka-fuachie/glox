package parser

import (
	"glox/pkg/ast"
	gloxError "glox/pkg/error"
	"glox/pkg/token"
	"slices"
)

type Parser struct {
	tokens  []token.Token
	current int
}

func New(tokens []token.Token) *Parser {
	return &Parser{tokens: tokens, current: 0}
}

func (p *Parser) Parse() (ast.Expr, error) {
	expr, err := p.expression()
	if err != nil {
		return nil, err
	}

	return expr, nil
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

	return token.Token{}, gloxError.TokenError(p.peek(), message)
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
	return p.equality()
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

	return nil, gloxError.TokenError(p.peek(), "Expect expression.")
}
