package suitespec

import (
	"fmt"
	"sort"
	"sync"
)

var (
	mu    sync.RWMutex
	names = make(map[string]struct{})
)

func register(name string) {
	mu.Lock()
	defer mu.Unlock()
	names[name] = struct{}{}
}

// Names returns sorted registered suite names.
func Names() []string {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]string, 0, len(names))
	for name := range names {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// ValidateNames reports whether every name is registered.
func ValidateNames(selected []string) error {
	mu.RLock()
	defer mu.RUnlock()
	for _, name := range selected {
		if _, ok := names[name]; !ok {
			return fmt.Errorf("unknown test suite %q", name)
		}
	}
	return nil
}