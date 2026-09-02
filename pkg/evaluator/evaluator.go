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
	Trace           bool
	ruleset         *nir.Ruleset
	signalListeners map[string][]SignalListener
	slotPool        sync.Pool
	prepareOnce     sync.Once
	prepareErr      error
	signalWG        *sync.WaitGroup
}

type EvaluationResult struct {
	Outputs map[string]any      `json:"outputs"`
	Signals []nir.EmittedSignal `json:"emitted_signals"`
	Trace   *nir.Trace          `json:"-"`
}

func NewEvaluator(ruleset *nir.Ruleset) *Evaluator {
	return &Evaluator{ruleset: ruleset, signalListeners: make(map[string][]SignalListener)}
}

func (e *Evaluator) SetWaitGroup(wg *sync.WaitGroup) {
	e.signalWG = wg
}

func (e *Evaluator) OnSignal(signalName string, listener func([]any) error) {
	e.signalListeners[signalName] = append(e.signalListeners[signalName], SignalListenerFunc(listener))
}

func (e *Evaluator) OnSignalNamed(signalName string, listener func(map[string]any) error) error {
	signal, ok := e.ruleset.GetSignal(signalName)
	if !ok {
		keys := maps.Keys(e.ruleset.Signals)
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
	e.prepareOnce.Do(func() { e.prepareErr = e.ruleset.Prepare() })
	if e.prepareErr != nil {
		return nil, e.prepareErr
	}

	rule, ok := e.ruleset.GetRule(ruleName)
	if !ok {
		return nil, fmt.Errorf("rule not found: %s", ruleName)
	}

	ptr := e.acquireSlots(rule.Symslots)
	defer e.releaseSlots(ptr)

	slots := *ptr
	slots[0] = core.V(input)

	evalCtx := &nir.EvaluationContext{
		Slots:     slots,
		Emissions: []nir.EmittedSignal{},
	}

	var recorder *nir.Recorder
	if e.Trace {
		recorder = nir.NewRecorder()
		evalCtx.Tracer = recorder
	}

	for i := range rule.Ops {
		if err := rule.Ops[i].Execute(evalCtx); err != nil {
			err = fmt.Errorf("execution error at op[%d]: %v", i, err)
			if recorder == nil {
				return nil, err
			}
			trace := nir.NewTrace(ruleName, rule, input, evalCtx, recorder.Root())
			return &EvaluationResult{Trace: trace}, err
		}
	}

	outputs := make(map[string]any, len(rule.Outputs))
	for name, output := range rule.Outputs {
		if output.Sym >= len(evalCtx.Slots) {
			return nil, fmt.Errorf("output index out of bounds for %s", name)
		}
		val := evalCtx.Slots[output.Sym]
		if !val.IsUndefined() {
			outputs[name] = val.ToRaw()
		}
	}

	for i := range evalCtx.Emissions {
		es := &evalCtx.Emissions[i]
		e.invokeListeners(es.Name, es.Args, e.signalWG)
	}

	result := &EvaluationResult{
		Outputs: outputs,
		Signals: evalCtx.Emissions,
	}

	if recorder != nil {
		result.Trace = nir.NewTrace(ruleName, rule, input, evalCtx, recorder.Root())
	}

	return result, nil
}

// Returns a pooled slot buffer, stored as a pointer to avoid boxing the slice.
func (e *Evaluator) acquireSlots(n int) *[]core.Value {
	p, _ := e.slotPool.Get().(*[]core.Value)
	if p == nil {
		s := make([]core.Value, n)
		return &s
	}
	if cap(*p) < n {
		*p = make([]core.Value, n)
	} else {
		*p = (*p)[:n]
		clear(*p)
	}
	return p
}

func (e *Evaluator) releaseSlots(p *[]core.Value) {
	e.slotPool.Put(p)
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
	names := make([]string, 0, len(e.ruleset.Rules))
	for name := range e.ruleset.Rules {
		names = append(names, name)
	}
	return names
}

func (e *Evaluator) GetSignalNames() []string {
	names := make([]string, 0, len(e.ruleset.Signals))
	for name := range e.ruleset.Signals {
		names = append(names, name)
	}
	return names
}
