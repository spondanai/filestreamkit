package filestream

import (
	"bytes"
	"encoding/base64"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func writeTempFile(t *testing.T, dir, name string, data []byte) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, data, 0o600); err != nil {
		t.Fatalf("write %s: %v", p, err)
	}
	return p
}

func TestStreamFileToBase64_SmallFile(t *testing.T) {
	dir := t.TempDir()
	p := writeTempFile(t, dir, "a.txt", []byte("hello world"))

	got, err := StreamFileToBase64(p)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	raw, _ := os.ReadFile(p)
	want := base64.StdEncoding.EncodeToString(raw)
	if got != want {
		t.Fatalf("mismatch: got %q want %q", got, want)
	}
}

func TestStreamFileToBase64Writer_Match(t *testing.T) {
	dir := t.TempDir()
	p := writeTempFile(t, dir, "b.txt", bytes.Repeat([]byte("x"), 4096))

	buf := new(bytes.Buffer)
	if err := StreamFileToBase64Writer(buf, p); err != nil {
		t.Fatalf("writer err: %v", err)
	}
	want, _ := StreamFileToBase64(p)
	if buf.String() != want {
		t.Fatalf("writer mismatch")
	}
}

func TestStreamFilesToMap_SkipMissing(t *testing.T) {
	dir := t.TempDir()
	p1 := writeTempFile(t, dir, "ok.txt", []byte("ok"))
	// p2 does not exist
	out, err := StreamFilesToMap([]FileEntry{{Name: "ok", Path: p1}, {Name: "missing", Path: filepath.Join(dir, "nope.txt")}}, &Options{SkipMissing: true})
	if err != nil {
		t.Fatalf("map err: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(out))
	}
	if _, ok := out["ok"]; !ok {
		t.Fatalf("expected key 'ok'")
	}
}

func TestBuildEntriesFromDir(t *testing.T) {
	dir := t.TempDir()
	_ = writeTempFile(t, dir, "a.txt", []byte("a"))
	_ = writeTempFile(t, dir, "b.txt", []byte("b"))
	// nested directory should be ignored
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	es, err := BuildEntriesFromDir(dir)
	if err != nil {
		t.Fatalf("build entries: %v", err)
	}
	if len(es) != 2 {
		t.Fatalf("expected 2 files, got %d", len(es))
	}
}

func BenchmarkStreamFileToBase64_Small(b *testing.B) {
	dir := b.TempDir()
	p := writeTempFile(&testing.T{}, dir, "bench.txt", bytes.Repeat([]byte("z"), 32*1024))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := StreamFileToBase64(p); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStreamFileToBase64Writer_Small(b *testing.B) {
	dir := b.TempDir()
	p := writeTempFile(&testing.T{}, dir, "benchw.txt", bytes.Repeat([]byte("z"), 32*1024))
	b.ReportAllocs()
	buf := new(bytes.Buffer)
	for i := 0; i < b.N; i++ {
		buf.Reset()
		if err := StreamFileToBase64Writer(io.Discard, p); err != nil {
			b.Fatal(err)
		}
	}
}
