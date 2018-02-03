package ion

import (
	"bufio"
	"bytes"
	"io"
	"strings"
)

type Token int

const (
	ILLEGAL Token = iota
	EOF
	WHITESPACE
	COMMA
	COLON
	DOUBLE_COLON
	SYMBOL
	STRING
	OPEN_BRACE
	CLOSE_BRACE
	OPEN_BRACKET
	CLOSE_BRACKET
	OPEN_PAREN
	CLOSE_PAREN
	NUMBER
)

func (t Token) String() string {
	switch t {
	case EOF:
		return "EOF"
	case WHITESPACE:
		return "WHITESPACE"
	case COMMA:
		return "COMMA"
	case COLON:
		return "COLON"
	case DOUBLE_COLON:
		return "DOUBLE_COLON"
	case SYMBOL:
		return "SYMBOL"
	case STRING:
		return "STRING"
	case OPEN_BRACE:
		return "OPEN_BRACE"
	case CLOSE_BRACE:
		return "CLOSE_BRACE"
	case OPEN_BRACKET:
		return "OPEN_BRACKET"
	case CLOSE_BRACKET:
		return "CLOSE_BRACKET"
	case OPEN_PAREN:
		return "OPEN_PAREN"
	case CLOSE_PAREN:
		return "CLOSE_PAREN"
	case NUMBER:
		return "NUMBER"
	}
	return "ILLEGAL"
}

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n'
}

func isLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isDigit(ch rune) bool {
	return (ch >= '0' && ch <= '9')
}

var eof = rune(0)

type Scanner struct {
	r           *bufio.Reader
	lastToken   Token
	lastLiteral string
}

func NewScanner(r io.Reader) *Scanner {
	return &Scanner{r: bufio.NewReader(r)}
}
func (s *Scanner) read() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	return ch
}

func (s *Scanner) unread() { _ = s.r.UnreadRune() }

func (s *Scanner) Unscan(tok Token, lit string) {
	s.lastToken = tok
	s.lastLiteral = lit
}

func (s *Scanner) Scan() (tok Token, lit string) {
	if s.lastToken != ILLEGAL {
		tok := s.lastToken
		lit := s.lastLiteral
		s.lastToken = ILLEGAL
		s.lastLiteral = ""
		return tok, lit
	}
	ch := s.read()

	if isWhitespace(ch) {
		s.unread()
		return s.scanWhitespace()
	} else if isLetter(ch) {
		s.unread()
		return s.scanIdentifier()
	}

	// Otherwise read the individual character.
	switch ch {
	case eof:
		return EOF, ""
	case '/':
		ch = s.read()
		if ch == '/' {
			s.skipLine()
			return s.Scan()
		} else {
			s.unread()
			return ILLEGAL, "/"
		}

	case ':':
		ch = s.read()
		if ch == ':' {
			return DOUBLE_COLON, "::"
		} else {
			s.unread()
			return COLON, ":"
		}
	case '\'':
		return s.scanUntil(SYMBOL, ch)
	case '"':
		return s.scanUntil(STRING, ch)
	case ',':
		return COMMA, string(ch)
	case '{':
		return OPEN_BRACE, string(ch)
	case '}':
		return CLOSE_BRACE, string(ch)
	case '[':
		return OPEN_BRACKET, string(ch)
	case ']':
		return CLOSE_BRACKET, string(ch)
	case '(':
		return OPEN_PAREN, string(ch)
	case ')':
		return CLOSE_PAREN, string(ch)
	}
	if isDigit(ch) {
		return s.scanNumber(ch)
	}
	return ILLEGAL, string(ch)
}

func (s *Scanner) skipLine() {
	for {
		if ch := s.read(); ch == eof || ch == '\n' {
			break
		}
	}
}

func (s *Scanner) scanNumber(first rune) (Token, string) {
	var buf bytes.Buffer
	buf.WriteRune(first)
	digits := "0123456789."
	if ch := s.read(); ch != eof {
		if first == '0' {
			if ch == 'x' {
				digits = "0123456789abcdefABCDEF."
				buf.WriteRune(ch)
			} else if ch == 'b' {
				digits = "01."
				buf.WriteRune(ch)
			} else {
				s.unread()
			}
		} else {
			s.unread()
		}
		for {
			if ch := s.read(); ch == eof {
				break
			} else if strings.Index(digits, string(ch)) >= 0 {
				buf.WriteRune(ch)
			} else {
				s.unread()
				break
			}
		}
	}
	return NUMBER, buf.String()
}

func (s *Scanner) scanUntil(tok Token, delim rune) (Token, string) {
	var buf bytes.Buffer
	buf.WriteRune(s.read())
	escape := false
	for {
		if ch := s.read(); ch == eof {
			break
		} else if escape {
			escape = false
			switch ch {
			case '"':
				buf.WriteRune('"')
			case 't':
				buf.WriteRune('\t')
			case 'n':
				buf.WriteRune('\n')
			case 'r':
				buf.WriteRune('\r')
			case '\n':
				//if newline, ignore subsequent whitespace before continuing with the string
				for {
					if ch := s.read(); ch == eof || !isWhitespace(ch) {
						break
					}
				}
				s.unread()
			default:
				return ILLEGAL, "\\" + string(ch)
			}
		} else if ch == delim {
			break
		} else if ch == '\\' {
			escape = true
		} else {
			buf.WriteRune(ch)
		}
	}
	return tok, buf.String()
}

func (s *Scanner) scanWhitespace() (tok Token, lit string) {
	var buf bytes.Buffer
	buf.WriteRune(s.read())
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isWhitespace(ch) {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}
	return WHITESPACE, buf.String()
}

func (s *Scanner) scanIdentifier() (tok Token, lit string) {
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isLetter(ch) && !isDigit(ch) && ch != '_' {
			s.unread()
			break
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}
	return SYMBOL, buf.String()
}
