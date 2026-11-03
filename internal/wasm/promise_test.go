//go:build js && wasm

package wasm

import (
	"strings"
	"syscall/js"
	"testing"
	"time"
)

func TestInvokeWithTimeoutAndLateSettlement(t *testing.T) {
	var resolve js.Value
	executor := js.FuncOf(func(_ js.Value, args []js.Value) any { resolve = args[0]; return nil })
	pending := js.Global().Get("Promise").New(executor)
	executor.Release()
	callback := js.FuncOf(func(js.Value, []js.Value) any { return pending })
	defer callback.Release()
	if _, err := InvokeWithTimeout(callback.Value, 5*time.Millisecond); err == nil || !strings.Contains(err.Error(), "deadline") {
		t.Fatalf("hung callback = %v", err)
	}
	// A late resolve must not call Go handlers that have already been released.
	resolve.Invoke("late")
	time.Sleep(5 * time.Millisecond)
	value := js.FuncOf(func(js.Value, []js.Value) any { return "immediate" })
	defer value.Release()
	got, err := InvokeWithTimeout(value.Value, time.Second)
	if err != nil || got.String() != "immediate" {
		t.Fatalf("immediate result = %v, %v", got, err)
	}
}
