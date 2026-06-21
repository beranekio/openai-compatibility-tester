// Package main runs the in-process mock server as a standalone HTTP service,
// printing its base URL (including /v1) to stdout on startup. Intended for
// local manual testing and CI smoke tests of the tester container.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/beranekio/openai-compatibility-tester/internal/mockserver"
)

func main() {
	s := mockserver.New()
	defer s.Close()
	fmt.Println(s.BaseURL())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}
