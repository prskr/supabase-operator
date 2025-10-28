# AGENTS.md

Use `bazel` to build the project instead of using the go CLI directly

## Build/Lint/Test Commands
- `bazel build //:supabase-operator` - Build the manager binary
- `gofumpt -l -w .` - Format all go code
- `bazel test //...` - Run unit tests
- `make test-e2e` - Run end-to-end tests
- `bazel test //internal/controller/...` - Run a single test package
- `go list -u -f '{{if (and (not (or .Main .Indirect)) .Update)}}{{.Path}}: {{.Version}} -> {{.Update.Version}}{{end}}' -m all` - Check for dependency updates

## Code Style Guidelines
- Use `gofumpt` and `goimports` for formatting
- Follow Go naming conventions (PascalCase for public, camelCase for private)
- Use descriptive variable and function names
- Import paths should follow the module structure with local-prefixes
- All files must have proper copyright headers
- Error handling should check and handle errors appropriately
- Prefer clear, readable code over clever optimizations
