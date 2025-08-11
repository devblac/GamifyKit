Contributing to GamifyKit

Thank you for your interest in contributing! Please:
- Open an issue to discuss significant changes before a PR.
- Run tests with race detector: `go test -race ./...`.
- Run linters: `golangci-lint run` and `gosec ./...` if available.
- Add tests for new features. Maintain 100% coverage in touched packages.

Development
- Go 1.22+
- Use `make test` (or `go test -race ./...`).

Code Style
- Small, composable packages. Prefer pure functions in `core`.
- Avoid global state. Favor dependency injection.
- Use context and deadlines for I/O.

License
By contributing, you agree your contributions are licensed under Apache-2.0.


