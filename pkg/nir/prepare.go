package nir

import (
	"fmt"
	"strings"

	"nodora.org/nodora/pkg/registry"
)

// integer form of OpKind
type opCode uint8

const (
	opInvalid opCode = iota
	opcCopy
	opcAdd
	opcSub
	opcMul
	opcDiv
	opcMod
	opcAnd
	opcOr
	opcGt
	opcGte
	opcLt
	opcLte
	opcEq
	opcNeq
	opcIn
	opcNot
	opcSelect
	opcEmit
)

type opSpec struct {
	code opCode
	args int
}

var opSpecs = map[OpKind]opSpec{
	OpCopy:   {opcCopy, 1},
	OpAdd:    {opcAdd, 2},
	OpSub:    {opcSub, 2},
	OpMul:    {opcMul, 2},
	OpDiv:    {opcDiv, 2},
	OpMod:    {opcMod, 2},
	OpAnd:    {opcAnd, 2},
	OpOr:     {opcOr, 2},
	OpGt:     {opcGt, 2},
	OpGte:    {opcGte, 2},
	OpLt:     {opcLt, 2},
	OpLte:    {opcLte, 2},
	OpEq:     {opcEq, 2},
	OpNeq:    {opcNeq, 2},
	OpIn:     {opcIn, 2},
	OpNot:    {opcNot, 1},
	OpSelect: {opcSelect, 3},
	OpEmit:   {opcEmit, 1},
}

// Prepares the ruleset for evaluation, resolving references and validating it.
func (p *Ruleset) Prepare() error {
	return p.prepare()
}

func (p *Ruleset) prepare() error {
	for name, rule := range p.Rules {
		if err := prepareOps(rule.Ops, rule.Symslots); err != nil {
			return fmt.Errorf("rule %s: %w", name, err)
		}
	}
	return nil
}

func prepareOps(ops []Op, symslots int) error {
	for i := range ops {
		if err := prepareOp(&ops[i], symslots); err != nil {
			return err
		}
	}
	return nil
}

func prepareOp(op *Op, symslots int) error {
	spec, ok := opSpecs[op.Kind]
	if !ok {
		return fmt.Errorf("unknown operation %q", op.Kind)
	}
	op.code = spec.code

	if len(op.Args) != spec.args {
		return fmt.Errorf("operation %q expects %d argument(s), got %d", op.Kind, spec.args, len(op.Args))
	}

	if op.code != opcEmit {
		if op.Out == nil {
			return fmt.Errorf("operation %q requires an output slot", op.Kind)
		}
		if *op.Out < 0 || *op.Out >= symslots {
			return fmt.Errorf("operation %q output slot %d out of bounds [0,%d)", op.Kind, *op.Out, symslots)
		}
	}

	for i := range op.Args {
		if err := prepareExpr(op.Args[i].Expr, symslots); err != nil {
			return err
		}
	}
	return nil
}

func prepareExpr(e Expr, symslots int) error {
	switch ex := e.(type) {
	case nil, *ImmExpr:
		// nothing to resolve

	case *SymExpr:
		if ex.Index < 0 || ex.Index >= symslots {
			return fmt.Errorf("symbol index %d out of bounds [0,%d)", ex.Index, symslots)
		}

	case *CallExpr:
		fn, ok := registry.Global().Get(ex.Func.Namespace, ex.Func.Name)
		if !ok {
			return fmt.Errorf("undefined function '%s'", ex.Func.String())
		}
		if req := fn.RequiredArgCount(); len(ex.Args) < req {
			return fmt.Errorf("function '%s' expects %d argument(s), got %d", fn.FullPath(), req, len(ex.Args))
		}
		ex.fn = fn
		for i := range ex.Args {
			if err := prepareExpr(ex.Args[i].Expr, symslots); err != nil {
				return err
			}
		}

	case *ArrExpr:
		for i := range ex.Value {
			if err := prepareExpr(ex.Value[i].Expr, symslots); err != nil {
				return err
			}
		}

	case *ObjExpr:
		for k := range ex.Value {
			if err := prepareExpr(ex.Value[k].Expr, symslots); err != nil {
				return err
			}
		}

	case *IdxExpr:
		if err := prepareExpr(ex.From.Expr, symslots); err != nil {
			return err
		}
		if err := prepareExpr(ex.Index.Expr, symslots); err != nil {
			return err
		}

	case *SelExpr:
		ex.keys = strings.Split(ex.Path, ".")
		if err := prepareExpr(ex.From.Expr, symslots); err != nil {
			return err
		}
		for i := range ex.Exprs {
			if err := prepareExpr(ex.Exprs[i].Expr, symslots); err != nil {
				return err
			}
		}

	case *SignalExpr:
		for i := range ex.Args {
			if err := prepareExpr(ex.Args[i].Expr, symslots); err != nil {
				return err
			}
		}
		if ex.When != nil {
			if err := prepareExpr(ex.When.Expr, symslots); err != nil {
				return err
			}
		}

	case *LambdaExpr:
		if err := prepareOps(ex.Ops, symslots); err != nil {
			return err
		}
	}
	return nil
}
