---
weight: 902
title: "Local Development"
description: ""
icon: "article"
date: "2026-04-01T21:44:12+02:00"
lastmod: "2026-04-01T21:44:12+02:00"
draft: false
toc: false
---

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

This will invoke the `controller-gen` CLI to generate all dependent code and also update the CRD documentation
