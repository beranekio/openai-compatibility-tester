package main

import (
	"bytes"
	"errors"
	"flag"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/beranekio/openai-compatibility-tester/internal/config"
)

func TestLoadHelpReturnsErrHelp(t *testing.T) {
	_, err := config.Load([]string{"-h"})
	if !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("Load(-h) error = %v, want flag.ErrHelp", err)
	}
}

func TestPrintSuitesMarksDeprecatedSuites(t *testing.T) {
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	t.Cleanup(func() {
		os.Stdout = oldStdout
		_ = r.Close()
	})
	os.Stdout = w

	printSuites()

	if err := w.Close(); err != nil {
		t.Fatalf("close stdout pipe error = %v", err)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("read stdout error = %v", err)
	}
	output := buf.String()
	for _, line := range []string{
		"(deprecated) assistants -",
		"(deprecated) assistants_threads -",
	} {
		if !strings.Contains(output, line) {
			t.Fatalf("printSuites() output missing %q:\n%s", line, output)
		}
	}
}