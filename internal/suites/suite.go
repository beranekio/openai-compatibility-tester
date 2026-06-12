package suites

import (
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"
	"github.com/openai/openai-go/v3"
)

// Suite exercises one area of the OpenAI API through the official Go SDK.
type Suite interface {
	Name() string
	Description() string
	Run(ctx context.Context, client openai.Client, cfg *config.Config) error
}

// All returns every registered compatibility suite.
func All() []Suite {
	return []Suite{
		Models{},
		ChatCompletions{},
		ChatCompletionsStream{},
		Completions{},
		Embeddings{},
		Responses{},
		ResponsesStream{},
	}
}

// ByName maps suite names to implementations.
func ByName() map[string]Suite {
	registry := make(map[string]Suite, len(All()))
	for _, suite := range All() {
		registry[suite.Name()] = suite
	}
	return registry
}

// ValidateNames reports whether every name is a registered suite.
func ValidateNames(names []string) error {
	registry := ByName()
	for _, name := range names {
		if _, ok := registry[name]; !ok {
			return fmt.Errorf("unknown test suite %q (use --list-suites to see options)", name)
		}
	}
	return nil
}

// Names returns sorted suite names for display.
func Names() []string {
	all := All()
	names := make([]string, len(all))
	for i, suite := range all {
		names[i] = suite.Name()
	}
	return names
}