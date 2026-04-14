# Contributing to gswag

Thank you for considering a contribution to gswag.

## Getting started

1. Fork the repository and clone your fork.
2. Install Go 1.24 or later.
3. Run the full test suite to confirm a clean baseline:

```sh
make test
make examples
```

## Development workflow

### Testing

```sh
make test           # unit + golden tests
make test-verbose   # with -v
make test-race      # race detector
make cover          # coverage summary
make cover-html     # HTML coverage report
make examples       # build and run all example suites
```

### Golden files

Some tests compare generated OpenAPI specs against committed golden files under `test/*/`.
If your change intentionally alters spec output, regenerate them:

```sh
make update-golden
```

Commit the updated golden files alongside your code change.

### Linting

```sh
make vet
make fmt
make lint   # requires golangci-lint
```

## How to contribute

### Reporting bugs

Open a GitHub issue with:

- Go version (`go version`)
- gswag version or commit hash
- Minimal reproducer — ideally a failing test or a code snippet
- Expected vs actual behaviour

### Suggesting features

Open an issue describing the use case before starting implementation.
For large changes, discuss the approach first to avoid wasted effort.

### Submitting a pull request

1. Create a branch from `main`.
2. Keep changes focused — one logical change per PR.
3. Add or update tests for every behavioural change.
4. Update golden files if spec output changes (`make update-golden`).
5. Ensure `make test && make examples` pass locally.
6. Write a clear PR description explaining *why*, not just *what*.

### Commit style

Use conventional commits:

```
feat: short description
fix: short description
chore: short description
docs: short description
test: short description
refactor: short description
```

## Project layout

| Path | Purpose |
|---|---|
| `*.go` (root) | Core DSL, spec collection, output |
| `internal/schemautil` | JSON schema inference from response bytes |
| `internal/golden` | Golden-file test helpers |
| `test/*/` | Golden integration test suites |
| `examples/*/` | Full example apps (gin, echo, chi, fiber, stdlib, gorilla, todo) |

See [ARCHITECTURE.md](ARCHITECTURE.md) for a deeper walkthrough of the three-phase model and internal types.

## License

By contributing you agree that your contributions will be licensed under the [MIT License](LICENSE).
