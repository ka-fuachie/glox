package lexer

import (
	gloxError "glox/pkg/error"
	"glox/pkg/token"
	"iter"
	"strconv"
)

type Lexer struct {
	source               []rune
	start, current, line int
	reporter             *gloxError.Reporter
}

func New(source string, reporter *gloxError.Reporter) *Lexer {
	return &Lexer{source: []rune(source), start: 0, current: 0, line: 1, reporter: reporter}
}

func (l *Lexer) Tokens() iter.Seq[token.Token] {
	return func(yield func(token.Token) bool) {
		for {
			tok, err := l.NextToken()
			if err != nil {
				l.reporter.Report(err)
				continue
			}

			if !yield(tok) {
				break
			}

			if tok.Type == token.EOF {
				break
			}
		}
	}
}

func (l *Lexer) NextToken() (token.Token, error) {
	l.start = l.current
	if l.isAtEnd() {
		return l.token(token.EOF), nil
	}

	c := l.advance()

	switch c {
	case '(':
		return l.token(token.LEFT_PAREN), nil
	case ')':
		return l.token(token.RIGHT_PAREN), nil
	case '{':
		return l.token(token.LEFT_BRACE), nil
	case '}':
		return l.token(token.RIGHT_BRACE), nil
	case ',':
		return l.token(token.COMMA), nil
	case '.':
		return l.token(token.DOT), nil
	case '-':
		return l.token(token.MINUS), nil
	case '+':
		return l.token(token.PLUS), nil
	case ';':
		return l.token(token.SEMICOLON), nil
	case '*':
		return l.token(token.STAR), nil
	case '!':
		if l.match('=') {
			return l.token(token.BANG_EQUAL), nil
		} else {
			return l.token(token.BANG), nil
		}
	case '=':
		if l.match('=') {
			return l.token(token.EQUAL_EQUAL), nil
		} else {
			return l.token(token.EQUAL), nil
		}
	case '<':
		if l.match('=') {
			return l.token(token.LESS_EQUAL), nil
		} else {
			return l.token(token.LESS), nil
		}
	case '>':
		if l.match('=') {
			return l.token(token.GREATER_EQUAL), nil
		} else {
			return l.token(token.GREATER), nil
		}
	case '/':
		if l.match('/') {
			for l.peek() != '\n' && !l.isAtEnd() {
				l.advance()
			}

			return l.NextToken()
		} else if l.match('*') {
			_, err := l.blockComment()
			if err != nil {
				return token.Token{}, err
			}
			return l.NextToken()
		} else {
			return l.token(token.SLASH), nil
		}

	case ' ', '\t', '\r':
		return l.NextToken()

	case '\n':
		l.line++
		return l.NextToken()

	case '"':
		return l.string()

	default:
		if isDigit(c) {
			return l.number()
		} else if isAlpha(c) {
			return l.identifier()
		}

		return token.Token{}, gloxError.LineStaticError(l.line, "Unexpected character.")
	}
}

func (l *Lexer) isAtEnd() bool {
	return l.current >= len(l.source)
}

func (l *Lexer) advance() rune {
	c := l.source[l.current]
	l.current++
	return c
}

func (l *Lexer) token(t token.TokenType) token.Token {
	return l.tokenWithLiteral(t, nil)
}

func (l *Lexer) tokenWithLiteral(t token.TokenType, literal any) token.Token {
	text := string(l.source[l.start:l.current])
	return token.Token{Type: t, Lexeme: text, Literal: literal, Line: l.line}
}

func (l *Lexer) match(c rune) bool {
	if l.isAtEnd() {
		return false
	}
	if l.source[l.current] != c {
		return false
	}

	l.current++
	return true
}

func (l *Lexer) peek() rune {
	if l.isAtEnd() {
		return '\x00'
	}

	return l.source[l.current]
}

func (l *Lexer) peekNext() rune {
	if l.current+1 >= len(l.source) {
		return '\x00'
	}

	return l.source[l.current+1]
}

func isDigit(c rune) bool {
	return c >= '0' && c <= '9'
}

func isAlpha(c rune) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		c == '_'
}

func isAlphaNumeric(c rune) bool {
	return isAlpha(c) || isDigit(c)
}

func (l *Lexer) string() (token.Token, error) {
	for l.peek() != '"' && !l.isAtEnd() {
		if l.peek() == '\n' {
			l.line++
		}
		l.advance()
	}

	if l.isAtEnd() {
		return token.Token{}, gloxError.LineStaticError(l.line, "Unterminated string.")
	}

	l.advance() // past closing "

	value := string(l.source[l.start+1 : l.current-1])
	return l.tokenWithLiteral(token.STRING, value), nil
}

func (l *Lexer) number() (token.Token, error) {
	for isDigit(l.peek()) {
		l.advance()
	}

	if l.peek() == '.' && isDigit(l.peekNext()) {
		l.advance()
		for isDigit(l.peek()) {
			l.advance()
		}
	}

	value, err := strconv.ParseFloat(string(l.source[l.start:l.current]), 64)
	if err != nil {
		panic(err)
	}

	return l.tokenWithLiteral(token.NUMBER, value), nil
}

var keywords = map[string]token.TokenType{
	"and":    token.AND,
	"class":  token.CLASS,
	"else":   token.ELSE,
	"false":  token.FALSE,
	"for":    token.FOR,
	"fun":    token.FUN,
	"if":     token.IF,
	"nil":    token.NIL,
	"or":     token.OR,
	"print":  token.PRINT,
	"return": token.RETURN,
	"super":  token.SUPER,
	"this":   token.THIS,
	"true":   token.TRUE,
	"var":    token.VAR,
	"while":  token.WHILE,
}

func (l *Lexer) identifier() (token.Token, error) {
	for isAlphaNumeric(l.peek()) {
		l.advance()
	}

	text := string(l.source[l.start:l.current])
	if t, ok := keywords[text]; ok {
		return l.token(t), nil
	}

	return l.token(token.IDENTIFIER), nil
}

func (l *Lexer) blockComment() (token.Token, error) {
	for (l.peek() != '*' || l.peekNext() != '/') && !l.isAtEnd() {
		if l.peek() == '/' && l.peekNext() == '*' {
			// Nested block comment
			// for starting "/*"
			l.advance()
			l.advance()
			_, err := l.blockComment()
			if err != nil {
				return token.Token{}, err
			}
		} else {
			if l.peek() == '\n' {
				l.line++
			}
			l.advance()
		}
	}

	if l.isAtEnd() {
		return token.Token{}, gloxError.LineStaticError(l.line, "Unterminated block comment.")
	}

	// for ending "*/"
	l.advance()
	l.advance()

	return token.Token{}, nil
}
