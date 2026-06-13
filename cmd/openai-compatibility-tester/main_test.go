package main

import (
	"errors"
	"flag"
	"testing"

	"github.com/beranekio/openai-compatibility-tester/internal/config"
)

func TestLoadHelpReturnsErrHelp(t *testing.T) {
	_, err := config.Load([]string{"-h"})
	if !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("Load(-h) error = %v, want flag.ErrHelp", err)
	}
}