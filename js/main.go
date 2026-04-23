//go:build js && wasm

package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"syscall/js"

	"nodora.org/nodora/pkg/compiler"
	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/evaluator"
	"nodora.org/nodora/pkg/nir"
	"nodora.org/nodora/pkg/registry"
	_ "nodora.org/nodora/pkg/registry/all"
	"nodora.org/nodora/pkg/types"
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
	js.Global().Set("__nodoraRegisterFunction", js.FuncOf(registerFunction))

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
		return errorObject(err)
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

	ev.OnSignal(signalName, func(args []any) error {
		jsArgs := make([]any, len(args))
		for i, arg := range args {
			jsArgs[i] = js.ValueOf(arg)
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

func registerFunction(this js.Value, args []js.Value) any {
	if len(args) < 2 {
		return errorObject("expected arguments: spec, callback")
	}

	spec := args[0]
	callback := args[1]

	name := spec.Get("name").String()
	namespace := spec.Get("namespace").String()

	returnType, err := parseTypeString(spec.Get("returnType").String())
	if err != nil {
		return errorObject("invalid returnType: " + err.Error())
	}

	var argSpecs []types.ArgSpec
	argsVal := spec.Get("args")
	length := argsVal.Length()
	argSpecs = make([]types.ArgSpec, length)

	for i := range length {
		argObj := argsVal.Index(i)
		argName := argObj.Get("name").String()

		argType, err := parseTypeString(argObj.Get("type").String())
		if err != nil {
			return errorObject(fmt.Sprintf("invalid args[%d].type: %s", i, err.Error()))
		}

		required := argObj.Get("required").Bool()
		argSpecs[i] = types.ArgSpec{
			Name:     argName,
			Type:     argType,
			Required: required,
		}
	}

	generator := func() types.Func {
		return types.Func{
			Name:       name,
			Args:       argSpecs,
			ReturnType: returnType,
			Fn: func(fnArgs []core.Value) (result core.Value, err error) {
				defer func() {
					if r := recover(); r != nil {
						err = fmt.Errorf("JS function '%s' threw: %v", name, r)
						result = core.U()
					}
				}()

				jsArgs := make([]any, len(fnArgs))
				for i, arg := range fnArgs {
					jsArgs[i] = valueToJS(arg)
				}

				jsResult := callback.Invoke(jsArgs...)

				// reject if the result is a Promise
				if jsResult.Type() == js.TypeObject && !jsResult.Get("then").IsUndefined() {
					return core.U(), fmt.Errorf("async functions are not supported")
				}

				// sync function, return directly
				return jsToValue(jsResult), nil
			},
		}
	}

	if err := registry.Global().Register(namespace, generator); err != nil {
		return errorObject(err.Error())
	}

	return js.Undefined()
}

func errorObject(err any) any {
	jsonBytes, err := json.Marshal(map[string]any{
		"error": err,
	})
	return js.Global().Get("JSON").Call("parse", string(jsonBytes))
}

// converts a JS object to core.ValueMap
func jsToValueMap(v js.Value) core.ValueMap {
	if v.IsNull() || v.IsUndefined() {
		return core.ValueMap{}
	}

	result := make(core.ValueMap)
	keys := js.Global().Get("Object").Call("keys", v)
	length := keys.Length()

	for i := range length {
		key := keys.Index(i).String()
		value := v.Get(key)
		result[key] = jsToValue(value)
	}

	return result
}

// converts a JS value to core.Value
func jsToValue(v js.Value) core.Value {
	switch v.Type() {
	case js.TypeNull:
		return core.V(nil)
	case js.TypeUndefined:
		return core.U()
	case js.TypeBoolean:
		return core.V(v.Bool())
	case js.TypeNumber:
		return core.V(v.Float())
	case js.TypeString:
		return core.V(v.String())
	case js.TypeObject:
		if v.InstanceOf(js.Global().Get("Array")) {
			length := v.Length()
			arr := make([]core.Value, length)
			for i := range length {
				arr[i] = jsToValue(v.Index(i))
			}
			return core.V(arr)
		}
		return core.V(jsToValueMap(v))
	default:
		return core.V(v.String())
	}
}

// converts a core.Value to a JS value
func valueToJS(v core.Value) js.Value {
	if v.Undefined {
		return js.Undefined()
	}
	if v.Raw == nil {
		return js.Null()
	}

	switch val := v.Raw.(type) {
	case string:
		return js.ValueOf(val)
	case float64:
		return js.ValueOf(val)
	case bool:
		return js.ValueOf(val)
	case []core.Value:
		arr := js.Global().Get("Array").New(len(val))
		for i, item := range val {
			arr.SetIndex(i, valueToJS(item))
		}
		return arr
	case core.ValueMap:
		obj := js.Global().Get("Object").New()
		for k, item := range val {
			obj.Set(k, valueToJS(item))
		}
		return obj
	default:
		return js.ValueOf(fmt.Sprintf("%v", val))
	}
}

// parses a type string (eg. "number") to types.Type
func parseTypeString(s string) (types.Type, error) {
	s = strings.TrimSpace(s)

	// union types: "string|number"
	if strings.Contains(s, "|") {
		parts := strings.Split(s, "|")
		members := make([]types.Type, len(parts))
		for i, p := range parts {
			t, err := parseTypeString(p)
			if err != nil {
				return nil, err
			}
			members[i] = t
		}
		return types.NewUnionType(members...), nil
	}

	// array types: "array<string>", "array"
	if strings.HasPrefix(s, "array") {
		if s == "array" {
			return types.NewArrayType(types.AnyType), nil
		}
		if strings.HasPrefix(s, "array<") && strings.HasSuffix(s, ">") {
			inner := s[6 : len(s)-1]
			elem, err := parseTypeString(inner)
			if err != nil {
				return nil, err
			}
			return types.NewArrayType(elem), nil
		}
		return nil, fmt.Errorf("invalid array type: %s", s)
	}

	switch s {
	case "string":
		return types.StringType, nil
	case "number":
		return types.NumberType, nil
	case "bool":
		return types.BoolType, nil
	case "object":
		return types.ObjectType, nil
	case "any":
		return types.AnyType, nil
	default:
		return nil, fmt.Errorf("unknown type: %s", s)
	}
}
