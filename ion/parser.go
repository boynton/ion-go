package ion

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type Parser struct {
	scanner *Scanner
	err     error
	source  string
	buf     struct {
		tok Token  // last read token
		lit string // last read literal
		n   int    // buffer size (max=1)
	}
}

func ParseFile(path string) (*Value, error) {
	fi, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fi.Close()
	reader := bufio.NewReader(fi)
	return parseFrom(path, reader)
}

func Parse(reader io.Reader) (*Value, error) {
	return parseFrom("", reader)
}

func parseFrom(source string, reader io.Reader) (*Value, error) {
	p := &Parser{scanner: NewScanner(reader), source: source}
	return p.parse()
}

func (p *Parser) scan() (tok Token, lit string) {
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok, p.buf.lit
	}
	tok, lit = p.scanner.Scan()
	p.buf.tok, p.buf.lit = tok, lit

	return
}

func (p *Parser) unscan() { p.buf.n = 1 }

func (p *Parser) scanIgnoreWhitespace() (tok Token, lit string) {
	tok, lit = p.scan()
	if tok == WHITESPACE {
		tok, lit = p.scan()
	}
	return
}

func (p *Parser) parse() (*Value, error) {
	tok, lit := p.scanIgnoreWhitespace()
	return p.parseToken(tok, lit)
}

func (p *Parser) parseToken(tok Token, lit string) (*Value, error) {
	if tok != EOF {
		if tok == ILLEGAL {
			p.err = fmt.Errorf("token not handled: %s - %q", tok, lit)
			return nil, p.err
		}
		switch tok {
		case SYMBOL:
			nextTok, _ := p.scanIgnoreWhitespace()
			if nextTok == DOUBLE_COLON {
				//fix me to not be a hack that assumes only s single annotation
				val, err := p.parse()
				if err == nil && val != nil {
					val.Annotations = []string{lit}
				}
				return val, err
			} else {
				p.unscan()
			}
			if lit == "true" {
				return &Value{Type: BoolType, Int: 1}, nil
			} else if lit == "false" {
				return &Value{Type: BoolType, Int: 0}, nil
			} else if lit == "null" {
				return &Value{Type: NullType}, nil
			}
			return &Value{Type: SymbolType, Text: lit}, nil
		case OPEN_PAREN:
			return p.parseSequence(CLOSE_PAREN)
		case OPEN_BRACKET:
			return p.parseSequence(CLOSE_BRACKET)
		case OPEN_BRACE:
			return p.parseStruct()
		case CLOSE_BRACE, CLOSE_BRACKET, CLOSE_PAREN, DOUBLE_COLON:
			return nil, fmt.Errorf("Unexpected %q", string(tok))
		case COMMA, COLON:
			return nil, nil //we basically ignore commas
		case NUMBER:
			if strings.Index(lit, ".") >= 0 {
				//to do: handle arbitrary precision decimal
				if !strings.HasPrefix(lit, "0x") && !strings.HasPrefix(lit, "0b") {
					n, err := strconv.ParseFloat(lit, 64)
					if err == nil {
						return &Value{Type: FloatType, Float: n}, nil
					}
				}
				return nil, fmt.Errorf("Cannot parse real number: %q", lit)
			} else {
				base := 10
				if strings.HasPrefix(lit, "0x") {
					base = 16
					lit = lit[2:]
				} else if strings.HasPrefix(lit, "0b") {
					base = 2
					lit = lit[2:]
				}
				i, err := strconv.ParseInt(lit, base, 64)
				if err != nil {
					return nil, fmt.Errorf("Cannot parse base %d integer: %q", base, lit)
				}
				return &Value{Type: IntType, Int: i}, nil
			}
		case STRING:
			return &Value{Type: StringType, Text: lit}, nil
		default:
			p.err = fmt.Errorf("token not handled: %s - %q", tok, lit)
			return nil, p.err
		}
	}
	return nil, nil
}

func (p *Parser) parseSequence(end Token) (*Value, error) {
	seq := make([]Value, 0)
	tok, lit := p.scanIgnoreWhitespace()
	for tok != EOF {
		if tok == CLOSE_BRACKET || tok == CLOSE_PAREN {
			if end != tok {
				return nil, fmt.Errorf("Bad sequence, expecting %v, encounted %s", end, tok)
			}
			if end == CLOSE_PAREN {
				return &Value{Type: SexpType, Sequence: seq}, nil
			}
			return &Value{Type: ListType, Sequence: seq}, nil
		} else {
			//to do: fix this to error on missing commas, this assumes they are optional
			elem, err := p.parseToken(tok, lit)
			if err != nil {
				return nil, err
			}
			if elem != nil {
				seq = append(seq, *elem)
			}
			tok, lit = p.scanIgnoreWhitespace()
		}
	}
	return nil, fmt.Errorf("Unexpected EOF")
}

func (p *Parser) parseStruct() (*Value, error) {
	fields := make([]Field, 0)
	tok, lit := p.scanIgnoreWhitespace()
	for tok != EOF {
		if tok == CLOSE_BRACE {
			return &Value{Type: StructType, Struct: fields}, nil
		} else if tok == COMMA {
			tok, lit = p.scanIgnoreWhitespace()
		} else {
			elem, err := p.parseToken(tok, lit)
			if err != nil {
				return nil, err
			}
			if elem.Type != SymbolType && elem.Type != StringType {
				return nil, fmt.Errorf("Invalid struct field name: %v", elem)
			}
			var field Field
			field.Name = elem.Text
			tok, lit = p.scanIgnoreWhitespace()
			if tok != COLON {
				return nil, fmt.Errorf("Bad struct syntax, encountered %v", tok)
			}
			elem, err = p.parse()
			if err != nil {
				return nil, err
			}
			field.Value = *elem
			fields = append(fields, field)
			tok, lit = p.scanIgnoreWhitespace()
		}
	}
	return nil, fmt.Errorf("Unexpected EOF")
}
