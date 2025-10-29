#!/bin/bash

kubectl config use-context kind-supabase-operator-debug

kubectl port-forward -n supabase-demo service/cluster-example-rw 5432:5432 &

kubectl_pid=$!

supabase db push \
    --include-seed \
    --db-url "postgresql://supabase_admin:1n1t-R00t!@localhost:5432/app"

kill -SIGTERM $kubectl_pid

wait "$PID" 2>/dev/null && echo "kubectl port-forward terminated gracefully." || echo "kubectl port-forward may still be running."
