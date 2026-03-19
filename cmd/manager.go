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
	"os"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/metrics/filters"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/prskr/supabase-operator/internal/controller"
	webhooksupabasev1alpha1 "github.com/prskr/supabase-operator/internal/webhook/v1alpha1"
)

//nolint:lll // flag declaration needs to be in tags
type manager struct {
	MetricsAddr          string `name:"metrics-bind-address" default:"0" help:"The address the metrics endpoint binds to. Use :8443 for HTTPS or :8080 for HTTP, or leave as 0 to disable the metrics service."`
	EnableLeaderElection bool   `name:"leader-elect" default:"false" help:"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager."`
	ProbeAddr            string `name:"health-probe-bind-address" default:":8081" help:"The address the probe endpoint binds to."`
	SecureMetrics        bool   `name:"metrics-secure" default:"true" help:"If set, the metrics endpoint is served securely via HTTPS. Use --metrics-secure=false to use HTTP instead."`
	EnableHTTP2          bool   `name:"enable-http2" default:"false" help:"If set, HTTP/2 will be enabled for the metrics and webhook servers"`
	Namespace            string `name:"controller-namespace" env:"CONTROLLER_NAMESPACE" default:"" help:"Namespace where the controller is running, ideally set via downward API"`
	TLS                  struct {
		CACert FileContent `env:"CA_CERT" name:"ca-cert" required:"" help:"The path to the CA certificate file."`
		CAKey  FileContent `env:"CA_KEY" name:"ca-key" required:"" help:"The path to the CA key file."`
	} `embed:"" prefix:"tls." envprefix:"TLS_"`
}

func (m manager) Run(ctx context.Context) error {
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

	if !m.EnableHTTP2 {
		tlsOpts = append(tlsOpts, disableHTTP2)
	}

	webhookConfig := webhooksupabasev1alpha1.WebhookConfig{
		CurrentNamespace: m.Namespace,
	}

	webhookServer := webhook.NewServer(webhook.Options{
		TLSOpts: tlsOpts,
	})

	caCert, err := tls.X509KeyPair(m.TLS.CACert, m.TLS.CAKey)
	if err != nil {
		return fmt.Errorf("unable to load CA cert: %w", err)
	}

	// Metrics endpoint is enabled in 'config/default/kustomization.yaml'. The Metrics options configure the server.
	// More info:
	// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.1/pkg/metrics/server
	// - https://book.kubebuilder.io/reference/metrics.html
	metricsServerOptions := metricsserver.Options{
		BindAddress:   m.MetricsAddr,
		SecureServing: m.SecureMetrics,
		TLSOpts:       tlsOpts,
	}

	if m.SecureMetrics {
		// FilterProvider is used to protect the metrics endpoint with authn/authz.
		// These configurations ensure that only authorized users and service accounts
		// can access the metrics endpoint. The RBAC are configured in 'config/rbac/kustomization.yaml'. More info:
		// https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.1/pkg/metrics/filters#WithAuthenticationAndAuthorization
		metricsServerOptions.FilterProvider = filters.WithAuthenticationAndAuthorization
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                        scheme,
		Metrics:                       metricsServerOptions,
		WebhookServer:                 webhookServer,
		HealthProbeBindAddress:        m.ProbeAddr,
		LeaderElection:                m.EnableLeaderElection,
		BaseContext:                   func() context.Context { return ctx },
		LeaderElectionID:              "05f9463f.k8s.icb4dc0.de",
		LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		return fmt.Errorf("unable to start manager: %w", err)
	}

	if err = (&controller.CoreDbReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("unable to create controller Core DB: %w", err)
	}

	if err = (&controller.CoreJwtReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("unable to create controller Core JWT: %w", err)
	}

	if err = (&controller.CorePostgrestReconiler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("unable to create controller Core Postgrest: %w", err)
	}

	if err = (&controller.CoreAuthReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("unable to create controller Core Auth: %w", err)
	}

	if err = (&controller.DashboardPGMetaReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("unable to create controller Dashboard PG-Meta: %w", err)
	}

	if err = (&controller.DashboardStudioReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("unable to create controller Dashboard PG-Meta: %w", err)
	}

	if err = (&controller.APIGatewayReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		CACert: caCert,
	}).SetupWithManager(ctx, mgr); err != nil {
		return fmt.Errorf("unable to create controller APIGateway: %w", err)
	}

	if err = (&controller.StorageApiReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("unable to create controller APIGateway: %w", err)
	}

	if err = (&controller.StorageImgProxyReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("unable to create controller APIGateway: %w", err)
	}

	if err = (&controller.StorageS3CredentialsReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("unable to create controller APIGateway: %w", err)
	}

	// nolint:goconst
	if os.Getenv("ENABLE_WEBHOOKS") != "false" {
		if err = webhooksupabasev1alpha1.SetupCoreWebhookWithManager(mgr); err != nil {
			return fmt.Errorf("unable to create webhook: %w", err)
		}

		if err = webhooksupabasev1alpha1.SetupAPIGatewayWebhookWithManager(mgr, webhookConfig); err != nil {
			return fmt.Errorf("unable to create webhook: %w", err)
		}

		if err = webhooksupabasev1alpha1.SetupDashboardWebhookWithManager(mgr); err != nil {
			return fmt.Errorf("unable to create webhook: %w", err)
		}

		if err = webhooksupabasev1alpha1.SetupStorageWebhookWithManager(mgr); err != nil {
			return fmt.Errorf("unable to create webhook: %w", err)
		}
	}

	// +kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
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
