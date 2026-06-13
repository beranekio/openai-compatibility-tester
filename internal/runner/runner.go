package runner

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/beranekio/openai-compatibility-tester/internal/config"
	"github.com/beranekio/openai-compatibility-tester/internal/suites"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// Result summarizes a single suite execution.
type Result struct {
	Name    string
	Passed  bool
	Error   error
	Elapsed time.Duration
}

// Runner executes selected compatibility suites against an endpoint.
type Runner struct {
	Client openai.Client
	Config *config.Config
	Output io.Writer
}

// New creates a runner configured for the target endpoint.
func New(cfg *config.Config) *Runner {
	client := openai.NewClient(
		option.WithBaseURL(cfg.BaseURL),
		option.WithAPIKey(cfg.APIKey),
		option.WithMaxRetries(0),
	)

	return &Runner{
		Client: client,
		Config: cfg,
		Output: os.Stdout,
	}
}

// Run executes all configured suites and returns per-suite results.
func (r *Runner) Run(ctx context.Context) ([]Result, error) {
	if err := suites.ValidateNames(r.Config.Suites); err != nil {
		return nil, err
	}

	registry := suites.ByName()
	results := make([]Result, 0, len(r.Config.Suites))

	for _, name := range r.Config.Suites {
		suite := registry[name]

		fmt.Fprintf(r.Output, "==> running suite %q\n", suite.Name())
		start := time.Now()

		suiteCtx, cancel := context.WithTimeout(ctx, r.Config.RequestTimeout)
		err := suite.Run(suiteCtx, r.Client, r.Config)
		cancel()

		result := Result{
			Name:    suite.Name(),
			Passed:  err == nil,
			Error:   err,
			Elapsed: time.Since(start),
		}
		results = append(results, result)

		if result.Passed {
			fmt.Fprintf(r.Output, "    PASS (%s)\n", result.Elapsed.Round(time.Millisecond))
		} else {
			fmt.Fprintf(r.Output, "    FAIL (%s): %v\n", result.Elapsed.Round(time.Millisecond), err)
		}
	}

	return results, nil
}

// ExitCode returns 0 when all suites passed, 1 otherwise.
func ExitCode(results []Result) int {
	for _, result := range results {
		if !result.Passed {
			return 1
		}
	}
	return 0
}

// FormatSummary renders a human-readable report.
func FormatSummary(results []Result) string {
	var b strings.Builder
	passed := 0
	for _, result := range results {
		if result.Passed {
			passed++
		}
	}
	fmt.Fprintf(&b, "\n%d of %d suites passed\n", passed, len(results))
	for _, result := range results {
		status := "PASS"
		if !result.Passed {
			status = "FAIL"
		}
		fmt.Fprintf(&b, "  [%s] %s", status, result.Name)
		if result.Error != nil {
			fmt.Fprintf(&b, ": %v", result.Error)
		}
		fmt.Fprintln(&b)
	}
	return b.String()
}

// FirstError returns the first suite error, if any.
func FirstError(results []Result) error {
	for _, result := range results {
		if result.Error != nil {
			return result.Error
		}
	}
	return nil
}

// RunAll is a convenience helper for CLI usage.
func RunAll(ctx context.Context, cfg *config.Config) (int, error) {
	runner := New(cfg)
	results, err := runner.Run(ctx)
	if err != nil {
		return 2, err
	}

	fmt.Print(FormatSummary(results))
	code := ExitCode(results)
	if code != 0 {
		return code, FirstError(results)
	}
	return 0, nil
}

