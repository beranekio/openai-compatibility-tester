package main

import "testing"

func TestListenHost(t *testing.T) {
	cases := []struct{ in, want string }{
		{":8080", "127.0.0.1:8080"},
		{"0.0.0.0:8080", "127.0.0.1:8080"},
		{"[::]:8080", "127.0.0.1:8080"},
		{"127.0.0.1:8080", "127.0.0.1:8080"},
		{"127.0.0.1:9090", "127.0.0.1:9090"},
		{"[::1]:8080", "[::1]:8080"},
		{"localhost:8080", "localhost:8080"},
		// A malformed address is returned unchanged.
		{"no-port", "no-port"},
	}
	for _, c := range cases {
		if got := listenHost(c.in); got != c.want {
			t.Errorf("listenHost(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
