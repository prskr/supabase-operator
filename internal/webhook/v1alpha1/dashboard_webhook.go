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

package v1alpha1

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	supabasev1alpha1 "code.icb4dc0.de/prskr/supabase-operator/api/v1alpha1"
)

// nolint:unused
// log is for logging in this package.
var dashboardlog = logf.Log.WithName("dashboard-resource")

// +kubebuilder:webhook:path=/mutate-supabase-k8s-icb4dc0-de-v1alpha1-dashboard,mutating=true,failurePolicy=fail,sideEffects=None,groups=supabase.k8s.icb4dc0.de,resources=dashboards,verbs=create;update,versions=v1alpha1,name=mdashboard-v1alpha1.kb.io,admissionReviewVersions=v1

// DashboardCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind Dashboard when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type DashboardCustomDefaulter struct {
	// TODO(user): Add more fields as needed for defaulting
}

var _ webhook.CustomDefaulter = &DashboardCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind Dashboard.
func (d *DashboardCustomDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	dashboard, ok := obj.(*supabasev1alpha1.Dashboard)

	if !ok {
		return fmt.Errorf("expected an Dashboard object but got %T", obj)
	}
	dashboardlog.Info("Defaulting for Dashboard", "name", dashboard.GetName())

	// TODO(user): fill in your defaulting logic.

	return nil
}

// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-supabase-k8s-icb4dc0-de-v1alpha1-dashboard,mutating=false,failurePolicy=fail,sideEffects=None,groups=supabase.k8s.icb4dc0.de,resources=dashboards,verbs=create;update,versions=v1alpha1,name=vdashboard-v1alpha1.kb.io,admissionReviewVersions=v1

// DashboardCustomValidator struct is responsible for validating the Dashboard resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type DashboardCustomValidator struct {
	// TODO(user): Add more fields as needed for validation
}

var _ webhook.CustomValidator = &DashboardCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type Dashboard.
func (v *DashboardCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	dashboard, ok := obj.(*supabasev1alpha1.Dashboard)
	if !ok {
		return nil, fmt.Errorf("expected a Dashboard object but got %T", obj)
	}
	dashboardlog.Info("Validation for Dashboard upon creation", "name", dashboard.GetName())

	// TODO(user): fill in your validation logic upon object creation.

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type Dashboard.
func (v *DashboardCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	dashboard, ok := newObj.(*supabasev1alpha1.Dashboard)
	if !ok {
		return nil, fmt.Errorf("expected a Dashboard object for the newObj but got %T", newObj)
	}
	dashboardlog.Info("Validation for Dashboard upon update", "name", dashboard.GetName())

	// TODO(user): fill in your validation logic upon object update.

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type Dashboard.
func (v *DashboardCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	dashboard, ok := obj.(*supabasev1alpha1.Dashboard)
	if !ok {
		return nil, fmt.Errorf("expected a Dashboard object but got %T", obj)
	}
	dashboardlog.Info("Validation for Dashboard upon deletion", "name", dashboard.GetName())

	// TODO(user): fill in your validation logic upon object deletion.

	return nil, nil
}
