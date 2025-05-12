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
	"context"
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	supabasev1alpha1 "code.icb4dc0.de/prskr/supabase-operator/api/v1alpha1"
	"code.icb4dc0.de/prskr/supabase-operator/internal/oidc"
	"code.icb4dc0.de/prskr/supabase-operator/internal/supabase"
)

// +kubebuilder:webhook:path=/mutate-supabase-k8s-icb4dc0-de-v1alpha1-apigateway,mutating=true,failurePolicy=fail,sideEffects=None,groups=supabase.k8s.icb4dc0.de,resources=apigateways,verbs=create;update,versions=v1alpha1,name=mapigateway-v1alpha1.kb.io,admissionReviewVersions=v1

// APIGatewayCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind APIGateway when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type APIGatewayCustomDefaulter struct {
	CurrentNamespace string
	Recorder         record.EventRecorder
}

var _ webhook.CustomDefaulter = &APIGatewayCustomDefaulter{}

var errObjectTypeMismatch = errors.New("object type mismatch")

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind APIGateway.
func (d *APIGatewayCustomDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	const (
		defaultManagerNamespace = "supabase-system"
	)

	logger := log.FromContext(ctx)
	apiGateway, ok := obj.(*supabasev1alpha1.APIGateway)

	if !ok {
		return fmt.Errorf("%w: expected an APIGateway object but got %T", errObjectTypeMismatch, obj)
	}
	apigatewaylog.Info("Defaulting for APIGateway", "name", apiGateway.GetName())

	if apiGateway.Spec.ApiEndpoint == nil {
		apiGateway.Spec.ApiEndpoint = new(supabasev1alpha1.ApiEndpointSpec)
	}

	if apiGateway.Spec.ApiEndpoint.JWKSSelector == nil {
		apiGateway.Spec.ApiEndpoint.JWKSSelector = &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: supabase.ServiceConfig.JWT.ObjectName(apiGateway),
			},
			Key: supabase.ServiceConfig.JWT.Defaults.JwksKey,
		}
	}

	if apiGateway.Spec.Envoy == nil {
		apiGateway.Spec.Envoy = new(supabasev1alpha1.EnvoySpec)
	}

	if apiGateway.Spec.Envoy.NodeName == "" {
		apiGateway.Spec.Envoy.NodeName = apiGateway.Name
	}

	if apiGateway.Spec.Envoy.ControlPlane == nil {
		if d.CurrentNamespace == defaultManagerNamespace {
			d.Recorder.Event(
				apiGateway,
				corev1.EventTypeNormal,
				"Guessing Envoy control plane endpoint",
				"Making guess of control plane config, most likely this is correct as the current namespace is the default namespace where the operator is deployed but of course it could be wrong as well",
			)

			apiGateway.Spec.Envoy.ControlPlane = &supabasev1alpha1.ControlPlaneSpec{
				Host: "supabase-control-plane.supabase-system.svc",
				Port: 18000,
			}
		} else {
			d.Recorder.Eventf(
				apiGateway,
				corev1.EventTypeWarning,
				"Guessing Envoy control plane endpoint",
				"Making guess of control plane config based on the namespace of the manager (%s) - could be wrong if control plane was manually deployed to another namespace",
				d.CurrentNamespace,
			)

			apiGateway.Spec.Envoy.ControlPlane = &supabasev1alpha1.ControlPlaneSpec{
				Host: fmt.Sprintf("supabase-control-plane.%s.svc", d.CurrentNamespace),
				Port: 18000,
			}
		}
	}

	if oauth2Spec := apiGateway.Spec.DashboardEndpoint.OAuth2(); oauth2Spec != nil {
		if oauth2Spec.OpenIDIssuer != "" {
			logger.Info("Fetching OIDC discovery document", "discovery_url", oauth2Spec.OpenIDIssuer)
			discoveryDoc, err := oidc.IssuerConfiguration(ctx, oauth2Spec.OpenIDIssuer)
			if err != nil {
				return fmt.Errorf("failed to fetch OIDC configuration: %w", err)
			}

			oauth2Spec.TokenEndpoint = discoveryDoc.TokenEndpoint
			oauth2Spec.AuthorizationEndpoint = discoveryDoc.AuthorizationEndpoint
		}
	}

	return nil
}
