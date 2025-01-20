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

package controller

// +kubebuilder:rbac:groups=supabase.k8s.icb4dc0.de,resources=cores,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=supabase.k8s.icb4dc0.de,resources=cores/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=supabase.k8s.icb4dc0.de,resources=cores/finalizers,verbs=update
// +kubebuilder:rbac:groups=supabase.k8s.icb4dc0.de,resources=apigateways,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=supabase.k8s.icb4dc0.de,resources=apigateways/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=supabase.k8s.icb4dc0.de,resources=apigateways/finalizers,verbs=update
// +kubebuilder:rbac:groups=supabase.k8s.icb4dc0.de,resources=dashboards,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=supabase.k8s.icb4dc0.de,resources=dashboards/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=supabase.k8s.icb4dc0.de,resources=dashboards/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=secrets;configmaps;services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=create
