package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/beranekio/openai-compatibility-tester/internal/config"
	"github.com/beranekio/openai-compatibility-tester/internal/runner"
	"github.com/beranekio/openai-compatibility-tester/internal/suites"
)

func main() {
	cfg, err := config.Load(os.Args[1:])
	if errors.Is(err, flag.ErrHelp) {
		os.Exit(0)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "configuration error: %v\n", err)
		os.Exit(2)
	}

	if cfg.ListSuites {
		printSuites()
		os.Exit(0)
	}

	ctx := context.Background()
	code, err := runner.RunAll(ctx, cfg)
	if err != nil && code == 2 {
		fmt.Fprintf(os.Stderr, "runner error: %v\n", err)
		os.Exit(code)
	}
	os.Exit(code)
}

func printSuites() {
	fmt.Println("Available test suites:")
	for _, suite := range suites.All() {
		name := suite.Name()
		if suites.IsDeprecated(suite) {
			name = "(deprecated) " + name
		}
		fmt.Printf("  %s - %s\n", name, suite.Description())
	}
	fmt.Println()
	fmt.Printf("Default suites: %s\n", strings.Join(config.DefaultSuites, ", "))
}