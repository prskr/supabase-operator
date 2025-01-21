# About

This is a Kubernetes operator for managing self-hosted [Supabase](https://supabase.com/) instances.
It is built using the [Kubebuilder](https://book.kubebuilder.io/) framework.
This project is not affiliated with the Supabase project or company in any way.

## Description

This is currently a work-in-progress experiment to replace existing Helm charts for Supabase as they tend to be hard to deploy and to manage and the default Supabase stack - although working great as a single instance or in their SaaS instances - isn't a perfect fit for Kubernetes.
This operator replaces tedious Helm values files with a small set of custom resources that allow an user to quickly deploy a Supabase instance without having to know much (if anything) of the Supabase internals.

## Targets

- Make it as easy as possible to deploy Supabase on a Kubernetes cluster
- Manage updates of components
- Run Supabase specific migrations on the database (those managed in the supabase/postgres repository)
- Make Supabase as resource effective as possible (e.g. replaced Kong with Envoy)
- Keep it as secure as possible (e.g. adding OAuth2/Basic auth to studio if desired)
- Manage **all** aspects of the Kubernetes resources, this operator manages everything where a user would need deeper insights into Supabase like:
  - Deployments
  - Services
  - Secrets (although you can ship your own if you want to)
  - ConfigMaps
  - *soon*: NetworkPolicies

## Non-Targets

- Manage **all** Kubernetes aspects, it does **not** create:
  - PodDisruptionBudgets
  - HorizontalPodAutoscalers
  - Ingress or HTTPRoutes
- Replace existing Postgres operators like cloudnative-pg, CrunchyData, Zalando Postgres Operator, ...
- Manage the database instance e.g. making backups, ... that should be done by the Postgres operator or by the user
- Manage your application e.g. run app specific migrations, host your frontend, ...

This operator tries to be as un-opionionated as possible and thereofore does not make assumptions on how you expose APIs to your users (Ingress, GatewayAPI, LoadBalancer service (coming soon))...
