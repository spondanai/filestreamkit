# FileStreamKit

[![Go Reference](https://pkg.go.dev/badge/github.com/spondanai/FileStreamKit.svg)](https://pkg.go.dev/github.com/spondanai/FileStreamKit)
[![CI](https://github.com/spondanai/FileStreamKit/actions/workflows/ci.yml/badge.svg)](https://github.com/spondanai/FileStreamKit/actions/workflows/ci.yml)

**streamkit** is a Go library for efficient file streaming and packaging, designed for APIs and backend services that need to send files as base64 or zipped archives without excessive memory usage.

## Features

- **filestream**: Stream a single file as base64, either as a string or directly to an `io.Writer` (e.g., HTTP response). Supports adaptive memory usage for small/large files.
- **zipstream**: Zip multiple files and stream the archive as base64, with direct writer support for large archives. Automatically skips missing files (optional), and stores pre-compressed files without recompression for speed.
- **Simple API**: Minimal, easy-to-use functions for encoding, streaming, and directory listing.
- **Performance**: Designed for low RAM usage and high throughput, with benchmarks and tests included.
- **Security**: Path validation and filtering options to help prevent directory traversal and unsafe file access.

## Use Cases

- Sending files over REST/JSON APIs (e.g., download endpoints)
- Packaging files for web clients or mobile apps
- Efficient backend file transfer and archiving
- Avoiding memory spikes when handling large files or many files

## Why streamkit?

- Avoids loading entire files/archives into memory
- Easy integration with Go web frameworks (Fiber, Gin, etc.)
- Handles edge cases: missing files, pre-compressed formats, streaming to HTTP clients
- Well-documented, MIT licensed, ready for production

## Install

```
go get github.com/spondanai/FileStreamKit
```

Then import subpackages:

```
import (
    "github.com/spondanai/FileStreamKit/filestream"
    "github.com/spondanai/FileStreamKit/zipstream"
)
```

### Compatibility and Versioning

- Module path: `github.com/spondanai/FileStreamKit`
- Go versions: tested with Go 1.21+
- Semantic Versioning: stability is indicated by tags (`v0.x` may change; `v1+` follows semver)

### CI and Releases

This repository includes GitHub Actions for CI (build/test) on push/PR and automatic Release notes when pushing tags like `v0.1.0`.

## Quick Start

### filestream: encode one file

```go
b64, err := filestream.StreamFileToBase64("/path/to/file.dat")
if err != nil { /* handle */ }
// use b64
```

Or stream directly to an io.Writer (e.g., HTTP response):

```go
w := ctx.Response().BodyWriter() // example (Fiber)
if err := filestream.StreamFileToBase64Writer(w, "/path/to/file.dat"); err != nil { /* handle */ }
```

Encode multiple files into a map[name]base64:

```go
out, err := filestream.StreamFilesToMap([]filestream.FileEntry{
    {Name: "a.txt", Path: "/tmp/a.txt"},
    {Name: "b.txt", Path: "/tmp/b.txt"},
}, &filestream.Options{SkipMissing: true})
```

### zipstream: zip many files and stream as base64

```go
b64, err := zipstream.StreamZipToBase64([]zipstream.Entry{
    {Name: "docs/a.txt", Path: "/tmp/a.txt"},
    {Name: "docs/b.txt", Path: "/tmp/b.txt"},
}, &zipstream.Options{CompressionLevel: -1, SkipMissing: true})
```

Or stream to a writer with context:

```go
buf := new(bytes.Buffer)
if err := zipstream.StreamZipToBase64Writer(context.Background(), buf, entries, &zipstream.Options{}); err != nil { /* handle */ }
```

## Security and performance notes

- Large files: prefer Writer APIs to avoid building giant strings in RAM (use `StreamFileToBase64Writer` and `StreamZipToBase64Writer`). Combine with `context.Context` for timeouts/cancellation.
- Concurrency: for many files, stream sequentially; consider batching at the app layer if needed.
- Path safety: always use `SafeJoin()` or set `Options.BaseDir` to prevent directory traversal when paths come from users.
- Zip entry safety: zip entry names must be safe relative paths (no absolute paths, no `..` traversal, no drive letters/colons). The library validates and will error on unsafe entry names to prevent zip-slip.
- Pre-compressed files (e.g., .jpg, .mp4) are stored without recompression to save CPU; tune `CompressionLevel` for your workload.
- Application limits: enforce limits at the application layer (max number of files, per-file size, and total archive size) to prevent resource exhaustion and DoS.
- Error hygiene: avoid leaking internal paths in API errors; log detailed errors internally, return generic messages to clients.

Example (safe usage):

```go
// Restrict all file access to a safe base directory
opts := &zipstream.Options{
    BaseDir:     "/var/safe/uploads",
    SkipMissing: true,
    // Optional: filter to whitelist certain extensions
    Filter: func(e zipstream.Entry) bool {
        ext := strings.ToLower(filepath.Ext(e.Name))
        switch ext { case ".txt", ".pdf", ".png", ".jpg":
            return true
        }
        return false
    },
}

ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
defer cancel()

var buf bytes.Buffer
if err := zipstream.StreamZipToBase64Writer(ctx, &buf, entries, opts); err != nil {
    // Handle error safely without exposing internal paths
}
```

## Documentation

Docs: https://pkg.go.dev/github.com/spondanai/FileStreamKit

## License

MIT