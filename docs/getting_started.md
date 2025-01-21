# Getting Started

## Deploying the operator

The easiest way to deploy the operator is to fetch the manifest from the [releases](https://code.icb4dc0.de/prskr/supabase-operator/releases) and apply it like this:

```bash
kubectl apply --server-side -f manifest.yaml
```

The manifest is rendered as part of the release workflow and based on [kustomize](https://kustomize.io/).
If you want to customize the deployment, you can start from the [release/default](https://code.icb4dc0.de/prskr/supabase-operator/src/branch/main/config/release/default) layer and build your own manifest.

## Deploying a basic Supabase instance
