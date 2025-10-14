// Package filestream provides helpers to stream files as base64.
//
// It supports:
//   - Encoding a single file to a base64 string (adaptive memory)
//   - Streaming base64 directly to an io.Writer (for HTTP responses)
//   - Building entries from a directory and encoding multiple files to a map
package filestream
