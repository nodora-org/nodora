package parser

import (
	"strings"
)

// lexer implements a simple, goyacc-compatible lexer for the grammar in parser.y.
// It returns token codes defined in the generated parser.go and places literal
// values (identifiers, strings, numbers) into the yylval (yySymType).

type lexer struct {
	input     string
	pos       int
	line      int
	col       int
	lastError string
}

var keywords = map[string]int{
	"signal": SIGNAL,
	"rule":   RULE,
	"emit":   EMIT,
	"when":   WHEN,
	"out":    OUT,
	"in":     IN,
	"if":     IF,
	"then":   THEN,
	"else":   ELSE,
	"true":   TRUE,
	"false":  FALSE,
}

// NewLexer creates a new lexer for the given input string.
func NewLexer(input string) *lexer {
	return &lexer{input: input, line: 1, col: 0}
}

func (l *lexer) Lex(lval *yySymType) int {
	if l == nil {
		return 0
	}
	// skip whitespace and comments
	for {
		if l.pos >= len(l.input) {
			return 0
		}
		c := l.input[l.pos]
		// whitespace
		if c == ' ' || c == '\t' || c == '\r' {
			l.pos++
			l.col++
			continue
		}
		if c == '\n' {
			l.pos++
			l.line++
			l.col = 0
			continue
		}
		// line comment // ...
		if c == '/' && l.pos+1 < len(l.input) && l.input[l.pos+1] == '/' {
			l.pos += 2
			for l.pos < len(l.input) && l.input[l.pos] != '\n' {
				l.pos++
			}
			continue
		}
		break
	}
	if l.pos >= len(l.input) {
		return 0
	}

	// two-char tokens
	if l.pos+1 < len(l.input) {
		two := l.input[l.pos : l.pos+2]
		switch two {
		case "&&":
			l.pos += 2
			return AND
		case "||":
			l.pos += 2
			return OR
		case "==":
			l.pos += 2
			return EQ
		case "!=":
			l.pos += 2
			return NEQ
		case ">=":
			l.pos += 2
			return GTE
		case "<=":
			l.pos += 2
			return LTE
		}
	}

	ch := l.input[l.pos]

	// string literal
	if ch == '"' {
		l.pos++
		start := l.pos
		for l.pos < len(l.input) {
			if l.input[l.pos] == '"' {
				raw := l.input[start:l.pos]
				lval.str = unescape(raw)
				l.pos++
				return STRING
			}
			if l.input[l.pos] == '\\' && l.pos+1 < len(l.input) {
				// skip escape; the content is captured in unescape below
				l.pos += 2
				continue
			}
			l.pos++
		}
		l.lastError = "unterminated string"
		l.Error(l.lastError)
		return 0
	}

	// punctuation tokens
	switch ch {
	case '(':
		l.pos++
		return LPAREN
	case ')':
		l.pos++
		return RPAREN
	case '!':
		l.pos++
		return NOT
	case '{':
		l.pos++
		return LBRACE
	case '}':
		l.pos++
		return RBRACE
	case '[':
		l.pos++
		return LBRACKET
	case ']':
		l.pos++
		return RBRACKET
	case '?':
		l.pos++
		return QMARK
	case ':':
		l.pos++
		return COLON
	case ',':
		l.pos++
		return COMMA
	case '.':
		l.pos++
		return DOT
	case '>':
		l.pos++
		return GT
	case '<':
		l.pos++
		return LT
	case '=':
		l.pos++
		return ASSIGN
	case '+':
		l.pos++
		return PLUS
	case '-':
		l.pos++
		return MINUS
	case '*':
		l.pos++
		return STAR
	case '/':
		l.pos++
		return SLASH
	case '%':
		l.pos++
		return MOD
	}

	// identifiers or keywords
	if isAlpha(ch) {
		start := l.pos
		l.pos++
		for l.pos < len(l.input) {
			c := l.input[l.pos]
			if isAlphaNum(c) {
				l.pos++
			} else {
				break
			}
		}
		lit := l.input[start:l.pos]
		if tok, ok := keywords[lit]; ok {
			return tok
		}
		lval.str = lit
		return IDENT
	}

	// numbers
	if isDigit(ch) {
		start := l.pos
		for l.pos < len(l.input) && isDigit(l.input[l.pos]) {
			l.pos++
		}
		lval.str = l.input[start:l.pos]
		return NUMBER
	}

	// unknown character
	l.pos++
	l.Error("unexpected character: " + string(ch))
	return 0
}

func (l *lexer) Error(s string) {
	l.lastError = s
	// Could log here if desired
}

// Helpers
func isAlpha(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_'
}

func isAlphaNum(b byte) bool {
	return isAlpha(b) || (b >= '0' && b <= '9')
}

func isDigit(b byte) bool {
	return (b >= '0' && b <= '9') || b == '.'
}

// unescape handles basic string escapes (n, t, r, ", \\).
func unescape(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '\\' && i+1 < len(s) {
			i++
			switch s[i] {
			case 'n':
				b.WriteByte('\n')
			case 't':
				b.WriteByte('\t')
			case 'r':
				b.WriteByte('\r')
			case '"':
				b.WriteByte('"')
			case '\\':
				b.WriteByte('\\')
			default:
				b.WriteByte(s[i])
			}
		} else {
			b.WriteByte(c)
		}
	}
	return b.String()
}
