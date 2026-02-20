package evaluator

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/nir"
)

type SignalListener interface {
	Invoke(args []any) error
}

type SignalListenerFunc func(args []any) error

func (f SignalListenerFunc) Invoke(args []any) error {
	defer func() { recover() }() // silent recovery
	return f(args)
}

type Evaluator struct {
	program         *nir.Program
	signalListeners map[string][]SignalListener
	Debug           bool
	signalWG        *sync.WaitGroup
}

type EvaluationResult struct {
	Outputs map[string]any      `json:"outputs"`
	Signals []nir.EmittedSignal `json:"emitted_signals"`
}

func NewEvaluator(program *nir.Program) *Evaluator {
	return &Evaluator{program: program, signalListeners: make(map[string][]SignalListener)}
}

func (e *Evaluator) SetWaitGroup(wg *sync.WaitGroup) {
	e.signalWG = wg
}

func (e *Evaluator) OnSignal(signalName string, listener func([]any) error) {
	e.signalListeners[signalName] = append(e.signalListeners[signalName], SignalListenerFunc(listener))
}

func (e *Evaluator) OnSignalNamed(signalName string, listener func(map[string]any) error) error {
	signal, ok := e.program.GetSignal(signalName)
	if !ok {
		keys := maps.Keys(e.program.Signals)
		return fmt.Errorf("unknown signal %q (available signals: %s)",
			signalName,
			strings.Join(slices.Collect(keys), ", "),
		)
	}

	gen := func(args []any) error {
		argsMap := make(map[string]any, len(signal.Params))
		for i, param := range signal.Params {
			if i < len(args) {
				argsMap[param.Name] = args[i]
			}
		}
		return listener(argsMap)
	}

	e.OnSignal(signalName, gen)
	return nil
}

func (e *Evaluator) EvaluateRule(ruleName string, input core.ValueMap) (*EvaluationResult, error) {
	rule, ok := e.program.GetRule(ruleName)
	if !ok {
		return nil, fmt.Errorf("rule not found: %s", ruleName)
	}

	slots := make([]core.Value, rule.Symslots)
	slots[0] = core.V(input)

	evalCtx := &nir.EvaluationContext{
		Slots:     slots,
		Emissions: []nir.EmittedSignal{},
	}

	for i, op := range rule.Ops {
		if err := op.Execute(evalCtx); err != nil {
			return nil, fmt.Errorf("execution error at op[%d]: %v", i, err)
		}
	}

	outputs := make(map[string]any)
	for name, output := range rule.Outputs {
		if output.Sym >= len(evalCtx.Slots) {
			return nil, fmt.Errorf("output index out of bounds for %s", name)
		}
		val := evalCtx.Slots[output.Sym]
		if !val.Undefined {
			outputs[name] = val.ToRaw()
		}
	}

	for i := range evalCtx.Emissions {
		es := &evalCtx.Emissions[i]
		e.invokeListeners(es.Name, es.Args, e.signalWG)
	}

	if e.Debug {
		debugRule(ruleName, rule, input, evalCtx)
	}

	return &EvaluationResult{
		Outputs: outputs,
		Signals: evalCtx.Emissions,
	}, nil
}

func (e *Evaluator) invokeListeners(
	signalName string,
	args []any,
	wg *sync.WaitGroup,
) {
	listeners, exists := e.signalListeners[signalName]
	if !exists {
		return
	}
	for _, listener := range listeners {
		l := listener

		if wg == nil {
			l.Invoke(args)
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			l.Invoke(args)
		}()
	}
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

func debugRule(ruleName string, rule *nir.Rule, input core.ValueMap, evalCtx *nir.EvaluationContext) {
	fmt.Printf("=== rule %s ===\n", ruleName)

	fmt.Printf("\nInputs (len = %d)\n", len(input))
	for k, v := range input {
		fmt.Printf("  %s: %v\n", k, v)
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
