---
weight: 1
title: "Azure"
description: ""
icon: "article"
date: "2026-04-01T21:10:48+02:00"
lastmod: "2026-04-01T21:10:48+02:00"
draft: false
toc: true
---

Please make sure to check the general Supabase [Azure social login docs](https://supabase.com/docs/guides/auth/social-login/auth-azure) to get an overview on how to prepare an Azure "App Registration".

## Example

The following example illustrates how you can configure the `Core` resource to enable Azure social login support:

```yaml
apiVersion: supabase.k8s.icb4dc0.de/v1alpha1
kind: Core
metadata:
  # ...
spec:
  # ...
  auth:
    # ...
    providers:
      azure:
        enabled: true
        url: https://login.microsoftonline.com/<your-tenant-id>/v2.0"
        clientID: <your-client-id>
        clientSecretRef:
          name: <name-of-your-secret>
          key: <key-of-client-secret>
  # ...
```

## Reference

[See](../../api/supabase.k8s.icb4dc0.de.md#azureauthprovider)
