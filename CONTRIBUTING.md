# Contributing to hf-local-hub

Thank you for your interest in contributing to hf-local-hub!

## Development Setup

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/hf-local-hub.git
   cd hf-local-hub
   ```

3. Install Go 1.25+ and Python 3.11+

4. Install Python dependencies:
   ```bash
   cd python
   pip install -e ".[dev]"
   ```

5. Build the Go server:
   ```bash
   make server
   ```

## Running Tests

- Run all tests: `make test`
- Run Go tests: `make server-test`
- Run Python tests: `make python-test`

## Linting

- Run all linters: `make lint`
- Go lint: `make server-lint`
- Python lint: `make python-lint`

## Commit Conventions

We use [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation changes
- `style:` Code style changes (formatting)
- `refactor:` Code refactoring
- `test:` Test additions or changes
- `chore:` Build process or auxiliary tool changes

Examples:
- `feat: add support for model version tags`
- `fix: resolve issue with large file uploads`
- `docs: update API documentation`

## Pull Request Process

1. Create a branch from `main`:
   ```bash
   git checkout -b feat/your-feature-name
   ```

2. Make your changes and commit with conventional commits

3. Run tests and linting:
   ```bash
   make test lint
   ```

4. Push to your fork and create a pull request

5. Ensure all CI checks pass

## Code Style Guidelines

### Go
- Follow standard Go conventions (`gofmt`)
- Use `golangci-lint` with strict mode
- Keep functions under 50 lines
- Minimize dependencies

### Python
- Follow PEP 8
- Use `ruff` for linting
- Use `mypy` for type checking (strict mode)
- Maximum line length: 100 characters

## Project Structure

```
.
├── server/          # Go server implementation
│   ├── main.go     # Entry point
│   ├── api/        # API handlers
│   ├── storage/    # Storage layer
│   └── db/         # Database models
├── python/         # Python client package
│   ├── src/hf_local/
│   └── tests/
├── data/           # Local storage directory
└── docs/           # Documentation
```

## Questions?

Feel free to open an issue for questions or discussions!
