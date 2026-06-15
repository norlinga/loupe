package observe

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkObserveLargeDirectory(b *testing.B) {
	root := b.TempDir()
	for i := 0; i < 1000; i++ {
		writeBenchmarkFile(b, filepath.Join(root, fmt.Sprintf("file-%04d.txt", i)), "content")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Observe(root, Options{Depth: 1}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkObserveNestedDirectory(b *testing.B) {
	root := b.TempDir()
	for dir := 0; dir < 50; dir++ {
		for file := 0; file < 20; file++ {
			writeBenchmarkFile(b, filepath.Join(root, fmt.Sprintf("dir-%02d", dir), fmt.Sprintf("file-%02d.go", file)), "package bench")
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Observe(root, Options{Depth: 2}); err != nil {
			b.Fatal(err)
		}
	}
}

func writeBenchmarkFile(b *testing.B, path string, content string) {
	b.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		b.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		b.Fatal(err)
	}
}
