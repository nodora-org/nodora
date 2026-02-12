package evaluator

import (
	"fmt"

	"nodora.org/nodora/pkg/nir"
)

type Evaluator struct {
	program         *nir.Program
	signalListeners map[string][]nir.SignalListener
	Debug           bool
}

type EvaluationResult struct {
	Outputs map[string]any      `json:"outputs"`
	Signals []nir.EmittedSignal `json:"emitted_signals"`
}

func NewEvaluator(program *nir.Program) *Evaluator {
	return &Evaluator{program: program, signalListeners: make(map[string][]nir.SignalListener)}
}

func (e *Evaluator) OnSignalFunc(signalName string, listener func([]nir.Value) error) {
	e.signalListeners[signalName] = append(e.signalListeners[signalName], nir.SignalListenerFunc(listener))
}

func (e *Evaluator) EvaluateRule(ruleName string, input nir.ValueMap) (*EvaluationResult, error) {
	rule, ok := e.program.Rules[ruleName]
	if !ok {
		return nil, fmt.Errorf("rule not found: %s", ruleName)
	}

	slots := make([]nir.Value, rule.Symslots)
	slots[0] = input

	evalCtx := &nir.EvaluationContext{
		Slots:     slots,
		Emissions: []nir.EmittedSignal{},
		Listeners: e.signalListeners,
	}

	// execute operations
	for _, op := range rule.Ops {
		if err := op.Execute(evalCtx); err != nil {
			return nil, fmt.Errorf("execution error: %v", err)
		}
	}

	// collect outputs
	outputs := make(map[string]any)
	for name, output := range rule.Outputs {

		if output.Sym >= len(evalCtx.Slots) {
			return nil, fmt.Errorf("output index out of bounds for %s", name)
		}
		val := evalCtx.Slots[output.Sym]
		outputs[name] = val
	}

	if e.Debug {
		debugRule(ruleName, &rule, input, evalCtx)
	}

	return &EvaluationResult{
		Outputs: outputs,
		Signals: evalCtx.Emissions,
	}, nil
}

func (e *Evaluator) GetRuleNames() []string {
	names := make([]string, 0, len(e.program.Rules))
	for name := range e.program.Rules {
		names = append(names, name)
	}
	return names
}

func (e *Evaluator) GetSignalNames() []string {
	names := make([]string, 0, len(e.program.Signals))
	for name := range e.program.Signals {
		names = append(names, name)
	}
	return names
}

func debugRule(ruleName string, rule *nir.Rule, input nir.ValueMap, evalCtx *nir.EvaluationContext) {
	fmt.Printf("=== rule %s ===\n", ruleName)

	fmt.Printf("\nInputs (len = %d)\n", len(input))
	for k, v := range input {
		fmt.Printf("  %s: %v\n", k, v)
	}

	fmt.Printf("\nSymbols (len = %d)\n", len(evalCtx.Slots))
	for symName, slotIdx := range rule.Symbols {
		fmt.Printf("  %s -> [%d]\n", symName, slotIdx)
	}

	fmt.Printf("\nSlots (len = %d)\n", len(evalCtx.Slots))
	for i, value := range evalCtx.Slots {
		fmt.Printf("  [%d]: %v\n", i, value)
	}

	fmt.Printf("\nOutputs (len = %d)\n", len(rule.Outputs))
	for outName, output := range rule.Outputs {
		fmt.Printf("  [%d]: %s = %v\n", output.Sym, outName, evalCtx.Slots[output.Sym])
	}

	fmt.Printf("\nOperations (len = %d)\n", len(rule.Ops))
	for i, op := range rule.Ops {
		outStr := ""
		if op.Out != nil {
			outStr = fmt.Sprintf(" -> [%d]", *op.Out)
		}
		fmt.Printf("  %d: %s%v%s\n", i, op.Kind, op.Args, outStr)
	}

	fmt.Printf("\nSignals Emitted (len = %d)\n", len(evalCtx.Emissions))
	for _, sig := range evalCtx.Emissions {
		fmt.Printf("  %s: %v\n", sig.Name, sig.Args)
	}

	fmt.Println("================================")
}
