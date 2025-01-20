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

	"k8s.io/apimachinery/pkg/runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	supabasev1alpha1 "code.icb4dc0.de/prskr/supabase-operator/api/v1alpha1"
)

// nolint:unused
// log is for logging in this package.
var apigatewaylog = logf.Log.WithName("apigateway-resource")

var (
	ErrMissingEnvoySpec        = errors.New("envoy needs to be configured")
	ErrMissingControlPlaneSpec = errors.New("envoy control plane needs to be configured")
)

// +kubebuilder:webhook:path=/validate-supabase-k8s-icb4dc0-de-v1alpha1-apigateway,mutating=false,failurePolicy=fail,sideEffects=None,groups=supabase.k8s.icb4dc0.de,resources=apigateways,verbs=create;update,versions=v1alpha1,name=vapigateway-v1alpha1.kb.io,admissionReviewVersions=v1

// APIGatewayCustomValidator struct is responsible for validating the APIGateway resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type APIGatewayCustomValidator struct{}

var _ webhook.CustomValidator = &APIGatewayCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type APIGateway.
func (v *APIGatewayCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	apigateway, ok := obj.(*supabasev1alpha1.APIGateway)
	if !ok {
		return nil, fmt.Errorf("expected a APIGateway object but got %T", obj)
	}
	apigatewaylog.Info("Validation for APIGateway upon creation", "name", apigateway.GetName())

	return validateEnvoyControlPlane(apigateway)
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type APIGateway.
func (v *APIGatewayCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	apigateway, ok := newObj.(*supabasev1alpha1.APIGateway)
	if !ok {
		return nil, fmt.Errorf("expected a APIGateway object for the newObj but got %T", newObj)
	}
	apigatewaylog.Info("Validation for APIGateway upon update", "name", apigateway.GetName())

	return validateEnvoyControlPlane(apigateway)
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type APIGateway.
func (v *APIGatewayCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	apigateway, ok := obj.(*supabasev1alpha1.APIGateway)
	if !ok {
		return nil, fmt.Errorf("expected a APIGateway object but got %T", obj)
	}
	apigatewaylog.Info("Validation for APIGateway upon deletion", "name", apigateway.GetName())

	return nil, nil
}

func validateEnvoyControlPlane(gateway *supabasev1alpha1.APIGateway) (admission.Warnings, error) {
	envoySpec := gateway.Spec.Envoy

	if envoySpec == nil {
		return nil, ErrMissingEnvoySpec
	}

	if envoySpec.ControlPlane == nil {
		return nil, ErrMissingControlPlaneSpec
	}

	return nil, nil
}
