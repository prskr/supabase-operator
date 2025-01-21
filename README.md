# supabase-operator

This is a Kubernetes operator for managing Supabase instances.
It is built using the [Kubebuilder](https://book.kubebuilder.io/) framework.

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

## Getting Started

### Prerequisites
- go version v1.23.x+
- docker version 27.+.
- kubectl version v1.30.0+.
- Access to a Kubernetes v1.30.+ cluster.

### To Deploy on the cluster
**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=<some-registry>/supabase-operator:tag
```

**NOTE:** This image ought to be published in the personal registry you specified.
And it is required to have access to pull the image from the working environment.
Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=<some-registry>/supabase-operator:tag
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin
privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

>**NOTE**: Ensure that the samples has default values to test it out.

### To Uninstall
**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## Project Distribution

Following are the steps to build the installer and distribute this project to users.

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/supabase-operator:tag
```

NOTE: The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without
its dependencies.

2. Using the installer

Users can just run kubectl apply -f <URL for YAML BUNDLE> to install the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/supabase-operator/<tag or branch>/dist/install.yaml
```

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
