//go:build !wasm

package nir

func (s *SignalExpr) invokeListeners(listeners []SignalListener, args []any) {
	for _, listener := range listeners {
		l := listener
		go l.Invoke(args)
	}
}
