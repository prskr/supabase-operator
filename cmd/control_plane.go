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
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	clusterservice "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	discoverygrpc "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	endpointservice "github.com/envoyproxy/go-control-plane/envoy/service/endpoint/v3"
	listenerservice "github.com/envoyproxy/go-control-plane/envoy/service/listener/v3"
	routeservice "github.com/envoyproxy/go-control-plane/envoy/service/route/v3"
	runtimeservice "github.com/envoyproxy/go-control-plane/envoy/service/runtime/v3"
	secretservice "github.com/envoyproxy/go-control-plane/envoy/service/secret/v3"
	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/log"
	"github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"google.golang.org/grpc/credentials"

	"github.com/go-logr/logr"
	"google.golang.org/grpc"
	grpchealth "google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	mgr "sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/metrics/filters"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/prskr/supabase-operator/internal/certs"
	"github.com/prskr/supabase-operator/internal/controlplane"
	"github.com/prskr/supabase-operator/internal/health"
)

var errCouldNoParseCert = errors.New("could not parse certificate")

//nolint:lll // flag declaration with struct tags is as long as it is
type controlPlane struct {
	caCert tls.Certificate `kong:"-"`

	ListenAddr string `name:"listen-address" default:":18000" help:"The address the control plane binds to."`
	Tls        struct {
		CA struct {
			Cert FileContent `env:"CERT" name:"server-cert" required:"" help:"The path to the server certificate file."`
			Key  FileContent `env:"KEY" name:"server-key" required:"" help:"The path to the server key file."`
		} `embed:"" prefix:"ca." envprefix:"CA_"`
		ServerSecretName string `name:"server-secret-name" help:"The name of the secret containing the server certificate and key." default:"control-plane-xds-tls"`
	} `embed:"" prefix:"tls." envprefix:"TLS_"`
	MetricsAddr          string `name:"metrics-bind-address" default:"0" help:"The address the metrics endpoint binds to. Use :8443 for HTTPS or :8080 for HTTP, or leave as 0 to disable the metrics service."`
	EnableLeaderElection bool   `name:"leader-elect" default:"false" help:"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager."`
	ProbeAddr            string `name:"health-probe-bind-address" default:":8081" help:"The address the probe endpoint binds to."`
	SecureMetrics        bool   `name:"metrics-secure" default:"true" help:"If set, the metrics endpoint is served securely via HTTPS. Use --metrics-secure=false to use HTTP instead."`
	EnableHTTP2          bool   `name:"enable-http2" default:"false" help:"If set, HTTP/2 will be enabled for the metrics and webhook servers"`
	ServiceName          string `name:"service-name" env:"CONTROL_PLANE_SERVICE_NAME" default:"" required:"" help:"The name of the control plane service."`
	Namespace            string `name:"namespace" env:"CONTROL_PLANE_NAMESPACE" default:"" required:"" help:"Namespace where the controller is running, ideally set via downward API"`
}

func (cp *controlPlane) Run(ctx context.Context, logger logr.Logger) error {
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

	//nolint:contextcheck
	bootstrapClient, err := client.New(ctrl.GetConfigOrDie(), client.Options{Scheme: scheme})
	if err != nil {
		return fmt.Errorf("unable to create bootstrap client: %w", err)
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

	cacheLoggerInst := cacheLogger(logger.WithName("envoy-snapshot-cache"))
	envoySnapshotCache := cachev3.NewSnapshotCache(false, cachev3.IDHash{}, cacheLoggerInst)

	serverCert, err := cp.ensureControlPlaneTLSCert(ctx, bootstrapClient)
	if err != nil {
		return fmt.Errorf("failed to ensure control plane TLS cert: %w", err)
	}

	envoySrv, err := cp.envoyServer(ctx, logger, envoySnapshotCache, serverCert)
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

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return fmt.Errorf("unable to set up health check: %w", err)
	}

	if err := mgr.AddHealthzCheck("server-cert", health.CertValidCheck(serverCert)); err != nil {
		return fmt.Errorf("unable to set up health check: %w", err)
	}

	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		return fmt.Errorf("unable to set up ready check: %w", err)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		return fmt.Errorf("problem running manager: %w", err)
	}

	return nil
}

