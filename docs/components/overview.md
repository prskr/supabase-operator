# Overview

## Architecture

The following diagram explains the overall architecture of Supabase as it is deployed from this operator:

```kroki-excalidraw
@from_file:./images/overview.excalidraw
```

The dashed parts on the left are not managed by this operator but illustrate how traffic can be routed to the Supabase instance.

## Core

The `Core` custom resource (CR) is the basic resource that - if not in an 'edge case' (pun intended) - configures all basic aspects of a Supabase instance:

- PostgREST API
- Supabase Auth
- initial database migrations
