package suitespec

import "testing"

func TestValidateNamesRejectsUnknown(t *testing.T) {
	Register("known-suite")
	if err := ValidateNames([]string{"known-suite"}); err != nil {
		t.Fatalf("ValidateNames() error = %v", err)
	}
	if err := ValidateNames([]string{"not-a-suite"}); err == nil {
		t.Fatal("expected error for unknown suite")
	}
}