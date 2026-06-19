package suites

import "testing"

func TestDeprecatedSuitesMarkedDeprecated(t *testing.T) {
	for _, suite := range All() {
		switch suite.Name() {
		case "assistants", "assistants_threads":
			if !IsDeprecated(suite) {
				t.Fatalf("suite %q should be deprecated", suite.Name())
			}
		default:
			if IsDeprecated(suite) {
				t.Fatalf("suite %q should not be deprecated", suite.Name())
			}
		}
	}
}