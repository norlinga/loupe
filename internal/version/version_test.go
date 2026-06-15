package version

import "testing"

func TestStringFallsBackToDev(t *testing.T) {
	old := Version
	t.Cleanup(func() {
		Version = old
	})
	Version = ""

	if got := String(); got != "dev" {
		t.Fatalf("String() = %q, want dev", got)
	}
}

func TestStringReturnsBuildVersion(t *testing.T) {
	old := Version
	t.Cleanup(func() {
		Version = old
	})
	Version = "1.2.3"

	if got := String(); got != "1.2.3" {
		t.Fatalf("String() = %q, want 1.2.3", got)
	}
}
