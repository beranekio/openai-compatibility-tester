package suites

import (
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"
	"github.com/beranekio/openai-compatibility-tester/internal/testutil"

	"github.com/openai/openai-go/v3"
)

// ContentProvenanceChecks verifies POST /v1/content_provenance_checks.
type ContentProvenanceChecks struct{}

func (ContentProvenanceChecks) Name() string { return "content_provenance_checks" }
func (ContentProvenanceChecks) Description() string {
	return "Content provenance checks (POST /v1/content_provenance_checks)"
}

func (ContentProvenanceChecks) Run(ctx context.Context, client openai.Client, _ *config.Config) error {
	resp, err := client.ContentProvenanceChecks.New(ctx, openai.ContentProvenanceCheckNewParams{
		File: testutil.SmallPNGReader(),
	})
	if err != nil {
		return fmt.Errorf("content provenance check failed: %w", err)
	}
	return validateContentProvenanceCheck("content_provenance_checks", resp)
}

func validateContentProvenanceCheck(suite string, resp *openai.ContentProvenanceCheck) error {
	if resp == nil {
		return fail(suite, "response is nil")
	}
	if string(resp.Object) != "content_provenance_check" {
		return fail(suite, fmt.Sprintf("object is %q, want content_provenance_check", resp.Object))
	}
	if !resp.JSON.CreatedAt.Valid() {
		return fail(suite, "response missing created_at")
	}
	if !resp.JSON.Results.Valid() {
		return fail(suite, "response missing results")
	}
	if len(resp.Results) == 0 {
		return fail(suite, "results is empty")
	}
	for i, result := range resp.Results {
		if result.Type != "c2pa" && result.Type != "synthid" {
			return fail(suite, fmt.Sprintf("results[%d] type is %q, want c2pa or synthid", i, result.Type))
		}
		if result.Outcome != "detected" && result.Outcome != "not_detected" {
			return fail(suite, fmt.Sprintf("results[%d] outcome is %q, want detected or not_detected", i, result.Outcome))
		}
	}
	return nil
}
