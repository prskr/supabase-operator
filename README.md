# supabase-operator

This is a Kubernetes operator for managing Supabase instances.
It is built using the [Kubebuilder](https://book.kubebuilder.io/) framework.

## Description

This is currently a work-in-progress experiment to replace existing Helm charts for Supabase as they tend to be hard to deploy and to manage and the default Supabase stack - although working great as a single instance or in their SaaS instances - isn't a perfect fit for Kubernetes.
This operator replaces tedious Helm values files with a small set of custom resources that allow a user to quickly deploy a Supabase instance without having to know much (if anything) of the Supabase internals.

## Goals

- Make it as easy as possible to deploy Supabase on a Kubernetes cluster
- Manage updates of components
- Run Supabase specific migrations on the database (those managed in the [supabase/postgres](https://github.com/supabase/postgres) repository)
- Make Supabase as resource effective as possible (e.g. replaced Kong with Envoy)
- Keep it as secure as possible (e.g. adding OAuth2/Basic auth to studio if desired)
- Manage **all deployment** aspects of the Kubernetes resources, this operator manages everything where a user would need deeper insights into Supabase like:
  - Deployments
  - Services
  - Secrets (although you can ship your own if you want to)
  - ConfigMaps
  - **optionally:** Observability
  - *soon*: NetworkPolicies

## Non-Goals

- Manage **all** Kubernetes aspects, it does **not** create:
  - PodDisruptionBudgets
  - HorizontalPodAutoscalers
  - Ingress or HTTPRoutes
- Replace existing Postgres operators like cloudnative-pg, CrunchyData, Zalando Postgres Operator, ...
- Manage the database instance e.g. making backups, ... that should be done by the Postgres operator or by the user
- Manage your application e.g. run app specific migrations, host your frontend, ...

This operator tries to be as unopionionated as possible and therefor does not make assumptions on how you expose APIs to your users (Ingress, GatewayAPI, LoadBalancer service (coming soon))...

## Limitations

### Edge Functions

This operator does **not** interact with Supabase's edge function feature.
The reason for that is that, even though the edge runtime works in the `docker compose` setup by copying your built function into a Docker volume, this does not scale very well in Kubernetes.
Furthermore, the designated way to deploy edge functions in a self-hosted way is, to build a container image and deploy this image [^1].
As previously mentioned, this operator does not manage the application routing beyond providing you with a Kubernetes `Service` for the Supabase API.

[^1]: See [docs](https://supabase.com/docs/reference/self-hosting-functions/introduction)

Also, there are many Kubernetes native ways to deploy serverless workloads:

- [Fermyon / SpinKube](https://www.spinkube.dev/)
- [Fission](https://fission.io/)
- [Knative](https://knative.dev/)
- [OpenFaaS](https://www.openfaas.com/)
- [OpenWhisk](https://openwhisk.apache.org/)
- [wasmCloud](https://wasmcloud.com/)

to mention just a few options.
Since the Supabase Edge Runtime does not stand out from others in terms of Kubernetes, it was decided not to include it for now.


## Getting Started

[See docs](https://prskr.github.io/supabase-operator/getting_started/)

## License

Copyright 2025 Peter Kurfer.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
