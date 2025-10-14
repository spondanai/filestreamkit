# Security Policy

## Supported Versions

We provide security updates for the following versions of StreamKit:

| Version | Supported |
|--------:|:---------:|
|   1.x.x |     ✔     |
|    <1.0 |     ✖     |

## Built-in Protections

- Path traversal prevention via `SafeJoin(baseDir, relPath)` and `Options.BaseDir`
- Zip entry validation to prevent zip-slip (no absolute paths, no `..` segments, no drive letters/colons)
- Context-aware streaming for cancellation and timeouts
- Pre-compressed file detection to avoid unnecessary CPU usage

## Best Practices for Application Developers

1. Always restrict file access to a base directory

```go
safePath, err := filestream.SafeJoin("/var/uploads", userInput)
if err != nil { return err }
```

Or set `BaseDir` in options:

```go
opts := &zipstream.Options{ BaseDir: "/var/uploads", SkipMissing: true }
```

2. Validate and sanitize inputs
- Reject paths containing `..`, absolute paths, or reserved characters
- Whitelist file extensions where possible

3. Use Writer APIs and context for large workloads

```go
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
defer cancel()
_ = zipstream.StreamZipToBase64Writer(ctx, w, entries, &zipstream.Options{})
```

4. Enforce application-level limits
- Maximum number of files per request
- Maximum per-file size and total archive size
- Rate limiting for endpoints handling file operations

5. Handle errors safely
- Log detailed errors internally
- Return generic messages to clients to avoid leaking internal paths

## Reporting a Vulnerability

Please do not open public issues for security vulnerabilities.

Email: security@streamkit.dev

Include:
- Affected version(s)
- Reproduction steps and proof-of-concept
- Expected vs. actual behavior
- Environment details (OS, Go version)

We aim to acknowledge reports within 48 hours and provide status updates within 7 days.
