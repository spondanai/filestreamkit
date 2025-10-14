// Package zipstream provides streaming helpers to create a zip archive and
// send it as base64 without buffering the entire archive in memory.
//
// Use StreamZipToBase64 for convenience or StreamZipToBase64Writer to stream
// directly to an io.Writer with optional context for cancellation.
package zipstream
