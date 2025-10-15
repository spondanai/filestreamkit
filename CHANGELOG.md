# Changelog

All notable changes to this project will be documented in this file.

## v1.0.0 - 2025-10-15

- First stable release.
- Public API stabilized for filestream and zipstream.
- Security hardening: path traversal prevention (SafeJoin), zip entry validation to prevent zip-slip, context-aware streaming for cancellation.
- CI matrix across Linux/macOS/Windows; release automation on tags.
- Documentation overhaul: README, EXAMPLES, SECURITY, CONTRIBUTING; releasing guide in .github/RELEASING.md.

## v0.1.x - 2025-10

- Initial public releases and iterative improvements.
