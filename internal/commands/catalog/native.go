package catalogcmd

import (
	"context"

	"polyglot-rosetta/internal/catalog"
	catalogsync "polyglot-rosetta/internal/catalog/sync"
)

type nativeHost struct {
	options options
	sync    syncCatalog
}

func (h nativeHost) Activate(ctx context.Context) (string, *catalog.Catalog, error) {
	var root string
	var err error
	if h.options.url == "" {
		root, err = catalogsync.LocalDir()
	} else {
		root, err = h.sync(ctx, h.options.url, h.options.cache)
	}
	return root, nil, err
}
