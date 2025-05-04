# Contributing to GoVisual

Thank you for your interest in contributing to GoVisual! This document provides guidelines and instructions for contributing to the project.

## Development Setup

### Prerequisites

- Go 1.20 or higher
- Docker and Docker Compose (for running the examples with databases)
- Git

### Getting Started

1. Fork the repository on GitHub
2. Clone your fork locally
   ```bash
   git clone https://github.com/yourusername/govisual.git
   cd govisual
   ```
3. Add the original repository as an upstream remote
   ```bash
   git remote add upstream https://github.com/doganarif/govisual.git
   ```
4. Install dependencies
   ```bash
   go mod download
   ```

## Running Tests

Run the tests with:

```bash
go test ./...
```

For tests involving storage backends, you can use the provided Docker Compose files:

```bash
# For PostgreSQL tests
cd cmd/examples/multistorage
GOVISUAL_STORAGE_TYPE=postgres \
GOVISUAL_PG_CONN="postgres://postgres:postgres@localhost:5432/govisual?sslmode=disable" \
go test ../../internal/store/...

# For Redis tests
GOVISUAL_STORAGE_TYPE=redis \
GOVISUAL_REDIS_CONN="redis://localhost:6379/0" \
go test ../../internal/store/...
```

## Code Style Guidelines

GoVisual follows standard Go coding conventions:

- Run `go fmt` before committing to ensure consistent formatting
- Follow [Effective Go](https://golang.org/doc/effective_go) guidelines
- Use `golint` and `go vet` to check for common issues
- Write meaningful comments, especially for exported functions and types
- Keep functions small and focused on a single responsibility
- Use meaningful variable and function names that describe their purpose

## Contribution Workflow

1. Create a new branch for your feature or bugfix

   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes, following the code style guidelines

3. Add tests for your changes

4. Run tests to make sure everything works

   ```bash
   go test ./...
   ```

5. Commit your changes with a clear and descriptive commit message

   ```bash
   git commit -m "Add support for new feature X"
   ```

6. Push to your fork

   ```bash
   git push origin feature/your-feature-name
   ```

7. Create a Pull Request against the main repository

## Pull Request Guidelines

- Provide a clear description of the problem you're solving
- Update documentation if necessary
- Add or update tests as appropriate
- Keep PRs focused on a single issue/feature to make them easier to review
- Make sure CI tests pass

## Adding Storage Backends

When adding a new storage backend:

1. Implement the `Store` interface in `internal/store/store.go`
2. Add relevant configuration options in `options.go`
3. Update factory methods in `internal/store/factory.go`
4. Add documentation in `docs/storage-backends.md`
5. Create examples showing usage

## Reporting Issues

When reporting issues, please include:

- A clear description of the problem
- Steps to reproduce
- Expected vs. actual behavior
- Version of GoVisual you're using
- Go version and OS
- Any relevant logs or error messages

## License

By contributing to GoVisual, you agree that your contributions will be licensed under the project's MIT license.
