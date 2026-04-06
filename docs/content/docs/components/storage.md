---
weight: 304
title: "Storage"
description: ""
icon: "article"
date: "2026-04-01T21:10:48+02:00"
lastmod: "2026-04-01T21:10:48+02:00"
draft: false
toc: true
---

The `Storage` resources configures the following optional services:

- [Supabase Storage](https://supabase.com/docs/guides/storage) (S3 compatible)
- [Image proxy](https://supabase.com/docs/guides/storage/serving/image-transformations)

The image proxy can be omitted if you don't need it.

The Supabase Storage service can use two different storage backends:

- an upstream S3 storage
- a local volume (e.g. a `PVC`)

If you want to use another object storage (like Azure Storage Accounts), it is advised to check whether there's a Container Storage Interface (CSI) driver available that allows you to use the object storage in question as a persistent volume.

## Upstream S3 storage

The following example illustrates how to connect your `Storage` API to an upstream S3 storage:

{{% code file="/static/samples/supabase_v1alpha1_storage.yaml" language="yaml" %}}

Please note that the credentials are referenced via a Kubernetes `Secret`.
The keys of the secret **can** be configured, but by default the `Secret` would look like this:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: storage-s3-credentials
stringData:
  accessKeyId: <value>
  secretAccessKey: <value>
```

if you want or need different secret keys, please have a look at the [API reference](../api/supabase.k8s.icb4dc0.de.md#s3credentialsref).

## Local volume

Alternatively, you can use any 'local' storage in the Pod.
It is strongly recommended to create and mount a `PVC` or a host path to ensure persistence, but it's strictly necessary.

The following example shows how you can configure the `Storage` API for local storage and how you can customize the workload to mount a volume.

 {{% code file="/static/samples/supabase_v1alpha1_storage_file_backend.yaml" language="yaml" %}}

Please note that, the API workload will be a Kubernetes `Deployment`.
But even if this would change in the future to a `StatefulSet` for some reason, the `Storage` API is really **only** an API and does not replicate or distribute data across instances.
In consequence, when using Kubernetes volumes, you should either use volumes that ideally support `ReadWriteMany` mode or you might want to configure the `strategy` of the `Storage` API workload to `Recreate` (see also [upstream docs](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#recreate-deployment)).

For further details on how to configure the `strategy` please check out the [reference docs](../api/supabase.k8s.icb4dc0.de.md#storageworkloadspec).
