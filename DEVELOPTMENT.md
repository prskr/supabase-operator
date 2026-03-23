# Development

## Prerequisites

- Bazel ( latest stable) / Bazelisk (will pick up version automatically)
- *optional:* [`ibazel`](https://github.com/bazelbuild/bazel-watcher) to watch for changes
- Docker

All other tools are managed via Bazel either as `go tool` or with the help of [`rules_multitool`](https://github.com/bazel-contrib/rules_multitool).

Although not strictly necessary, it is a lot more convenient if you also have [`direnv`](https://direnv.net/) set up.
When set up correctly, it is enough to run `bazel run //tools:bazel_env` to make all the necessary tools available in your `$PATH`.

## Language server

You can use a local version of `gopls` if you prefer, but to be sure that the `gopls` version is compatible with the Go SDK used by this project, it's recommended to use the Bazel managed version of `gopls` as well.
To determine the path to the local binary, you can use the following command:

```shell
bazel cquery --output files @org_golang_x_tools_gopls//:gopls 2>/dev/null

# Example output
# bazel-out/darwin_arm64-fastbuild/bin/external/gazelle++go_deps+org_golang_x_tools_gopls/gopls_/gopls
```

Alternatively you can also customize the binary for the language server to `bazel run @org_golang_x_tools_gopls//:gopls` to let Bazel take care of the path.

## Local Dev Cluster

To get started with local development, you can run `bazel run //dev:start_cluster` to ramp up a local `kind` based K8s cluster.
All necessary binaries (including `kind`) are managed by Bazel.
The only local dependency is a running Docker daemon (or compatible alternative) and the `docker` CLI in your `$PATH`.

To remove the local dev cluster, run `bazel run //dev:teardown_cluster`.

## Starting the operator

To start the dev loop, it is recommended to use `ibazel run //dev:dev`.
Without `ibazel` you would have to trigger a build whenever you changed some files.
No matter, what you prefer, the `//dev:dev` target takes care of the following things:

- Build an OCI image
- Push the image to the local container image registry
- update the `config/dev/kustomization.yaml` to reflect the latest image build
- apply the rendered Kubernetes manifests to the local dev cluster (including custom resources to bootstrap a Supabase instance)

When starting development, you might need to run the command twice to ensure that the CRDs are installed to the cluster before the custom resources are created.

## Code Genration

### Go code and CRD manifests

To update the generated Go code and the CRD manifests, run:

```shell
bazel run --run_in_cwd //dev:generate_crd_code
```

**Note:** the `--run_in_cwd` flag is mandatory to ensure that the generated code is actually written to the working tree, make sure to run the command in the repo-root

This will invoke the `controller-gen` CLI to generate all depending code

### CRD documentation

To update the generated CRD documentation, run

```shell
bazel run //dev:generate_crd_docs
```
