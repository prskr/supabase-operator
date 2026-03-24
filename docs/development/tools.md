# Tools

## Essentials

- Bazel ( latest stable) / Bazelisk (will pick up version automatically)
- - *optional:* [`ibazel`](https://github.com/bazelbuild/bazel-watcher) to watch for changes
- **optionally:** current Go version (see `go.mod`)
- Any Docker daemon[^1]
- kubectl (v1.30.0+.)
- Access to a Kubernetes v1.30.+ cluster[^2]

All other tools are managed via Bazel either as `go tool` or with the help of [`rules_multitool`](https://github.com/bazel-contrib/rules_multitool).

[^1]: Docker Desktop, Orbstack, Podman, ...
[^2]: kind, Orbstack, Docker Desktop, Rancher Desktop, k3s, microk8s, ...

## Recommended

Although not strictly necessary, it is a lot more convenient if you also have [`direnv`](https://direnv.net/) set up.
When set up correctly, it is enough to run `bazel run //tools:bazel_env` to make all the necessary tools available in your `$PATH`.

- [uv](https://docs.astral.sh/uv/) to manage a virtual environment to work on the docs

## Language server

You can use a local version of `gopls` if you prefer, but to be sure that the `gopls` version is compatible with the Go SDK used by this project, it's recommended to use the Bazel managed version of `gopls` as well.
To determine the path to the local binary, you can use the following command:

```shell
bazel cquery --output files @org_golang_x_tools_gopls//:gopls 2>/dev/null

# Example output
# bazel-out/darwin_arm64-fastbuild/bin/external/gazelle++go_deps+org_golang_x_tools_gopls/gopls_/gopls
```

Alternatively you can also customize the binary for the language server to `bazel run --run_in_cwd @org_golang_x_tools_gopls//:gopls` to let Bazel take care of the path.
