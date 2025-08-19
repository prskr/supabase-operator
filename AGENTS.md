# AGENTS.md

## Build/Lint/Test Commands
- `make build` - Build the manager binary
- `make fmt` - Run go fmt against code
- `make vet` - Run go vet against code
- `make lint` - Run golangci-lint linter
- `make test` - Run unit tests
- `make test-e2e` - Run end-to-end tests
- `go test ./internal/controller/... -v` - Run a single test package
- `go list -u -f '{{if (and (not (or .Main .Indirect)) .Update)}}{{.Path}}: {{.Version}} -> {{.Update.Version}}{{end}}' -m all` - Check for dependency updates

## Code Style Guidelines
- Use `gofumpt` and `goimports` for formatting
- Follow Go naming conventions (PascalCase for public, camelCase for private)
- Use descriptive variable and function names
- Import paths should follow the module structure with local-prefixes
- All files must have proper copyright headers
- Use `go vet` and `golangci-lint` for static analysis
- Error handling should check and handle errors appropriately
- Prefer clear, readable code over clever optimizations
