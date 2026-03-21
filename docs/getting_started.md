# Getting Started

## Dependencies

The operator has the following dependencies:

- [cert-manager](https://cert-manager.io/) (to issue webhook certificates)

## Deploying the operator

The easiest way to deploy the operator is to fetch the manifest from the [releases](https://github.com/prskr/supabase-operator/releases) and apply it like this:

```bash
kubectl apply --server-side -f manifest.yaml
```

The manifest is rendered as part of the release workflow and based on [kustomize](https://kustomize.io/).
If you want to customize the deployment, you can start from the [release/default](https://github.com/prskr/supabase-operator/src/branch/main/config/release/default) layer and build your own manifest.

## Deploying a basic Supabase instance

As described in the [overview](./components/overview.md) the custom resources are 'grouping' the Supabase services into 'modules'.
A common basic instance requires:

- a [`Core`](./components/core.md) instance (PostgREST, Auth & DB migrations)
- an [`APIGateway`](./components/apigateway.md) instance (gateway to handle JWT auth and routing)

it is perfectly possible to deploy for instance only an `APIGateway` and a `Storage` instance as well if you don't need the API or you can also manage your own API gateway if you prefer it that way.
The operator setup tries to be as 'unopinionated' as possible.
