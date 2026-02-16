//go:build wasm

package nir

func (s *SignalExpr) invokeListeners(listeners []SignalListener, args []Value) {
	for _, listener := range listeners {
		listener.Invoke(args) // synchronous invokation
	}
}
