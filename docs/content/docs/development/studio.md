---
weight: 903
title: "Supabase Studio"
description: ""
icon: "article"
date: "2026-04-01T21:44:12+02:00"
lastmod: "2026-04-01T21:44:12+02:00"
draft: false
toc: false
---

During local development, using the Supabase Studio is a convenient way to interact with the database and also to validate the integration between the Supabase components.
When it comes to the optional Studio authentication, there are some things to consider for local testing.

## Auth

The `bazel run //dev:start_cluster` script does not only start a local `kind` cluster and a local `registry` instance but also a [Dex](https://dexidp.io/) instance to be able to test the OAuth2 authentication for the Studio.
The credentials for the `admin` user can be found in the `dev/dex_config.yaml` file.
For simplicity, both users and applications are statically configured for the scenario of local development.
You can find further details for the so called 'builtin connector' in [the docs](https://dexidp.io/docs/connectors/local/).

There's one limitation of this approach.
For the Studio container to be able to fetch tokens from Dex, the `issuer` is configured as `http://studio-idp:5556` which can be resolved from the Studio container but by default, can't be resolved on your local machine.
To be able to correctly log in, you can either:

1. Define a `/etc/hosts` entry to map `studio-idp` to `127.0.0.1`
2. Manually change the URL in the browser during the login process from `http://studio-idp:5556/...` to `http://localhost:5556/...` to complete the login
