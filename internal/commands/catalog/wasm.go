//go:build js && wasm

package catalogcmd

import (
	"context"
	"fmt"
	"syscall/js"

	"polyglot-rosetta/internal/catalog"
	"polyglot-rosetta/internal/wasm"
)

// WASM returns a binding that loads or reads the shared in-memory catalog.
func WASM(store *wasm.CatalogStore) wasm.Binding {
	return wasm.Binding{Name: "catalog", Help: wasmHelp(), Handler: func(_ js.Value, args []js.Value) any {
		selection := ""
		selectConcept := false
		if len(args) == 1 && args[0].Type() == js.TypeString {
			selectConcept = true
			selection = args[0].String()
			args = nil
		}
		result, _ := Catalog(context.Background(), browserHost{store: store, args: args})
		if selectConcept {
			result = Resolve(result, selection)
		}
		return wasm.Value(result)
	}}
}

type browserHost struct {
	store *wasm.CatalogStore
	args  []js.Value
}

func (h browserHost) Activate(context.Context) (string, *catalog.Catalog, error) {
	if err := loadWASMCatalog(h.store, h.args); err != nil {
		return "", nil, err
	}
	loaded, err := h.store.Catalog()
	return "", loaded, err
}

func loadWASMCatalog(store *wasm.CatalogStore, args []js.Value) error {
	if len(args) == 0 {
		return nil
	}
	if len(args) == 1 {
		index, err := wasm.Bytes(args[0])
		if err != nil {
			return err
		}
		return store.LoadIndex(index)
	}
	if len(args) != 3 || args[2].Type() != js.TypeObject || args[2].IsNull() || js.Global().Get("Array").Call("isArray", args[2]).Bool() {
		return fmt.Errorf("catalog accepts no arguments, a concept ID, index bytes, or index bytes, template bytes, and a category archive object")
	}
	index, err := wasm.Bytes(args[0])
	if err != nil {
		return fmt.Errorf("read catalog index: %w", err)
	}
	templates, err := wasm.Bytes(args[1])
	if err != nil {
		return fmt.Errorf("read templates: %w", err)
	}
	keys := js.Global().Get("Object").Call("keys", args[2])
	archives := make(map[string][]byte, keys.Length())
	for i := 0; i < keys.Length(); i++ {
		name := keys.Index(i).String()
		data, err := wasm.Bytes(args[2].Get(name))
		if err != nil {
			return fmt.Errorf("read archive %q: %w", name, err)
		}
		archives[name] = data
	}
	return store.Load(index, templates, archives)
}
