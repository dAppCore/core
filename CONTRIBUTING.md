# Contributing

Thank you for your interest in contributing!

## Requirements
- **Go Version**: 1.26 or higher is required.
- **Tools**: `golangci-lint` and `task` (Taskfile.dev) are recommended.

## Development Workflow
1. **Testing**: Ensure all tests pass before submitting changes.
   ```bash
   go test ./...
   ```
2. **Code Style**: All code must follow standard Go formatting.
   ```bash
   gofmt -w .
   go vet ./...
   ```
3. **Linting**: We use `golangci-lint` to maintain code quality.
   ```bash
   golangci-lint run ./...
   ```

## Commit Message Format
We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:
- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation changes
- `refactor`: A code change that neither fixes a bug nor adds a feature
- `chore`: Changes to the build process or auxiliary tools and libraries

Example: `feat: add new endpoint for health check`

## License
By contributing to this project, you agree that your contributions will be licensed under the **European Union Public Licence (EUPL-1.2)**.
