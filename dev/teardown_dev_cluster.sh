#!/usr/bin/env bash

set -o errexit

# --- Tool paths passed in from Bazel ---
KIND="${1:?Usage: $0 <kind_binary> <kubectl_binary>}"

reg_name='kind-registry'

"${KIND}" delete cluster --name supabase-operator-debug

if [ "$(docker inspect -f '{{.State.Running}}' "${reg_name}" 2>/dev/null || true)" == 'true' ]; then
  docker stop "${reg_name}"
  docker rm -f "${reg_name}"
fi
