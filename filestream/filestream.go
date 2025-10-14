package filestream

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// FileEntry describes a file to encode.
type FileEntry struct {
	Name string
	Path string
}

// Options controls multi-file streaming behavior.
type Options struct {
	SkipMissing bool
	Filter      func(FileEntry) bool
	ChunkSize   int
	// BaseDir: if set, StreamFilesToMap will resolve Entry.Path as a relative
	// path under this directory using SafeJoin to prevent path traversal.
	BaseDir string
}

// StreamFileToBase64 returns the base64 encoding of the file at path.
func StreamFileToBase64(path string) (string, error) {
	const smallFileThreshold = 8 << 20 // 8MB
	fi, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if fi.Size() <= smallFileThreshold {
		b, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		return base64.StdEncoding.EncodeToString(b), nil
	}

	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	origSize := fi.Size()
	encodedLen := ((origSize + 2) / 3) * 4
	var capHint int
	if encodedLen > int64(^uint(0)>>1) {
		capHint = 0
	} else {
		capHint = int(encodedLen)
	}
	buf := bytes.NewBuffer(make([]byte, 0, capHint))
	enc := base64.NewEncoder(base64.StdEncoding, buf)

	copyBuf := make([]byte, 256<<10)
	if _, err := io.CopyBuffer(enc, f, copyBuf); err != nil {
		enc.Close()
		return "", fmt.Errorf("copy file: %w", err)
	}
	if err := enc.Close(); err != nil {
		return "", fmt.Errorf("close encoder: %w", err)
	}
	return buf.String(), nil
}

// StreamFileToBase64Writer streams the base64 of the file into an io.Writer.
func StreamFileToBase64Writer(w io.Writer, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := base64.NewEncoder(base64.StdEncoding, w)
	if _, err := io.Copy(enc, f); err != nil {
		enc.Close()
		return err
	}
	return enc.Close()
}

// StreamFileToBase64WriterContext streams base64 with cancellation support.
func StreamFileToBase64WriterContext(ctx context.Context, w io.Writer, path string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := base64.NewEncoder(base64.StdEncoding, w)
	defer enc.Close()
	buf := make([]byte, 256<<10)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		n, rerr := f.Read(buf)
		if n > 0 {
			if _, werr := enc.Write(buf[:n]); werr != nil {
				return werr
			}
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			return rerr
		}
	}
	return nil
}

// SafeJoin joins baseDir and relPath, ensuring the result stays within baseDir.
func SafeJoin(baseDir, relPath string) (string, error) {
	if baseDir == "" {
		return relPath, nil
	}
	combined := filepath.Join(baseDir, relPath)
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return "", err
	}
	absCombined, err := filepath.Abs(combined)
	if err != nil {
		return "", err
	}
	baseWithSep := absBase
	if !strings.HasSuffix(baseWithSep, string(os.PathSeparator)) {
		baseWithSep += string(os.PathSeparator)
	}
	if absCombined != absBase && !strings.HasPrefix(absCombined, baseWithSep) {
		return "", fmt.Errorf("path escapes base dir: %s", relPath)
	}
	return absCombined, nil
}

// StreamFileToBase64Safe resolves relPath under baseDir then encodes.
func StreamFileToBase64Safe(baseDir, relPath string) (string, error) {
	p, err := SafeJoin(baseDir, relPath)
	if err != nil {
		return "", err
	}
	return StreamFileToBase64(p)
}

// StreamFileToBase64WriterSafe resolves relPath under baseDir then streams.
func StreamFileToBase64WriterSafe(w io.Writer, baseDir, relPath string) error {
	p, err := SafeJoin(baseDir, relPath)
	if err != nil {
		return err
	}
	return StreamFileToBase64Writer(w, p)
}

// StreamFilesToMap encodes multiple files and returns map[name]base64.
func StreamFilesToMap(entries []FileEntry, opts *Options) (map[string]string, error) {
	if opts == nil {
		opts = &Options{}
	}
	out := make(map[string]string, len(entries))
	for _, e := range entries {
		if e.Name == "" {
			return nil, errors.New("file entry name empty")
		}
		if opts.Filter != nil && !opts.Filter(e) {
			continue
		}
		pathToRead := e.Path
		if opts.BaseDir != "" {
			var err error
			pathToRead, err = SafeJoin(opts.BaseDir, e.Path)
			if err != nil {
				return nil, fmt.Errorf("unsafe path %s: %w", e.Path, err)
			}
		}
		b64, err := StreamFileToBase64(pathToRead)
		if err != nil {
			if opts.SkipMissing && errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, fmt.Errorf("encode %s (%s): %w", e.Name, pathToRead, err)
		}
		out[e.Name] = b64
	}
	return out, nil
}

// BuildEntriesFromDir returns FileEntry list for files inside dir (non-recursive).
func BuildEntriesFromDir(dir string) ([]FileEntry, error) {
	list, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var entries []FileEntry
	for _, d := range list {
		if d.IsDir() {
			continue
		}
		p := filepath.Join(dir, d.Name())
		entries = append(entries, FileEntry{Name: d.Name(), Path: p})
	}
	return entries, nil
}
