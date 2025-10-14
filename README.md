# FileStreamKit

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

- Large files: prefer Writer APIs to avoid building giant strings in RAM.
- Concurrency: for many files, stream sequentially; see roadmap for batch concurrency.
- Path safety: validate/whitelist paths, or provide a base directory and join to prevent traversal.
- Pre-compressed files (e.g., .jpg, .mp4) are stored without recompression to save CPU.

## Documentation

Docs: https://pkg.go.dev/github.com/spondanai/FileStreamKit

## License

MIT