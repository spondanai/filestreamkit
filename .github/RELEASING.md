This document describes how to cut a release for the filestreamkit module.

1) Pre-flight
- Ensure main is green: CI passing on Linux/macOS/Windows for supported Go versions.
- Update CHANGELOG.md with a new section for the version.
- Verify docs (README, EXAMPLES) reference the current module path: github.com/spondanai/filestreamkit.

2) Versioning
- Follow SemVer: vMAJOR.MINOR.PATCH
- MINOR for new, backward-compatible features; PATCH for fixes; MAJOR for breaking changes.

3) Tag and push
```bash
git pull --rebase
git tag vX.Y.Z -m "Release vX.Y.Z"
git push origin vX.Y.Z
```

4) Release automation
- GitHub Actions will run tests and create a GitHub Release from the tag.
- Verify the release page contains assets and notes.

5) pkg.go.dev
- New tags are indexed automatically. If needed, request re-index on pkg.go.dev.

6) Aftercare
- Announce changes if needed.
- Create next milestone and issues.
