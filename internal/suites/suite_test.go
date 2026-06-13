package suites

import (
	"testing"

	"github.com/beranekio/openai-compatibility-tester/internal/config"
)

func TestFullSuitesMatchesRegisteredSuites(t *testing.T) {
	registered := Names()
	if len(config.FullSuites) != len(registered) {
		t.Fatalf("len(FullSuites) = %d, len(suites.All()) = %d", len(config.FullSuites), len(registered))
	}

	registeredSet := make(map[string]struct{}, len(registered))
	for _, name := range registered {
		registeredSet[name] = struct{}{}
	}
	fullSet := make(map[string]struct{}, len(config.FullSuites))
	for _, name := range config.FullSuites {
		fullSet[name] = struct{}{}
		if _, ok := registeredSet[name]; !ok {
			t.Fatalf("FullSuites contains %q not registered in suites.All()", name)
		}
	}
	for _, name := range registered {
		if _, ok := fullSet[name]; !ok {
			t.Fatalf("suites.All() contains %q missing from FullSuites", name)
		}
	}
}