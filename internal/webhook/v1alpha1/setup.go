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

package v1alpha1

import (
	ctrl "sigs.k8s.io/controller-runtime"

	supabasev1alpha1 "code.icb4dc0.de/prskr/supabase-operator/api/v1alpha1"
)

type WebhookConfig struct {
	CurrentNamespace string
}

// SetupAPIGatewayWebhookWithManager registers the webhook for APIGateway in the manager.
func SetupAPIGatewayWebhookWithManager(mgr ctrl.Manager, cfg WebhookConfig) error {
	mgr.GetEventRecorderFor("apigateway-defaulter")
	return ctrl.NewWebhookManagedBy(mgr).For(&supabasev1alpha1.APIGateway{}).
		WithValidator(&APIGatewayCustomValidator{}).
		WithDefaulter(&APIGatewayCustomDefaulter{
			CurrentNamespace: cfg.CurrentNamespace,
			Recorder:         mgr.GetEventRecorderFor("apigateway-defaulter"),
		}).
		Complete()
}

// SetupCoreWebhookWithManager registers the webhook for Core in the manager.
func SetupCoreWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&supabasev1alpha1.Core{}).
		WithValidator(&CoreCustomValidator{Client: mgr.GetClient()}).
		WithDefaulter(&CoreCustomDefaulter{Client: mgr.GetClient(), Scheme: mgr.GetScheme()}).
		Complete()
}

// SetupDashboardWebhookWithManager registers the webhook for Dashboard in the manager.
func SetupDashboardWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&supabasev1alpha1.Dashboard{}).
		WithValidator(&DashboardCustomValidator{}).
		WithDefaulter(&DashboardCustomDefaulter{}).
		Complete()
}

// SetupStorageWebhookWithManager registers the webhook for Storage in the manager.
func SetupStorageWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&supabasev1alpha1.Storage{}).
		WithValidator(&StorageCustomValidator{Client: mgr.GetClient()}).
		WithDefaulter(&StorageCustomDefaulter{Client: mgr.GetClient()}).
		Complete()
}
