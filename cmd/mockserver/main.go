package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/beranekio/openai-compatibility-tester/internal/mockserver"
)

// Standalone mock OpenAI-compatible server. Intended for testing gateways and
// SDK clients without a real backend: state is in memory, no authentication is
// enforced, and all data is lost on restart. Point a client at
// http://<addr>/v1 (e.g. OPENAI_BASE_URL=http://127.0.0.1:8080/v1).
func main() {
	addr := flag.String("addr", envDefault("MOCK_ADDR", ":8080"), "listen address (env: MOCK_ADDR)")
	flag.Parse()

	srv := &http.Server{
		Addr:              *addr,
		Handler:           mockserver.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("mockserver listening on %s (base URL: http://%s/v1)", *addr, listenHost(*addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Printf("mockserver shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
		os.Exit(1)
	}
}

func envDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// listenHost returns the address as it should appear in a printed URL, using
// localhost for the unspecified host. net.JoinHostPort brackets IPv6 hosts.
func listenHost(addr string) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	if host == "" {
		host = "127.0.0.1"
	}
	return net.JoinHostPort(host, port)
}
