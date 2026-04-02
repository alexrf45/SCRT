---
paths:
  - "**/Makefile"
  - "**/go.mod"
  - "**/*.go"
  - "**/README.md"
---
# SCRT Deployment Rules

## Build Method

ALWAYS use `go build` for deployment. NEVER use `go install` or suggest publishing the package.

```
make build       # produces bin/scrt
make all         # runs vet + test + build
```

## Documentation

- Do not include `go install` instructions in README.md or any docs
- Do not include instructions for publishing to pkg.go.dev or any package registry
- README.md must reflect the current status of the project and be kept up to date as features change
