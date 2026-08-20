package app

import "testing"

// The fixture carries a test because `go test ./...` reports success against a
// package with no test files at all. Without one, CheckGoTest's own coverage
// would prove only that the command exits zero — not that it compiled a test
// binary and ran it.
func TestGreeting(t *testing.T) {
	if got, want := Greeting(), "hello"; got != want {
		t.Fatalf("Greeting() = %q, want %q", got, want)
	}
}
