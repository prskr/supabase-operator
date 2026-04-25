# AGENTS.md

Use `bazel` to build the project instead of using the go CLI directly.

## Build/Lint/Test Commands
- `bazel build //:supabase-operator` - Build the manager binary.
- `bazel run //dev:generate_crd_code` - Regenerate CRD manifests, DeepCopy methods, and documentation. **Required after modifying `_types.go` files.**
- `bazel run //dev:update_snapshots` - Update snapshots for controller tests.
- `bazel run //:format` - Format all code.
- `bazel test //...` - Run unit tests.
- `bazel test //internal/controller/... --test_env "KUBEBUILDER_ASSETS=$(bazel run --run_in_cwd @io_k8s_sigs_controller_runtime_tools_setup_envtest//:setup-envtest -- use -p path)"` - Run controller tests with `envtest` binaries.
- `make test-e2e` - Run end-to-end tests.
- `go list -u -f '{{if (and (not (or .Main .Indirect)) .Update)}}{{.Path}}: {{.Version}} -> {{.Update.Version}}{{end}}' -m all` - Check for dependency updates.

## Technical Learnings & Patterns

### Kubernetes API Design
- **Optional vs. Required**:
    - Fields that are **required** or should have **reliable defaults** should be declared as **values** (non-pointers).
    - Fields that are truly **optional** (where `nil` has semantic meaning) should be **pointers**.
- **Defaulting**: When moving from pointers to values, ensure `zz_generated.deepcopy.go` is regenerated and that defaulter webhooks/kubebuilder tags are updated to check for zero-values (e.g., empty strings) instead of `nil`.

### Testing Reconcilers
- **Snapshot Testing**: Use `github.com/gkampitakis/go-snaps` to verify the state of Kubernetes objects created or updated by reconcilers.
- **Masking Dynamic Fields**: Always use `match.Any` to mask auto-generated fields in snapshots to ensure they are stable. Common fields to mask include:
    - `metadata.uid`
    - `metadata.resourceVersion`
    - `metadata.creationTimestamp`
    - `metadata.managedFields`
    - `metadata.ownerReferences.0.uid`
    - `data` fields containing secrets, certificates, or random tokens (e.g., `oauth2_hmac_secret`, `tls.key`).
- **Updating Snapshots**: When intentional changes to the reconciler logic occur, update the snapshots using `bazel run //dev:update_snapshots`. This command uses the Bazel Go toolchain and `envtest` binaries to ensure consistency.

## Code Style Guidelines
- Use `bazel run //:format` for formatting.
- Follow Go naming conventions (PascalCase for public, camelCase for private).
- All files must have proper copyright headers.
- Error handling should check and handle errors appropriately.
- Prefer clear, readable code over clever optimizations.

## Committing Changes
- **Conventional Commits**: Use the [Conventional Commits](https://www.conventionalcommits.org/) schema (e.g., `feat(storage): ...`, `fix(api): ...`, `refactor: ...`).
- **Pre-commit Hooks**: The project uses pre-commit hooks (e.g., end-of-file-fixer, yaml-check). If a hook modifies a file, the commit will fail; you must re-stage the modified files and attempt the commit again.
- **Commit Messages**: Focus on the "why" and "how" of the change. Keep the first line under 72 characters.
