# APIGateway

The `APIGateway` custom resource deploys one or multiple [`Envoy`](https://www.envoyproxy.io/) instances.
They operate as a cluster and are configured by a central 'control plane' that is part of the operator.

The central 'control plane' subscribes to events fired by the Kubernetes API server e.g. when a new PostgREST pod is ready, updates the individual cluster configurations and pushes the configuration actively to all cluster nodes.

Compared to the original `Kong` implementation in the `docker compose` setup, `Envoy` is relatively lightweight to run (~10MB memory footprint) and scales to high request volumes effortlessly.

## Why do I even need an API Gateway?

You might ask

> Kubernetes has already `Ingress` and `GatewayAPI` that are handling virtually the same use case, why do I need an additional layer?!

You ain't wrong!
The main reason for shipping an additional API Gateway is, that not all Supabase backend services are not handling the JWT authentication themselves.
Authentation (AuthN) and Authorization (AuthZ) **can** be implemented with `Ingress`, but the implementation heavily relies on the `Ingress Controller` and even though there's [GEP-149](https://gateway-api.sigs.k8s.io/geps/gep-1494/) to standardize AuthN and AuthZ in the Gateway API, the feature is experimental and not yet implemented by many providers.
Considering that one of the main goals of this project is to remain **independent** of `Ingress` controllers and `GatewayAPI` providers, relying on provider specific capabilities, wasn't an option.
Also, this project predates GEP-149, so it wasn't and at least for now isn't an option, either.
Additionally, being independent of `Ingress` / `GatewayAPI` is also beneficial for the modularity

To avoid even more network hubs the `Envoy` API Gateway skips the `Service` layer in Kubernetes entirely and directly communicates with the `Pod`s