func (cp *controlPlane) AfterApply() (err error) {
	cp.caCert, err = tls.X509KeyPair(cp.Tls.CA.Cert, cp.Tls.CA.Key)
	if err != nil {
		return fmt.Errorf("failed to parse server certificate: %w", err)
	}

	return nil
}

func (cp *controlPlane) envoyServer(
	ctx context.Context,
	logger logr.Logger,
	cache cachev3.SnapshotCache,
	serverCert tls.Certificate,
) (runnable mgr.Runnable, err error) {
	const (
		grpcKeepaliveTime        = 30 * time.Second
		grpcKeepaliveTimeout     = 5 * time.Second
		grpcKeepaliveMinTime     = 30 * time.Second
		grpcMaxConcurrentStreams = 1000000
	)
	srv := server.NewServer(ctx, cache, xdsServerCallbacks(logger))
	logger = logger.WithName("control-plane")

	// gRPC golang library sets a very small upper bound for the number gRPC/h2
	// streams over a single TCP connection. If a proxy multiplexes requests over
	// a single connection to the management server, then it might lead to
	// availability problems. Keepalive timeouts based on connection_keepalive parameter
	// https://www.envoyproxy.io/docs/envoy/latest/configuration/overview/examples#dynamic

	tlsCfg, err := cp.tlsConfig(serverCert)
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS config: %w", err)
	}

	grpcOptions := append(make([]grpc.ServerOption, 0, 4),
		grpc.Creds(credentials.NewTLS(tlsCfg)),
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
		//nolint:contextcheck
		return grpcServer.Serve(lis)
	}), nil
}

func (cp *controlPlane) ensureControlPlaneTLSCert(
	ctx context.Context,
	k8sClient client.Client,
) (tls.Certificate, error) {
	var (
		controlPlaneServerCert = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cp.Tls.ServerSecretName,
				Namespace: cp.Namespace,
			},
		}
		serverCert tls.Certificate
	)

	_, err := controllerutil.CreateOrUpdate(ctx, k8sClient, controlPlaneServerCert, func() (err error) {
		controlPlaneServerCert.Type = corev1.SecretTypeTLS

		if controlPlaneServerCert.Data == nil {
			controlPlaneServerCert.Data = make(map[string][]byte, 3)
		}

		var (
			cert       = controlPlaneServerCert.Data[corev1.TLSCertKey]
			privateKey = controlPlaneServerCert.Data[corev1.TLSPrivateKeyKey]
		)

		var requireRenewal bool
		if cert != nil && privateKey != nil {
			if serverCert, err = tls.X509KeyPair(cert, privateKey); err != nil {
				return fmt.Errorf("failed to parse server certificate: %w", err)
			}

			renewGracePeriod := time.Duration(float64(serverCert.Leaf.NotAfter.Sub(serverCert.Leaf.NotBefore)) * 0.1)
			if serverCert.Leaf.NotAfter.After(time.Now().Add(renewGracePeriod)) {
				requireRenewal = true
			}
		} else {
			requireRenewal = true
		}

		if requireRenewal {
			dnsNames := []string{
				strings.Join([]string{cp.ServiceName, cp.Namespace, "svc"}, "."),
				strings.Join([]string{cp.ServiceName, cp.Namespace, "svc", "cluster", "local"}, "."),
			}
			if certResult, err := certs.ServerCert("supabase-control-plane", dnsNames, cp.caCert); err != nil {
				return fmt.Errorf("failed to generate server certificate: %w", err)
			} else {
				serverCert = certResult.ServerCert
				controlPlaneServerCert.Data[corev1.TLSCertKey] = certResult.PublicKey
				controlPlaneServerCert.Data[corev1.TLSPrivateKeyKey] = certResult.PrivateKey
			}
		}

		return nil
	})
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to create or update control plane server certificate: %w", err)
	}

	return serverCert, nil
}

