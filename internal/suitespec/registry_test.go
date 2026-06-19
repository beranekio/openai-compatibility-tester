package suitespec

import "testing"

func TestValidateNamesRejectsUnknown(t *testing.T) {
	if err := ValidateNames([]string{"models"}); err != nil {
		t.Fatalf("ValidateNames() error = %v", err)
	}
	if err := ValidateNames([]string{"not-a-suite"}); err == nil {
		t.Fatal("expected error for unknown suite")
	}
}