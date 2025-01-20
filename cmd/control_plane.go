/*
Copyright 2025 Peter Kurfer.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	clusterservice "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	discoverygrpc "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	endpointservice "github.com/envoyproxy/go-control-plane/envoy/service/endpoint/v3"
	listenerservice "github.com/envoyproxy/go-control-plane/envoy/service/listener/v3"
	routeservice "github.com/envoyproxy/go-control-plane/envoy/service/route/v3"
	runtimeservice "github.com/envoyproxy/go-control-plane/envoy/service/runtime/v3"
	secretservice "github.com/envoyproxy/go-control-plane/envoy/service/secret/v3"
	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"google.golang.org/grpc"
	grpchealth "google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	ctrl "sigs.k8s.io/controller-runtime"
	mgr "sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/metrics/filters"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"code.icb4dc0.de/prskr/supabase-operator/internal/controlplane"
)

//nolint:lll // flag declaration with struct tags is as long as it is
type controlPlane struct {
	ListenAddr           string `name:"listen-address" default:":18000" help:"The address the control plane binds to."`
	MetricsAddr          string `name:"metrics-bind-address" default:"0" help:"The address the metrics endpoint binds to. Use :8443 for HTTPS or :8080 for HTTP, or leave as 0 to disable the metrics service."`
	EnableLeaderElection bool   `name:"leader-elect" default:"false" help:"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager."`
	ProbeAddr            string `name:"health-probe-bind-address" default:":8081" help:"The address the probe endpoint binds to."`
	SecureMetrics        bool   `name:"metrics-secure" default:"true" help:"If set, the metrics endpoint is served securely via HTTPS. Use --metrics-secure=false to use HTTP instead."`
	EnableHTTP2          bool   `name:"enable-http2" default:"false" help:"If set, HTTP/2 will be enabled for the metrics and webhook servers"`
}

func (cp controlPlane) Run(ctx context.Context) error {
	var tlsOpts []func(*tls.Config)

	// if the enable-http2 flag is false (the default), http/2 should be disabled
	// due to its vulnerabilities. More specifically, disabling http/2 will
	// prevent from being vulnerable to the HTTP/2 Stream Cancellation and
	// Rapid Reset CVEs. For more information see:
	// - https://github.com/advisories/GHSA-qppj-fm5r-hxr3
	// - https://github.com/advisories/GHSA-4374-p667-p6c8
	disableHTTP2 := func(c *tls.Config) {
		setupLog.Info("disabling http/2")
		c.NextProtos = []string{"http/1.1"}
	}

	if !cp.EnableHTTP2 {
		tlsOpts = append(tlsOpts, disableHTTP2)
	}

	// Metrics endpoint is enabled in 'config/default/kustomization.yaml'. The Metrics options configure the server.
	// More info:
	// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.1/pkg/metrics/server
	// - https://book.kubebuilder.io/reference/metrics.html
	metricsServerOptions := metricsserver.Options{
		BindAddress:   cp.MetricsAddr,
		SecureServing: cp.SecureMetrics,
		TLSOpts:       tlsOpts,
	}

	if cp.SecureMetrics {
		// FilterProvider is used to protect the metrics endpoint with authn/authz.
		// These configurations ensure that only authorized users and service accounts
		// can access the metrics endpoint. The RBAC are configured in 'config/rbac/kustomization.yaml'. More info:
		// https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.1/pkg/metrics/filters#WithAuthenticationAndAuthorization
		metricsServerOptions.FilterProvider = filters.WithAuthenticationAndAuthorization
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                        scheme,
		Metrics:                       metricsServerOptions,
		HealthProbeBindAddress:        cp.ProbeAddr,
		LeaderElection:                cp.EnableLeaderElection,
		BaseContext:                   func() context.Context { return ctx },
		LeaderElectionID:              "30f6fafb.k8s.icb4dc0.de",
		LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		return fmt.Errorf("unable to start control plane: %w", err)
	}

	envoySnapshotCache := cachev3.NewSnapshotCache(false, cachev3.IDHash{}, nil)

	envoySrv, err := cp.envoyServer(ctx, envoySnapshotCache)
	if err != nil {
		return err
	}

	if err := mgr.Add(envoySrv); err != nil {
		return fmt.Errorf("failed to add enovy server to manager: %w", err)
	}

	if err = (&controlplane.APIGatewayReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Cache:  envoySnapshotCache,
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("unable to create controller Core DB: %w", err)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		return fmt.Errorf("problem running manager: %w", err)
	}

	return nil
}

func (cp controlPlane) envoyServer(
	ctx context.Context,
	cache cachev3.SnapshotCache,
) (runnable mgr.Runnable, err error) {
	const (
		grpcKeepaliveTime        = 30 * time.Second
		grpcKeepaliveTimeout     = 5 * time.Second
		grpcKeepaliveMinTime     = 30 * time.Second
		grpcMaxConcurrentStreams = 1000000
	)

	var (
		logger = ctrl.Log.WithName("control-plane")
		srv    = server.NewServer(ctx, cache, nil)
	)

	// gRPC golang library sets a very small upper bound for the number gRPC/h2
	// streams over a single TCP connection. If a proxy multiplexes requests over
	// a single connection to the management server, then it might lead to
	// availability problems. Keepalive timeouts based on connection_keepalive parameter
	// https://www.envoyproxy.io/docs/envoy/latest/configuration/overview/examples#dynamic

	grpcOptions := append(make([]grpc.ServerOption, 0, 4),
		grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    grpcKeepaliveTime,
			Timeout: grpcKeepaliveTimeout,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             grpcKeepaliveMinTime,
			PermitWithoutStream: true,
		}),
	)
	grpcServer := grpc.NewServer(grpcOptions...)

	logger.Info("Opening listener", "addr", cp.ListenAddr)
	lis, err := net.Listen("tcp", cp.ListenAddr)
	if err != nil {
		return nil, fmt.Errorf("opening listener: %w", err)
	}

	logger.Info("Preparing health endpoints")

	healthService := grpchealth.NewServer()
	healthService.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	reflection.Register(grpcServer)
	discoverygrpc.RegisterAggregatedDiscoveryServiceServer(grpcServer, srv)
	endpointservice.RegisterEndpointDiscoveryServiceServer(grpcServer, srv)
	clusterservice.RegisterClusterDiscoveryServiceServer(grpcServer, srv)
	routeservice.RegisterRouteDiscoveryServiceServer(grpcServer, srv)
	listenerservice.RegisterListenerDiscoveryServiceServer(grpcServer, srv)
	secretservice.RegisterSecretDiscoveryServiceServer(grpcServer, srv)
	runtimeservice.RegisterRuntimeDiscoveryServiceServer(grpcServer, srv)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthService)

	return mgr.RunnableFunc(func(ctx context.Context) error {
		go func(ctx context.Context) {
			<-ctx.Done()
			grpcServer.GracefulStop()
		}(ctx)
		return grpcServer.Serve(lis)
	}), nil
}
