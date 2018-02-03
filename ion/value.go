package ion

import (
	"bytes"
	"fmt"
)

type Type int

const (
	NullType Type = iota
	BoolType
	IntType
	FloatType
	StringType
	SymbolType
	StructType
	ListType
	SexpType
)

//a simplified view of what this can actually be
type Value struct {
	Type        Type
	Annotations []string
	Int         int64
	Float       float64
	Text        string
	Sequence    []Value
	Struct      []Field
}

type Field struct {
	Name  string
	Value Value
}

func (v Value) String() string {
	switch v.Type {
	case NullType:
		return "null"
	case BoolType:
		if v.Int == 0 {
			return "false"
		}
		return "true"
	case IntType:
		return fmt.Sprintf("%d", v.Int)
	case FloatType:
		return fmt.Sprintf("%g", v.Float)
	case StringType:
		return fmt.Sprintf("%q", v.Text)
	case SymbolType:
		return symbolToString(v.Text)
	case StructType:
		return annotate(v) + structToString(v.Struct)
	case ListType:
		return sequenceToString(v.Sequence, '[', ',', ']')
	case SexpType:
		return sequenceToString(v.Sequence, '(', 0, ')')
	default:
		return "?FIXME?"
	}
}

func annotate(val Value) string {
	if len(val.Annotations) > 0 {
		var buf bytes.Buffer
		for _, anno := range val.Annotations {
			buf.WriteString(anno)
			buf.WriteString("::")
		}
		return buf.String()
	}
	return ""
}

func symbolToString(val string) string {
	//to do: escape embedded single quotes
	//for now, also single-quote, so we can distinguish them from keywords when debugging
	//if strings.Index(val, " ") >= 0 {
	return fmt.Sprintf("'%s'", val)
	//}
	//return val
}

func structToString(fields []Field) string {
	switch len(fields) {
	case 0:
		return "{}"
	case 1:
		return "{" + fields[0].Name + ": " + fields[0].Value.String() + "}"
	default:
		var buf bytes.Buffer
		buf.WriteRune('{')
		first := true
		for _, item := range fields {
			if first {
				first = false
			} else {
				buf.WriteString(", ")
			}
			buf.WriteString(item.Name)
			buf.WriteString(": ")
			buf.WriteString(item.Value.String())
		}
		buf.WriteRune('}')
		return buf.String()
	}
}

func sequenceToString(values []Value, openChar, delimChar, closeChar rune) string {
	switch len(values) {
	case 0:
		return string(openChar) + string(closeChar)
	case 1:
		return string(openChar) + values[0].String() + string(closeChar)
	default:
		var buf bytes.Buffer
		buf.WriteRune(openChar)
		first := true
		for _, item := range values {
			if first {
				first = false
			} else {
				if delimChar != 0 {
					buf.WriteRune(delimChar)
				}
				buf.WriteRune(' ')
			}
			buf.WriteString(item.String())
		}
		buf.WriteRune(closeChar)
		return buf.String()
	}
}
