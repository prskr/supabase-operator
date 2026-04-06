---
weight: 2
title: "GitHub"
description: ""
icon: "article"
date: "2026-04-01T21:10:48+02:00"
lastmod: "2026-04-01T21:10:48+02:00"
draft: false
toc: true
---

Please make sure to check the general Supabase [GitHub social login docs](https://supabase.com/docs/guides/auth/social-login/auth-github) to get an overview on how to prepare a GitHub OAuth2 app.

## Example

The following example illustrates how you can configure the `Core` resource to enable GitHub social login support:

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
      github:
        enabled: true
        clientID: <your-client-id>
        clientSecretRef:
          name: <name-of-your-secret>
          key: <key-of-client-secret>
  # ...
```

## Reference

[See](../../api/supabase.k8s.icb4dc0.de.md#githubauthprovider)
