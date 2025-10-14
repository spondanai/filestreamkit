# Publish guide

Follow these steps to publish this module as a new GitHub repository and make it available via `go get`.

1) Create the GitHub repo `spondanai/FileStreamKit` (or fork under your namespace).

2) Initialize git in this folder and set origin:

```powershell
# from streamkit directory
git init
git add .
git commit -m "chore: initial import of streamkit"
# replace with your repo URL
git remote add origin https://github.com/spondanai/FileStreamKit.git
git branch -M main
git push -u origin main
```

3) Ensure `go.mod` module path matches the repo:

- `module github.com/spondanai/FileStreamKit`

4) Tag the first release so pkg.go.dev can index docs:

```powershell
git tag v0.1.0
git push origin v0.1.0
```

5) Verify docs on pkg.go.dev:

- https://pkg.go.dev/github.com/spondanai/FileStreamKit

Optional:

- Protect branches, enable Actions, and watch CI passing on PRs.
- Add release notes when creating tags.

Public visibility:

- Ensure the repository is set to Public in GitHub repo settings
- Push a tag `v0.1.0` or later so pkg.go.dev indexes docs
- Add badges to README (build status, pkg.go.dev):

```
[![Go Reference](https://pkg.go.dev/badge/github.com/spondanai/FileStreamKit.svg)](https://pkg.go.dev/github.com/spondanai/FileStreamKit)
[![Go CI](https://github.com/spondanai/FileStreamKit/actions/workflows/ci.yml/badge.svg)](https://github.com/spondanai/FileStreamKit/actions/workflows/ci.yml)
```