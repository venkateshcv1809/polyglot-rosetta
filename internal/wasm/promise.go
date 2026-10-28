//go:build js && wasm

package wasm

import (
	"fmt"
	"syscall/js"
)

// Promise runs work in a goroutine so JavaScript can service asynchronous work.
func Promise(work func() any) js.Value {
	executor := js.FuncOf(func(_ js.Value, args []js.Value) any {
		resolve, reject := args[0], args[1]
		go func() {
			defer func() {
				if recovered := recover(); recovered != nil {
					reject.Invoke(js.Global().Get("Error").New(fmt.Sprint(recovered)))
				}
			}()
			resolve.Invoke(Value(work()))
		}()
		return nil
	})
	defer executor.Release()
	return js.Global().Get("Promise").New(executor)
}

// Invoke calls a JavaScript backend and awaits either its value or its Promise.
func Invoke(callback js.Value, args ...any) (value js.Value, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("browser execution failed: %v", recovered)
		}
	}()
	pending := callback.Invoke(args...)
	type outcome struct {
		value js.Value
		err   error
	}
	completed := make(chan outcome, 1)
	resolve := js.FuncOf(func(_ js.Value, args []js.Value) any {
		completed <- outcome{value: args[0]}
		return nil
	})
	defer resolve.Release()
	reject := js.FuncOf(func(_ js.Value, args []js.Value) any {
		reason := args[0]
		message := reason.String()
		if reason.Type() == js.TypeObject && !reason.IsNull() && reason.Get("message").Type() == js.TypeString {
			message = reason.Get("message").String()
		}
		completed <- outcome{err: fmt.Errorf("browser execution failed: %s", message)}
		return nil
	})
	defer reject.Release()
	js.Global().Get("Promise").Call("resolve", pending).Call("then", resolve, reject)
	result := <-completed
	return result.value, result.err
}
