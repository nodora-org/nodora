//go:build js && wasm

package main

import (
	"encoding/json"
	"strconv"
	"syscall/js"

	"nodora.org/nodora/pkg/compiler"
	"nodora.org/nodora/pkg/evaluator"
	"nodora.org/nodora/pkg/nir"
)

var (
	evaluators  = make(map[int]*evaluator.Evaluator)
	evaluatorID = 0
)

func main() {
	js.Global().Set("__nodoraCompile", js.FuncOf(compile))
	js.Global().Set("__nodoraCreateEvaluator", js.FuncOf(createEvaluator))
	js.Global().Set("__nodoraEvaluate", js.FuncOf(evaluateRule))
	js.Global().Set("__nodoraOnSignal", js.FuncOf(registerCallback))
	js.Global().Set("__nodoraDestroy", js.FuncOf(destroyEvaluator))

	select {} // keep alive
}

func compile(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		return errorObject("expected arguments: src")
	}

	src := args[0].String()
	c := compiler.NewCompiler()

	prog, err := c.Compile(src)
	if err != nil {
		return errorObject(err.Error())
	}

	jsonBytes, err := json.Marshal(prog)
	if err != nil {
		return errorObject(err.Error())
	}

	return js.ValueOf(string(jsonBytes))
}

func createEvaluator(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		return errorObject("expected arguments: programJSON")
	}

	programJSON := args[0].String()

	var program nir.Program
	if err := json.Unmarshal([]byte(programJSON), &program); err != nil {
		return errorObject("failed to parse program: " + err.Error())
	}

	ev := evaluator.NewEvaluator(&program)

	evaluatorID++
	id := evaluatorID
	evaluators[id] = ev

	return js.ValueOf(id)
}

func evaluateRule(this js.Value, args []js.Value) any {
	if len(args) < 3 {
		return errorObject("expected arguments: evaluatorId, ruleName, input")
	}

	id := args[0].Int()
	ruleName := args[1].String()
	inputJS := args[2]

	ev, ok := evaluators[id]
	if !ok {
		return errorObject("evaluator not found: " + strconv.Itoa(id))
	}

	input := jsToValueMap(inputJS)
	result, err := ev.EvaluateRule(ruleName, input)
	if err != nil {
		return errorObject(err.Error())
	}

	jsonBytes, _ := json.Marshal(result)
	jsonObj := js.Global().Get("JSON")

	return jsonObj.Call("parse", string(jsonBytes))
}

func registerCallback(this js.Value, args []js.Value) any {
	if len(args) < 3 {
		return errorObject("expected arguments: evaluatorId, signalName, callback")
	}

	id := args[0].Int()
	signalName := args[1].String()
	callback := args[2]

	ev, ok := evaluators[id]
	if !ok {
		return errorObject("evaluator not found: " + strconv.Itoa(id))
	}

	ev.OnSignalFunc(signalName, func(args []nir.Value) error {
		jsArgs := make([]any, len(args))
		for i, arg := range args {
			jsArgs[i] = valueToJS(arg)
		}
		callback.Invoke(jsArgs...)
		return nil
	})

	return js.Undefined()
}

func destroyEvaluator(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		return errorObject("expected arguments: evaluatorId")
	}

	id := args[0].Int()
	delete(evaluators, id)

	return js.Undefined()
}

// converts a JS object to nir.ValueMap
func jsToValueMap(v js.Value) nir.ValueMap {
	if v.IsNull() || v.IsUndefined() {
		return nir.ValueMap{}
	}

	result := make(nir.ValueMap)
	keys := js.Global().Get("Object").Call("keys", v)
	length := keys.Length()

	for i := range length {
		key := keys.Index(i).String()
		value := v.Get(key)
		result[key] = jsToValue(value)
	}

	return result
}

// converts a JS value to nir.Value
func jsToValue(v js.Value) nir.Value {
	switch v.Type() {
	case js.TypeNull, js.TypeUndefined:
		return nil
	case js.TypeBoolean:
		return v.Bool()
	case js.TypeNumber:
		return v.Float()
	case js.TypeString:
		return v.String()
	case js.TypeObject:
		if v.InstanceOf(js.Global().Get("Array")) {
			length := v.Length()
			arr := make([]nir.Value, length)
			for i := range length {
				arr[i] = jsToValue(v.Index(i))
			}
			return arr
		}
		return jsToValueMap(v)
	default:
		return v.String()
	}
}

// converts a nir.Value to a JS-compatible value
func valueToJS(v nir.Value) js.Value {
	switch val := v.(type) {
	case nil:
		return js.Null()
	case bool:
		return js.ValueOf(val)
	case int:
		return js.ValueOf(val)
	case int64:
		return js.ValueOf(float64(val))
	case float64:
		return js.ValueOf(val)
	case string:
		return js.ValueOf(val)
	case []nir.Value:
		arr := js.Global().Get("Array").New(len(val))
		for i, item := range val {
			arr.SetIndex(i, valueToJS(item))
		}
		return arr
	case map[string]nir.Value:
		obj := js.Global().Get("Object").New()
		for k, v := range val {
			obj.Set(k, valueToJS(v))
		}
		return obj
	case nir.ValueMap:
		obj := js.Global().Get("Object").New()
		for k, v := range val {
			obj.Set(k, valueToJS(v))
		}
		return obj
	default:
		return js.ValueOf(val)
	}
}

func errorObject(err string) any {
	return js.ValueOf(map[string]any{
		"error": err,
	})
}
