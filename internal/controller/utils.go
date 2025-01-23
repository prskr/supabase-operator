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

import (
	"context"
	"hash/fnv"
	"maps"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"code.icb4dc0.de/prskr/supabase-operator/api"
)

func ptrOf[T any](val T) *T {
	return &val
}

func boolValueOf(ptr *bool) bool {
	if ptr == nil {
		return false
	}

	return *ptr
}

func MergeLabels(source map[string]string, toAppend ...map[string]string) map[string]string {
	result := make(map[string]string, len(source)+len(toAppend))
	maps.Copy(result, source)

	for _, additionalLabels := range toAppend {
		for k, v := range additionalLabels {
			if _, exists := result[k]; exists {
				continue
			}

			result[k] = v
		}
	}

	return result
}

func MergeEnv(source []corev1.EnvVar, toAppend ...corev1.EnvVar) []corev1.EnvVar {
	existingKeys := make(map[string]bool, len(source)+len(toAppend))

	merged := append(make([]corev1.EnvVar, 0, len(source)+len(toAppend)), source...)

	for _, v := range source {
		existingKeys[v.Name] = true
	}

	for _, v := range toAppend {
		if _, alreadyPresent := existingKeys[v.Name]; alreadyPresent {
			continue
		}
		merged = append(merged, v)
		existingKeys[v.Name] = true
	}

	return merged
}

func ValueOrFallback[T any](value, fallback T) T {
	rval := reflect.ValueOf(value)
	if rval.IsZero() {
		return fallback
	}

	return value
}

func HashStrings(vals ...string) []byte {
	h := fnv.New64a()

	for _, v := range vals {
		h.Write([]byte(v))
	}

	return h.Sum(nil)
}

func HashBytes(vals ...[]byte) []byte {
	h := fnv.New64a()

	for _, v := range vals {
		h.Write(v)
	}

	return h.Sum(nil)
}

func FieldSelectorEventHandler[TItem metav1.Object, TList api.ObjectList[TItem]](
	cli client.Client,
	fieldSelector string,
) handler.TypedEventHandler[client.Object, reconcile.Request] {
	return handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
		var (
			list             TList
			selectorInstance = client.MatchingFieldsSelector{
				Selector: fields.OneTermEqualSelector(fieldSelector, obj.GetName()),
			}
			logger = log.FromContext(ctx, "object", obj.GetName(), "namespace", obj.GetNamespace())
		)

		list = reflect.New(reflect.TypeOf(list).Elem()).Interface().(TList)

		if err := cli.List(ctx, list, selectorInstance, client.InNamespace(obj.GetNamespace())); err != nil {
			logger.Error(err, "could not list items for field selector event handler")
			return nil
		}

		requests := make([]reconcile.Request, 0)

		for item := range list.Iter() {
			requests = append(requests, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      item.GetName(),
					Namespace: obj.GetNamespace(),
				},
			})
		}

		return requests
	})
}
