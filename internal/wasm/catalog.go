//go:build js && wasm

package wasm

import catalogbundle "polyglot-rosetta/internal/catalog/bundle"

// CatalogStore uses the transport-independent catalog activation store.
type CatalogStore = catalogbundle.Store
