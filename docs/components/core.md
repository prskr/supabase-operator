# Core

The `Core` resource configures the essential Supabase services:

- PostgREST
- Auth

and it manages the initial DB migrations that are not done by the services themselves and are part of the [supabase/postgres](https://github.com/supabase/postgres) repository.

To deploy a basic `Core` instance, you can use the following snippet as a 'template':

```yaml title="Basic 'Core' example" linenums="1"
--8<-- "config/samples/supabase_v1alpha1_core.yaml"
```

## Postgres Database

This operator **only** manages the Supabase **specific** migrations and roles.
There are plenty of Postgres operators like:

- [cloudnative-pg/cloudnative-pg](https://github.com/cloudnative-pg/cloudnative-pg)
- [CrunchyData/postgres-operator](https://github.com/CrunchyData/postgres-operator)
- [zalando/postgres-operator](https://github.com/zalando/postgres-operator)

(Ordered alphabetically)

that handle all aspects of managing a Postgres instance or even a whole cluster including various backup and restore capabilities and much more.

This operator "only" needs a `superuser` access to the Postgres instance, where you would like to setup your Supabase instance.

The connection to the database is configured via a regular DSN in a Kubernetes `Secret` like this:

```yaml
---
apiVersion: v1
kind: Secret
metadata:
  name: supabase-demo-credentials
  namespace: supabase-demo
stringData:
  url: postgresql://supabase_admin:<password>@<postgres-host>:5432/<postgres DB>
```

Please note that the structure of Supabase migrations makes it hard to run multiple Supabase instances (or generally multiple apps, for that matter) on the same Postgres instance.
Independently of whether this might possible in the future or not, it does make sense to isolate Supabase apps against each other from a resource perspective, which is easier if you do not share database servers.

## PostgREST

The default configuration for PostgREST should be sufficient for most use cases.
If you want to customize the PostgREST, please check the [API reference](../api/supabase.k8s.icb4dc0.de.md#postgrestspec).

## Auth

[Supabase Auth](https://supabase.com/docs/guides/auth) supports a vast amount of auth providers.
This operator will eventually support the same set of auth providers, but, considering that it is currently still more of a PoC, it is currently still limited.
To get a detailed view of which providers are currently supported, please check the [auth overview](../auth/overview.md) and the [API reference](../api/supabase.k8s.icb4dc0.de.md#authproviders).

Auth providers are configured via environment variables exposed to the `auth` service. The configuration of each provider abstracts these environment variables, in order to provide you with a typed API, to avoid as many mistakes as possible.
If you are in desperate need of a provider that is not yet supported, please create a [GitHub issue](https://github.com/prskr/supabase-operator/issues/new).
