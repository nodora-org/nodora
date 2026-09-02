package nir

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"nodora.org/nodora/pkg/core"
)

// binary op kinds rendered back as source-level operators
var binaryOpSymbols = map[OpKind]string{
	OpAdd: "+",
	OpSub: "-",
	OpMul: "*",
	OpDiv: "/",
	OpMod: "%",
	OpAnd: "&&",
	OpOr:  "||",
	OpGt:  ">",
	OpGte: ">=",
	OpLt:  "<",
	OpLte: "<=",
	OpEq:  "==",
	OpNeq: "!=",
	OpIn:  "in",
}

func SlotRef(index int) string {
	if index == 0 {
		return "input"
	}
	return fmt.Sprintf("[%d]", index)
}

// Renders a runtime value in source-literal form.
func FormatValue(v core.Value) string {
	if v.IsUndefined() {
		return "undefined"
	}
	switch raw := v.Raw.(type) {
	case nil:
		return "null"
	case core.Value:
		return FormatValue(raw)
	case string:
		return strconv.Quote(raw)
	case bool:
		return strconv.FormatBool(raw)
	case float64:
		return strconv.FormatFloat(raw, 'g', -1, 64)
	case []core.Value:
		parts := make([]string, len(raw))
		for i, e := range raw {
			parts[i] = FormatValue(e)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case core.ValueMap:
		keys := make([]string, 0, len(raw))
		for k := range raw {
			keys = append(keys, k)
		}
		slices.Sort(keys)
		parts := make([]string, len(keys))
		for i, k := range keys {
			parts[i] = k + ": " + FormatValue(raw[k])
		}
		return "{" + strings.Join(parts, ", ") + "}"
	case *core.Lambda:
		return "<lambda>"
	default:
		return fmt.Sprintf("%v", raw)
	}
}

func (w RawExpr) String() string {
	if w.Expr == nil {
		return "<nil>"
	}
	if s, ok := w.Expr.(fmt.Stringer); ok {
		return s.String()
	}
	return fmt.Sprintf("%v", w.Expr)
}

func joinExprs(exprs []RawExpr) string {
	parts := make([]string, len(exprs))
	for i, e := range exprs {
		parts[i] = e.String()
	}
	return strings.Join(parts, ", ")
}

func (i *ImmExpr) String() string {
	return FormatValue(i.Value)
}

func (a *ArrExpr) String() string {
	return "[" + joinExprs(a.Value) + "]"
}

func (o *ObjExpr) String() string {
	keys := make([]string, 0, len(o.Value))
	for k := range o.Value {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = k + ": " + o.Value[k].String()
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

func (s *SymExpr) String() string {
	return SlotRef(s.Index)
}

func (i *IdxExpr) String() string {
	return i.From.String() + "[" + i.Index.String() + "]"
}

func (s *SelExpr) String() string {
	var sb strings.Builder
	sb.WriteString(s.From.String())
	if s.Path == "" {
		return sb.String()
	}

	exprIdx := 0
	for _, key := range strings.Split(s.Path, ".") {
		if key != "$" {
			sb.WriteString(".")
			sb.WriteString(key)
			continue
		}
		if exprIdx < len(s.Exprs) {
			sb.WriteString("[")
			sb.WriteString(s.Exprs[exprIdx].String())
			sb.WriteString("]")
			exprIdx++
		} else {
			sb.WriteString("[?]")
		}
	}
	return sb.String()
}

func (s *SignalExpr) String() string {
	out := s.Name + "(" + joinExprs(s.Args) + ")"
	if s.When != nil {
		out += " when " + s.When.String()
	}
	return out
}

func (c *CallExpr) String() string {
	name := c.Func.Name
	if c.Func.Namespace != "" {
		name = c.Func.Namespace + "::" + name
	}
	return name + "(" + joinExprs(c.Args) + ")"
}

func (le *LambdaExpr) Header() string {
	params := make([]string, len(le.Params))
	for i, p := range le.Params {
		params[i] = SlotRef(p.SymIndex)
	}
	return "|" + strings.Join(params, ", ") + "|"
}

func (le *LambdaExpr) String() string {
	body := make([]string, len(le.Ops))
	for i := range le.Ops {
		body[i] = le.Ops[i].String()
	}
	return le.Header() + " { " + strings.Join(body, "; ") + " }"
}

func (op *Op) ExprString() string {
	return op.exprStringWith(func(i int) string {
		if i < len(op.Args) {
			return op.Args[i].String()
		}
		return "?"
	})
}

func (op *Op) exprStringWith(arg func(i int) string) string {
	switch op.Kind {
	case OpCopy:
		return arg(0)
	case OpNot:
		return "!" + arg(0)
	case OpSelect:
		return arg(0) + " ? " + arg(1) + " : " + arg(2)
	case OpEmit:
		return "emit " + arg(0)
	}

	if sym, ok := binaryOpSymbols[op.Kind]; ok {
		return arg(0) + " " + sym + " " + arg(1)
	}
	parts := make([]string, len(op.Args))
	for i := range op.Args {
		parts[i] = arg(i)
	}
	return string(op.Kind) + "(" + strings.Join(parts, ", ") + ")"
}

func (op *Op) String() string {
	if op.Out == nil {
		return op.ExprString()
	}
	return SlotRef(*op.Out) + " = " + op.ExprString()
}
