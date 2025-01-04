/*
Copyright 2024 Peter Kurfer.

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
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/alecthomas/kong"
	clusterservice "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	discoverygrpc "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	endpointservice "github.com/envoyproxy/go-control-plane/envoy/service/endpoint/v3"
	listenerservice "github.com/envoyproxy/go-control-plane/envoy/service/listener/v3"
	routeservice "github.com/envoyproxy/go-control-plane/envoy/service/route/v3"
	runtimeservice "github.com/envoyproxy/go-control-plane/envoy/service/runtime/v3"
	secretservice "github.com/envoyproxy/go-control-plane/envoy/service/secret/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"google.golang.org/grpc"
	grpchealth "google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"code.icb4dc0.de/prskr/supabase-operator/internal/controlplane"
)

type controlPlane struct {
	ListenAddr string `name:"listen-address" default:":18000" help:"The address the control plane binds to."`
}

func (p controlPlane) Run(ctx context.Context, cache cache.SnapshotCache) (err error) {
	const (
		grpcKeepaliveTime        = 30 * time.Second
		grpcKeepaliveTimeout     = 5 * time.Second
		grpcKeepaliveMinTime     = 30 * time.Second
		grpcMaxConcurrentStreams = 1000000
	)

	logger := ctrl.Log.WithName("control-plane")

	clientOpts := client.Options{
		Scheme: scheme,
	}

	logger.Info("Creating client")
	watcherClient, err := client.NewWithWatch(ctrl.GetConfigOrDie(), clientOpts)
	if err != nil {
		return err
	}

	srv := server.NewServer(ctx, cache, nil)

	// gRPC golang library sets a very small upper bound for the number gRPC/h2
	// streams over a single TCP connection. If a proxy multiplexes requests over
	// a single connection to the management server, then it might lead to
	// availability problems. Keepalive timeouts based on connection_keepalive parameter https://www.envoyproxy.io/docs/envoy/latest/configuration/overview/examples#dynamic

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

	logger.Info("Opening listener", "addr", p.ListenAddr)
	lis, err := net.Listen("tcp", p.ListenAddr)
	if err != nil {
		return fmt.Errorf("opening listener: %w", err)
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

	// discoverygrpc.AggregatedDiscoveryService_ServiceDesc.ServiceName

	endpointsController := controlplane.EndpointsController{
		Client: watcherClient,
		Cache:  cache,
	}

	errOut := make(chan error)

	go func(errOut chan<- error) {
		logger.Info("Starting gRPC server")
		errOut <- grpcServer.Serve(lis)
	}(errOut)

	go func(errOut chan<- error) {
		logger.Info("Staring endpoints controller")
		errOut <- endpointsController.Run(ctx)
	}(errOut)

	go func(errOut chan error) {
		for out := range errOut {
			err = errors.Join(err, out)
		}
	}(errOut)

	<-ctx.Done()
	grpcServer.Stop()

	return err
}

func (p controlPlane) AfterApply(kongctx *kong.Context) error {
	kongctx.BindTo(cache.NewSnapshotCache(false, cache.IDHash{}, nil), (*cache.SnapshotCache)(nil))
	return nil
}
