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
		WithDefaulter(&CoreCustomDefaulter{Client: mgr.GetClient()}).
		Complete()
}
