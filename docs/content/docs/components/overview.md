---
weight: 301
title: "Overview"
description: ""
icon: "article"
date: "2026-04-01T21:10:48+02:00"
lastmod: "2026-04-01T21:10:48+02:00"
draft: false
toc: true
---

## Architecture

The following diagram explains the overall architecture of Supabase as it is deployed from this operator:

![](images/architecture.svg)

The dashed parts on the left are not managed by this operator but illustrate how traffic can be routed to the Supabase instance.

## Core

The `Core` custom resource (CR) is the basic resource that - if not in an 'edge case' (pun intended) - configures all basic aspects of a Supabase instance:

- PostgREST API
- Supabase Auth
- initial database migrations

## API Gateway

In the current `docker compose` stack the API Gateway is implemented with `Kong`.
Because `supabase-operator` tries to be as resource efficient as possible, `Kong` was replaced with `Envoy`.
The API Gateway is taking care of handling JWT tokens and routing traffic to the individual backend components.

## Dashboard

The `Dashboard` CR manages both the Supabase Studio and the `meta` service the Studio depends on.

## Storage

The `Storage` CR controls the Supabase Storage API.
It supports either local storage (e.g. based on a [`PVC`](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims)) or storing data on a S3 bucket.

## Realtime

The real time capability is currently not yet implemented.
