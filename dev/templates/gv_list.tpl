{{- define "gvList" -}}
{{- $groupVersions := . -}}
---
weight: 400
title: "API Reference"
description: ""
icon: "article"
date: "2026-04-01T21:10:48+02:00"
lastmod: "2026-04-01T21:10:48+02:00"
draft: false
toc: true
---

## Packages
{{- range $groupVersions }}
- {{ markdownRenderGVLink . }}
{{- end }}

{{ range $groupVersions }}
{{ template "gvDetails" . }}
{{ end }}

{{- end -}}
