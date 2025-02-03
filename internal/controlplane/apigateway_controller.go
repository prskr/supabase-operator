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

package controlplane

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	supabasev1alpha1 "code.icb4dc0.de/prskr/supabase-operator/api/v1alpha1"
	"code.icb4dc0.de/prskr/supabase-operator/internal/meta"
)

// APIGatewayReconciler reconciles a APIGateway object
type APIGatewayReconciler struct {
	initialReconciliation atomic.Bool
	client.Client
	Scheme *runtime.Scheme
	Cache  cachev3.SnapshotCache
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.1/pkg/reconcile
func (r *APIGatewayReconciler) Reconcile(ctx context.Context, req ctrl.Request) (res ctrl.Result, err error) {
	var (
		gateway           supabasev1alpha1.APIGateway
		logger            = log.FromContext(ctx)
		endpointSliceList discoveryv1.EndpointSliceList
	)

	logger.Info("Reconciling Envoy control-plane config")

	if err := r.Get(ctx, req.NamespacedName, &gateway); client.IgnoreNotFound(err) != nil {
		logger.Error(err, "unable to fetch Gateway")
		return ctrl.Result{}, err
	}

	selector, err := metav1.LabelSelectorAsSelector(gateway.Spec.ServiceSelector)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to create selector for EndpointSlices: %w", err)
	}

	if err := r.List(ctx, &endpointSliceList, client.MatchingLabelsSelector{Selector: selector}); err != nil {
		return ctrl.Result{}, err
	}

	services := EnvoyServices{
		ServiceLabelKey: gateway.Spec.ComponentTypeLabel,
		Gateway:         &gateway,
		Client:          r.Client,
	}
	services.UpsertEndpointSlices(endpointSliceList.Items...)

	instance := fmt.Sprintf("%s:%s", gateway.Spec.Envoy.NodeName, gateway.Namespace)

	logger.Info("Computing Envoy snapshot for current service targets", "version", gateway.Status.Envoy.ConfigVersion)
	snapshot, snapshotHash, err := services.snapshot(ctx, instance, gateway.Status.Envoy.ConfigVersion)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to prepare snapshot: %w", err)
	}

	if !r.initialReconciliation.CompareAndSwap(false, true) && bytes.Equal(gateway.Status.Envoy.ResourceHash, snapshotHash) {
		logger.Info("No changes detected, skipping update")
		return ctrl.Result{}, nil
	}

	logger.Info("Updating service targets")
	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, &gateway, func() error {
		gateway.Status.ServiceTargets = services.Targets()
		gateway.Status.Envoy.ConfigVersion = strconv.FormatInt(time.Now().UTC().UnixMilli(), 10)
		gateway.Status.Envoy.ResourceHash = snapshotHash

		return nil
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("Propagating Envoy snapshot", "version", gateway.Status.Envoy.ConfigVersion)
	if err := r.Cache.SetSnapshot(ctx, instance, snapshot); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to propagate snapshot: %w", err)
	}

	return ctrl.Result{}, nil
}

func (r *APIGatewayReconciler) SetupWithManager(mgr ctrl.Manager) error {
	gatewayTargetLabelSelector, err := predicate.LabelSelectorPredicate(metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{{
			Key:      meta.SupabaseLabel.ApiGatewayTarget,
			Operator: metav1.LabelSelectorOpExists,
		}},
	})
	if err != nil {
		return fmt.Errorf("failed to build gateway target predicate: %w", err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(new(supabasev1alpha1.APIGateway), builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Owns(new(corev1.Secret), builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Watches(
			new(discoveryv1.EndpointSlice),
			r.endpointSliceEventHandler(),
			builder.WithPredicates(gatewayTargetLabelSelector)).
		Complete(r)
}

// endpointSliceEventHandler - prepares an event handler that checks whether the EndpointSlice has a specific target
// or if it is targeting the only APIGateway in its namespace (default behavior for the operator)
func (r *APIGatewayReconciler) endpointSliceEventHandler() handler.TypedEventHandler[client.Object, reconcile.Request] {
	return handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
		var (
			logger         = log.FromContext(ctx)
			apiGatewayList supabasev1alpha1.APIGatewayList
		)

		endpointSlice, ok := obj.(*discoveryv1.EndpointSlice)
		if !ok {
			logger.Info("Cannot map event to reconcile request, because object has unexpected type", "type", fmt.Sprintf("%T", obj))
			return nil
		}

		logger.Info("Triggering APIGateway reconciliation", "obj_name", obj.GetName(), "obj_namespace", obj.GetNamespace())
		if err := r.Client.List(ctx, &apiGatewayList, client.InNamespace(endpointSlice.Namespace)); err != nil {
			logger.Error(err, "failed to list APIGateways to determine reconcile targets")
			return nil
		}

		target, ok := endpointSlice.Labels[meta.SupabaseLabel.ApiGatewayTarget]
		if !ok {
			// should not happen, just to be sure
			return nil
		}

		var reconcileRequests []reconcile.Request

		if target != "" {
			for _, gw := range apiGatewayList.Items {
				if strings.EqualFold(gw.Spec.Envoy.NodeName, target) {
					reconcileRequests = append(reconcileRequests, reconcile.Request{
						NamespacedName: types.NamespacedName{
							Name:      gw.Name,
							Namespace: gw.Namespace,
						},
					})
				}
			}
		} else {
			reconcileRequests = make([]reconcile.Request, 0, len(apiGatewayList.Items))
			for _, gw := range apiGatewayList.Items {
				reconcileRequests = append(reconcileRequests, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      gw.Name,
						Namespace: gw.Namespace,
					},
				})
			}
		}

		return reconcileRequests
	})
}
