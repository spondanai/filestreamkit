package zipstream

import (
	"bytes"
	"compress/flate"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestStreamZipToBase64_Basic(t *testing.T) {
	dir := t.TempDir()
	p1 := filepath.Join(dir, "a.txt")
	p2 := filepath.Join(dir, "b.txt")
	if err := os.WriteFile(p1, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p2, []byte("world"), 0o600); err != nil {
		t.Fatal(err)
	}
	b64, err := StreamZipToBase64([]Entry{{Name: "data/a.txt", Path: p1}, {Name: "data/b.txt", Path: p2}}, &Options{CompressionLevel: flate.DefaultCompression})
	if err != nil {
		t.Fatalf("zip err: %v", err)
	}
	if len(b64) == 0 {
		t.Fatalf("empty output")
	}
}

func TestStreamZipToBase64Writer_SkipMissing(t *testing.T) {
	dir := t.TempDir()
	p1 := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(p1, []byte("a"), 0o600); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	err := StreamZipToBase64Writer(context.Background(), &buf, []Entry{{Name: "a.txt", Path: p1}, {Name: "nope.txt", Path: filepath.Join(dir, "nope.txt")}}, &Options{SkipMissing: true})
	if err != nil {
		t.Fatalf("writer err: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatalf("empty writer output")
	}
}
