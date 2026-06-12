package parser

import (
	"strings"

	"nodora.org/nodora/internal/ast"
)

type lexer struct {
	input     string
	pos       int
	line      int
	col       int
	lastError string
	result    *ast.Program
}

var keywords = map[string]int{
	"signal": SIGNAL,
	"rule":   RULE,
	"const":  CONST,
	"emit":   EMIT,
	"when":   WHEN,
	"out":    OUT,
	"in":     IN,
	"if":     IF,
	"then":   THEN,
	"else":   ELSE,
	"true":   TRUE,
	"false":  FALSE,
	"match":  MATCH,
}

// NewLexer creates a new lexer for the given input string.
func newLexer(input string) *lexer {
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

	// Capture the start position of this token
	startLine := l.line
	startCol := l.col

	// Helper to set span in lval
	setSpan := func() {
		lval.span.Start = ast.Position{Line: startLine, Col: startCol}
		lval.span.End = ast.Position{Line: l.line, Col: l.col}
	}

	// two-char tokens
	if l.pos+1 < len(l.input) {
		two := l.input[l.pos : l.pos+2]
		switch two {
		case "&&":
			l.pos += 2
			l.col += 2
			setSpan()
			return AND
		case "||":
			l.pos += 2
			l.col += 2
			setSpan()
			return OR
		case "==":
			l.pos += 2
			l.col += 2
			setSpan()
			return EQ
		case "!=":
			l.pos += 2
			l.col += 2
			setSpan()
			return NEQ
		case ">=":
			l.pos += 2
			l.col += 2
			setSpan()
			return GTE
		case "<=":
			l.pos += 2
			l.col += 2
			setSpan()
			return LTE
		case "::":
			l.pos += 2
			l.col += 2
			setSpan()
			return NAMESPACE
		case "=>":
			l.pos += 2
			l.col += 2
			setSpan()
			return FATARROW
		}
	}

	ch := l.input[l.pos]

	// string literal
	if ch == '"' {
		l.pos++
		l.col++
		start := l.pos
		for l.pos < len(l.input) {
			if l.input[l.pos] == '"' {
				raw := l.input[start:l.pos]
				lval.str = unescape(raw)
				l.pos++
				l.col++
				setSpan()
				return STRING
			}
			if l.input[l.pos] == '\\' && l.pos+1 < len(l.input) {
				// skip escape; the content is captured in unescape below
				l.pos += 2
				l.col += 2
				continue
			}
			if l.input[l.pos] == '\n' {
				l.line++
				l.col = 0
			} else {
				l.col++
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
		l.col++
		setSpan()
		return LPAREN
	case ')':
		l.pos++
		l.col++
		setSpan()
		return RPAREN
	case '!':
		l.pos++
		l.col++
		setSpan()
		return NOT
	case '{':
		l.pos++
		l.col++
		setSpan()
		return LBRACE
	case '}':
		l.pos++
		l.col++
		setSpan()
		return RBRACE
	case '[':
		l.pos++
		l.col++
		setSpan()
		return LBRACKET
	case ']':
		l.pos++
		l.col++
		setSpan()
		return RBRACKET
	case '?':
		l.pos++
		l.col++
		setSpan()
		return QMARK
	case ':':
		l.pos++
		l.col++
		setSpan()
		return COLON
	case ',':
		l.pos++
		l.col++
		setSpan()
		return COMMA
	case '.':
		l.pos++
		l.col++
		setSpan()
		return DOT
	case '>':
		l.pos++
		l.col++
		setSpan()
		return GT
	case '<':
		l.pos++
		l.col++
		setSpan()
		return LT
	case '=':
		l.pos++
		l.col++
		setSpan()
		return ASSIGN
	case '+':
		l.pos++
		l.col++
		setSpan()
		return PLUS
	case '-':
		l.pos++
		l.col++
		setSpan()
		return MINUS
	case '*':
		l.pos++
		l.col++
		setSpan()
		return STAR
	case '/':
		l.pos++
		l.col++
		setSpan()
		return SLASH
	case '%':
		l.pos++
		l.col++
		setSpan()
		return MOD
	case '|':
		l.pos++
		l.col++
		setSpan()
		return PIPE
	}

	// identifiers or keywords
	if isAlpha(ch) {
		start := l.pos
		l.pos++
		l.col++
		for l.pos < len(l.input) {
			c := l.input[l.pos]
			if isAlphaNum(c) {
				l.pos++
				l.col++
			} else {
				break
			}
		}
		lit := l.input[start:l.pos]
		if tok, ok := keywords[lit]; ok {
			setSpan()
			return tok
		}
		lval.str = lit
		setSpan()
		return IDENT
	}

	// numbers
	if isDigit(ch) {
		start := l.pos
		for l.pos < len(l.input) && isDigit(l.input[l.pos]) {
			l.pos++
			l.col++
		}
		lval.str = l.input[start:l.pos]
		setSpan()
		return NUMBER
	}

	// unknown character
	l.pos++
	l.col++
	l.Error("unexpected character: " + string(ch))
	return 0
}

func (l *lexer) Error(s string) {
	l.lastError = s
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