func (cp *controlPlane) tlsConfig(serverCert tls.Certificate) (*tls.Config, error) {
	tlsCfg := &tls.Config{
		RootCAs:    x509.NewCertPool(),
		ClientCAs:  x509.NewCertPool(),
		ClientAuth: tls.RequireAndVerifyClientCert,
	}

	tlsCfg.Certificates = append(tlsCfg.Certificates, serverCert)
	if !tlsCfg.RootCAs.AppendCertsFromPEM(cp.Tls.CA.Cert) {
		return nil, fmt.Errorf("%w: root CA certificate", errCouldNoParseCert)
	}

	if !tlsCfg.ClientCAs.AppendCertsFromPEM(cp.Tls.CA.Cert) {
		return nil, fmt.Errorf("%w: client CA certificate", errCouldNoParseCert)
	}

	return tlsCfg, nil
}

func xdsServerCallbacks(logger logr.Logger) server.Callbacks {
	return server.CallbackFuncs{
		StreamOpenFunc: func(ctx context.Context, streamId int64, nodeId string) error {
			logger.Info("Stream opened", "stream-id", streamId, "node-id", nodeId)
			return nil
		},
		StreamClosedFunc: func(streamId int64, node *corev3.Node) {
			logger.Info("Stream closed", "stream-id", streamId,
				"node.id", node.Id,
				"node.cluster", node.Cluster,
			)
		},
		StreamRequestFunc: func(streamId int64, request *discoverygrpc.DiscoveryRequest) error {
			logger.Info("Stream request",
				"stream-id", streamId,
				"request.node.id", request.Node.Id,
				"request.node.cluster", request.Node.Cluster,
				"request.version", request.VersionInfo,
				"request.error", request.ErrorDetail,
			)
			return nil
		},
		StreamResponseFunc: func(
			ctx context.Context,
			streamId int64,
			request *discoverygrpc.DiscoveryRequest,
			response *discoverygrpc.DiscoveryResponse,
		) {
			logger.Info("Stream delta response",
				"stream-id", streamId,
				"request.node.id", request.Node.Id,
				"request.node.cluster", request.Node.Cluster,
			)
		},
		DeltaStreamOpenFunc: func(ctx context.Context, streamId int64, nodeId string) error {
			logger.Info("Delta stream opened", "stream-id", streamId, "node-id", nodeId)
			return nil
		},

		DeltaStreamClosedFunc: func(streamId int64, node *corev3.Node) {
			logger.Info("Delta stream closed",
				"stream-id", streamId,
				"node.id", node.Id,
				"node.cluster", node.Cluster,
			)
		},
		StreamDeltaRequestFunc: func(i int64, request *discoverygrpc.DeltaDiscoveryRequest) error {
			logger.Info("Stream delta request",
				"stream-id", i,
				"request.node.id", request.Node.Id,
				"request.node.cluster", request.Node.Cluster,
				"request.error", request.ErrorDetail,
			)
			return nil
		},
		StreamDeltaResponseFunc: func(
			i int64,
			request *discoverygrpc.DeltaDiscoveryRequest,
			response *discoverygrpc.DeltaDiscoveryResponse,
		) {
			logger.Info("Stream delta response",
				"stream-id", i,
				"request.node", request.Node,
				"response.resources", response.Resources,
			)
		},
	}
}

func cacheLogger(logger logr.Logger) log.Logger {
	wrapper := func(delegate func(msg string, keysAndValues ...any)) func(string, ...any) {
		return func(s string, i ...any) {
			delegate(fmt.Sprintf(s, i...))
		}
	}

	return log.LoggerFuncs{
		DebugFunc: nil, // enable for debug info
		InfoFunc:  wrapper(logger.Info),
		WarnFunc:  wrapper(logger.Info),
		ErrorFunc: wrapper(logger.Info),
	}
}
