package compiler

import (
	"nodora.org/nodora/internal/parser"
	"nodora.org/nodora/internal/semantics"
	"nodora.org/nodora/pkg/nir"
)

type Compiler struct{}

func NewCompiler() *Compiler {
	return &Compiler{}
}

func (c *Compiler) Compile(src string) (*nir.Program, error) {
	p, err := parser.Parse(src)
	if err != nil {
		return nil, err
	}

	analyzer := semantics.NewSemanticAnalyzer()
	if err := analyzer.Analyze(p); err != nil {
		return nil, err
	}

	builder := nir.NewBuilder()
	prog, err := builder.Build(p)
	if err != nil {
		return nil, err
	}

	return prog, nil
}
