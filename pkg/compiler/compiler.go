package compiler

import (
	"nodora.org/nodora/internal/optimizer"
	"nodora.org/nodora/internal/optimizer/passes"
	"nodora.org/nodora/internal/parser"
	"nodora.org/nodora/internal/semantics"
	"nodora.org/nodora/pkg/nir"
)

type Compiler struct{}

func NewCompiler() *Compiler {
	return &Compiler{}
}

func (c *Compiler) Compile(src string) (*nir.Ruleset, error) {
	p, err := parser.Parse(src)
	if err != nil {
		return nil, err
	}

	analyzer := semantics.NewSemanticAnalyzer(src)
	if err := analyzer.Analyze(p); err != nil {
		return nil, err
	}

	builder := nir.NewBuilder()
	ruleset, err := builder.Build(p)
	if err != nil {
		return nil, err
	}

	opt := optimizer.NewOptimizer()
	opt.AddPass(passes.NewRepeatedPass(10, passes.NewConstantFolding(), passes.NewConstantPropagation()))
	opt.AddPass(passes.NewDeadCodeElimination())
	opt.AddPass(passes.NewSymbolRemap())

	if err := opt.Run(ruleset); err != nil {
		return nil, err
	}

	return ruleset, nil
}
