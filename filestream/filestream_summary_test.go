package filestream

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBenchmarkSummary(t *testing.T) {
	sizes := []int{1 << 20, 5 << 20, 20 << 20} // keep smaller for CI
	iterations := 2
	dir := t.TempDir()

	type row struct {
		size                     int
		normal, adaptive, writer time.Duration
	}
	results := make([]row, 0, len(sizes))

	buildFile := func(size int) string {
		name := fmt.Sprintf("sum_%d.dat", size)
		path := filepath.Join(dir, name)
		if info, err := os.Stat(path); err == nil && int(info.Size()) == size {
			return path
		}
		pattern := []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")
		buf := make([]byte, size)
		for i := 0; i < size; i++ {
			buf[i] = pattern[i%len(pattern)]
		}
		if err := os.WriteFile(path, buf, 0o600); err != nil {
			t.Fatalf("write file: %v", err)
		}
		return path
	}

	normalEncode := func(p string) (string, error) {
		raw, err := os.ReadFile(p)
		if err != nil {
			return "", err
		}
		return base64.StdEncoding.EncodeToString(raw), nil
	}

	for _, size := range sizes {
		path := buildFile(size)
		ref, err := normalEncode(path)
		if err != nil {
			t.Fatalf("ref encode size=%d: %v", size, err)
		}

		if _, err := os.ReadFile(path); err != nil {
			t.Fatalf("warm read: %v", err)
		}

		timeOne := func(fn func() error) time.Duration {
			start := time.Now()
			if err := fn(); err != nil {
				t.Fatalf("run error size=%d: %v", size, err)
			}
			return time.Since(start)
		}

		var sumNormal, sumAdaptive, sumWriter time.Duration
		for i := 0; i < iterations; i++ {
			sumNormal += timeOne(func() error {
				got, err := normalEncode(path)
				if err != nil {
					return err
				}
				if got != ref {
					return fmt.Errorf("normal mismatch")
				}
				return nil
			})
			sumAdaptive += timeOne(func() error {
				got, err := StreamFileToBase64(path)
				if err != nil {
					return err
				}
				if got != ref {
					return fmt.Errorf("adaptive mismatch")
				}
				return nil
			})
			sumWriter += timeOne(func() error {
				if i == 0 {
					var buf bytes.Buffer
					if err := StreamFileToBase64Writer(&buf, path); err != nil {
						return err
					}
					if buf.String() != ref {
						return fmt.Errorf("writer mismatch")
					}
				} else {
					if err := StreamFileToBase64Writer(io.Discard, path); err != nil {
						return err
					}
				}
				return nil
			})
		}
		results = append(results, row{size: size, normal: sumNormal / time.Duration(iterations), adaptive: sumAdaptive / time.Duration(iterations), writer: sumWriter / time.Duration(iterations)})
	}

	t.Logf("\n===== Performance Summary (avg of %d runs) =====", iterations)
	t.Logf("Size(MB)\tNormal(ms)\tAdaptive(ms)\tStream(ms)")
	for _, r := range results {
		nms := float64(r.normal.Microseconds()) / 1000.0
		ams := float64(r.adaptive.Microseconds()) / 1000.0
		wms := float64(r.writer.Microseconds()) / 1000.0
		t.Logf("%d\t%.2f\t%.2f\t%.2f", r.size>>20, nms, ams, wms)
	}
}
