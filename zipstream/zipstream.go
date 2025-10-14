package zipstream

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Entry struct {
	Name string
	Path string
}

type Options struct {
	CompressionLevel int
	SkipMissing      bool
	Filter           func(e Entry) bool
	NowFunc          func() time.Time
	// BaseDir: if set, resolve Entry.Path relative to this directory using SafeJoin.
	BaseDir string
}

var defaultCompressedExt = map[string]struct{}{
	".zip": {}, ".rar": {}, ".7z": {}, ".jpg": {}, ".jpeg": {}, ".png": {}, ".gif": {}, ".pdf": {},
	".mp4": {}, ".mov": {}, ".avi": {}, ".mkv": {}, ".webp": {}, ".docx": {}, ".xlsx": {}, ".pptx": {},
}

func isPreCompressed(name string) bool {
	_, ok := defaultCompressedExt[filepath.Ext(name)]
	return ok
}

func validateEntries(entries []Entry) error {
	seen := make(map[string]struct{}, len(entries))
	for _, e := range entries {
		if e.Name == "" {
			return errors.New("entry name empty")
		}
		// Prevent zip-slip and unsafe names: no absolute paths, no drive letters, no traversal
		// Normalize to forward slashes for checks
		n := strings.ReplaceAll(e.Name, "\\", "/")
		if strings.HasPrefix(n, "/") || strings.Contains(n, "../") {
			return fmt.Errorf("unsafe entry name: %s", e.Name)
		}
		// On Windows or other systems, disallow drive letters or colon usage in names
		if strings.Contains(e.Name, ":") || filepath.IsAbs(e.Name) {
			return fmt.Errorf("unsafe absolute entry name: %s", e.Name)
		}
		if _, dup := seen[e.Name]; dup {
			return fmt.Errorf("duplicate entry name: %s", e.Name)
		}
		seen[e.Name] = struct{}{}
	}
	return nil
}

// SafeJoin joins baseDir and relPath, ensuring result stays within baseDir.
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

// copyWithContext copies from src to dst using a fixed-size buffer while checking
// the provided context for cancellation between reads and writes.
func copyWithContext(ctx context.Context, dst io.Writer, src io.Reader) error {
	buf := make([]byte, 256<<10) // 256KB buffer
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		n, rerr := src.Read(buf)
		if n > 0 {
			if _, werr := dst.Write(buf[:n]); werr != nil {
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

func StreamZipToBase64(entries []Entry, opts *Options) (string, error) {
	var buf io.ReadWriter = &bytes.Buffer{}
	err := StreamZipToBase64Writer(context.Background(), buf, entries, opts)
	if err != nil {
		return "", err
	}
	return buf.(*bytes.Buffer).String(), nil
}

func StreamZipToBase64Writer(ctx context.Context, w io.Writer, entries []Entry, opts *Options) (err error) {
	if opts == nil {
		opts = &Options{}
	}
	if err = validateEntries(entries); err != nil {
		return
	}
	now := time.Now
	if opts.NowFunc != nil {
		now = opts.NowFunc
	}
	lvl := opts.CompressionLevel
	if lvl < flate.HuffmanOnly || lvl > flate.BestCompression {
		lvl = flate.BestSpeed
	}

	b64 := base64.NewEncoder(base64.StdEncoding, w)
	zw := zip.NewWriter(b64)
	zw.RegisterCompressor(zip.Deflate, func(w io.Writer) (io.WriteCloser, error) { return flate.NewWriter(w, lvl) })

	defer func() {
		if cerr := zw.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("close zip writer: %w", cerr)
		}
		if cerr := b64.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("close base64 encoder: %w", cerr)
		}
	}()

	for _, e := range entries {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if opts.Filter != nil && !opts.Filter(e) {
			continue
		}
		pathToRead := e.Path
		if opts.BaseDir != "" {
			var jerr error
			pathToRead, jerr = SafeJoin(opts.BaseDir, e.Path)
			if jerr != nil {
				return fmt.Errorf("unsafe path %s: %w", e.Path, jerr)
			}
		}
		f, openErr := os.Open(pathToRead)
		if openErr != nil {
			if opts.SkipMissing && errors.Is(openErr, os.ErrNotExist) {
				continue
			}
			return fmt.Errorf("open file failed (%s): %w", pathToRead, openErr)
		}
		func() {
			defer f.Close()
			fi, statErr := f.Stat()
			hdr := &zip.FileHeader{Name: e.Name, Method: zip.Deflate}
			if statErr == nil {
				hdr.Modified = fi.ModTime()
			} else {
				hdr.Modified = now()
			}
			if isPreCompressed(e.Name) {
				hdr.Method = zip.Store
			}
			ew, herr := zw.CreateHeader(hdr)
			if herr != nil {
				err = fmt.Errorf("create zip entry failed (%s): %w", e.Name, herr)
				return
			}
			if cerr := copyWithContext(ctx, ew, f); cerr != nil {
				err = fmt.Errorf("write zip entry failed (%s): %w", e.Name, cerr)
				return
			}
		}()
		if err != nil {
			return err
		}
	}
	return nil
}
