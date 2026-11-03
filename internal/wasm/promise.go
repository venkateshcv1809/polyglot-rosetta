//go:build js && wasm

package wasm

import (
	"fmt"
	"syscall/js"
	"time"
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
	return invoke(callback, 0, args...)
}

// InvokeWithTimeout bounds asynchronous host work. Promise.race keeps late
// settlements inside JavaScript, so they cannot invoke released Go callbacks.
// It cannot interrupt a synchronous JavaScript callback or terminate a worker;
// hosts must enforce their own execution deadlines and honor disposal.
func InvokeWithTimeout(callback js.Value, timeout time.Duration, args ...any) (js.Value, error) {
	return invoke(callback, timeout, args...)
}

func invoke(callback js.Value, timeout time.Duration, args ...any) (value js.Value, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("browser execution failed: %v", recovered)
		}
	}()
	pending := callback.Invoke(args...)
	if timeout > 0 {
		var timer js.Value
		var onTimeout js.Func
		executor := js.FuncOf(func(_ js.Value, args []js.Value) any {
			reject := args[1]
			onTimeout = js.FuncOf(func(js.Value, []js.Value) any {
				reject.Invoke(js.Global().Get("Error").New(fmt.Sprintf("host callback exceeded deadline of %v", timeout)))
				return nil
			})
			timer = js.Global().Call("setTimeout", onTimeout, max(1, timeout.Milliseconds()))
			return nil
		})
		deadline := js.Global().Get("Promise").New(executor)
		executor.Release()
		defer onTimeout.Release()
		defer js.Global().Call("clearTimeout", timer)
		pending = js.Global().Get("Promise").Call("race", js.ValueOf([]any{pending, deadline}))
	}
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
