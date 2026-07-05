package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"encoding/json"
	"sync"
	"unsafe"

	"nodora.org/nodora/pkg/compiler"
	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/evaluator"
	"nodora.org/nodora/pkg/nir"
	_ "nodora.org/nodora/pkg/registry/all"
)

var (
	mu         sync.Mutex
	evaluators = make(map[int64]*evaluator.Evaluator)
	nextID     int64
)

func ok(data any) *C.char {
	return envelope(map[string]any{"data": data})
}

func fail(err error) *C.char {
	return envelope(map[string]any{"error": err.Error()})
}

func failMsg(msg string) *C.char {
	return envelope(map[string]any{"error": msg})
}

func envelope(m map[string]any) *C.char {
	b, err := json.Marshal(m)
	if err != nil {
		return C.CString(`{"error":"internal: failed to encode response"}`)
	}
	return C.CString(string(b))
}

//export NodoraCompile
func NodoraCompile(src *C.char) *C.char {
	c := compiler.NewCompiler()
	ruleset, err := c.Compile(C.GoString(src))
	if err != nil {
		return fail(err)
	}
	return ok(ruleset)
}

//export NodoraNewEvaluator
func NodoraNewEvaluator(rulesetJSON *C.char) *C.char {
	var ruleset nir.Ruleset
	if err := json.Unmarshal([]byte(C.GoString(rulesetJSON)), &ruleset); err != nil {
		return failMsg("failed to parse ruleset: " + err.Error())
	}

	mu.Lock()
	nextID++
	id := nextID
	evaluators[id] = evaluator.NewEvaluator(&ruleset)
	mu.Unlock()

	return ok(id)
}

//export NodoraEvaluate
func NodoraEvaluate(id C.longlong, ruleName *C.char, inputJSON *C.char) *C.char {
	mu.Lock()
	ev, found := evaluators[int64(id)]
	mu.Unlock()
	if !found {
		return failMsg("evaluator not found")
	}

	var input core.ValueMap
	in := C.GoString(inputJSON)
	if in != "" {
		if err := json.Unmarshal([]byte(in), &input); err != nil {
			return failMsg("failed to parse input: " + err.Error())
		}
	}
	if input == nil {
		input = core.ValueMap{}
	}

	result, err := ev.EvaluateRule(C.GoString(ruleName), input)
	if err != nil {
		return fail(err)
	}
	return ok(result)
}

//export NodoraDestroyEvaluator
func NodoraDestroyEvaluator(id C.longlong) {
	mu.Lock()
	delete(evaluators, int64(id))
	mu.Unlock()
}

//export NodoraFree
func NodoraFree(p *C.char) {
	C.free(unsafe.Pointer(p))
}

func main() {}
