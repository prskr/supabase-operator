# -*- mode: Python -*-
load('ext://restart_process', 'docker_build_with_restart')

allow_k8s_contexts('kind-kind')

k8s_yaml(kustomize('config/dev'))

compile_cmd = 'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o out/supabase-operator ./cmd/'

local_resource(
  'manager-go-compile',
  compile_cmd,
  deps=[
    './api',
    './cmd',
    './assets',
    './infrastructure',
    './internal',
    './go.mod',
    './go.sum'
  ],
  resource_deps=[]
)

k8s_kind('Cluster', api_version='postgresql.cnpg.io/v1')

docker_build_with_restart(
  'controller',
  '.',
  entrypoint=['/app/bin/supabase-operator'],
  dockerfile='dev/Dockerfile',
  only=[
    './out',
  ],
  live_update=[
    sync('./out', '/app/bin'),
  ],
)

k8s_resource('supabase-controller-manager')
k8s_resource(
    workload='supabase-control-plane',
    port_forwards=18000,
    resource_deps=[]
)

k8s_resource(
    workload='cluster-example',
    new_name='Postgres cluster',
    port_forwards=5432
)

k8s_resource(workload='minio', port_forwards=[9000,9090])

k8s_resource(
    objects=["core-sample:Core:supabase-demo"],
    new_name='Supabase Core',
    resource_deps=[
        'Postgres cluster',
        'supabase-controller-manager'
    ],
)

k8s_resource(
    objects=["gateway-sample:APIGateway:supabase-demo"],
    extra_pod_selectors={"app.kubernetes.io/component": "api-gateway"},
    port_forwards=[3000, 8000, 19000],
    links=[
        link("https://localhost:3000", "Studio"),
        link("http://localhost:8000", "API"),
        link("http://localhost:19000", "Envoy Admin Interface")
    ],
    new_name='API Gateway',
    resource_deps=[
        'supabase-controller-manager'
    ],
)

k8s_resource(
    objects=["dashboard-sample:Dashboard:supabase-demo"],
    extra_pod_selectors={"app.kubernetes.io/component": "dashboard", "app.kubernetes.io/name": "studio"},
    discovery_strategy="selectors-only",
    new_name='Dashboard',
    resource_deps=[
        'supabase-controller-manager'
    ],
)

k8s_resource(
    objects=["storage-sample:Storage:supabase-demo"],
    new_name='Storage',
    resource_deps=[
        'supabase-controller-manager'
    ],
)
