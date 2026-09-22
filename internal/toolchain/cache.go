package toolchain

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"polyglot-rosetta/internal/domain/language"
)

// CacheStore manages persistent toolchain state with thread-safe access.
type CacheStore struct {
	mu        sync.RWMutex
	cacheFile string
	entries   map[language.Language]cachedInfo
}

type cachedInfo struct {
	Info      ToolchainInfo `json:"info"`
	UpdatedAt time.Time     `json:"updated_at"`
}

// NewCacheStore initializes a persistent cache, loading existing records from disk if available.
func NewCacheStore(cacheFile string) (*CacheStore, error) {
	store := &CacheStore{
		cacheFile: cacheFile,
		entries:   make(map[language.Language]cachedInfo),
	}
	if cacheFile != "" {
		_ = store.load()
	}
	return store, nil
}

func (s *CacheStore) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return json.Unmarshal(data, &s.entries)
}

// Save flushes the cache entries to disk (thread-safe).
func (s *CacheStore) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveLocked()
}

// saveLocked writes to disk assuming the lock is already held.
func (s *CacheStore) saveLocked() error {
	if s.cacheFile == "" {
		return nil
	}

	dir := filepath.Dir(s.cacheFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s.entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.cacheFile, data, 0644)
}

// Get returns cached toolchain info with zero latency if present.
func (s *CacheStore) Get(lang language.Language) (ToolchainInfo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, ok := s.entries[lang]
	if !ok {
		return ToolchainInfo{}, false
	}
	return entry.Info, true
}

// Set stores or updates toolchain info in the cache.
func (s *CacheStore) Set(info ToolchainInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.entries[info.Language] = cachedInfo{
		Info:      info,
		UpdatedAt: time.Now(),
	}
	_ = s.saveLocked()
}

// Invalidate removes a language entry from the cache, triggering a self-healing re-probe on next access.
func (s *CacheStore) Invalidate(lang language.Language) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.entries, lang)
	_ = s.saveLocked()
}
