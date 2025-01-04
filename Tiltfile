# -*- mode: Python -*-
load('ext://restart_process', 'docker_build_with_restart')

allow_k8s_contexts('kind-kind')

local('./dev/prepare-dev-cluster.sh')
k8s_yaml(kustomize('config/dev'))
k8s_yaml(kustomize('config/samples'))

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

docker_build_with_restart(
  'supabase-operator',
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
)

k8s_resource(
    objects=["cluster-example:Cluster:supabase-demo"],
    new_name='Postgres cluster',
    port_forwards=5432
)

k8s_resource(
    objects=["core-sample:Core:supabase-demo"],
    new_name='Supabase Core',
    resource_deps=[
        'Postgres cluster',
        'supabase-controller-manager'
    ],
)

k8s_resource(
    objects=["core-sample:APIGateway:supabase-demo"],
    extra_pod_selectors={"app.kubernetes.io/component": "api-gateway"},
    port_forwards=[8000, 19000],
    new_name='API Gateway',
    resource_deps=[
        'supabase-controller-manager'
    ],
)
