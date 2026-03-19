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
	"maps"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/prskr/supabase-operator/internal/meta"
)

func selectorLabels(
	object client.Object,
	name string,
) map[string]string {
	return map[string]string{
		meta.WellKnownLabel.Name:     name,
		meta.WellKnownLabel.Instance: object.GetName(),
		meta.WellKnownLabel.PartOf:   "supabase",
	}
}

func objectLabels(
	object client.Object,
	name,
	component,
	version string,
) map[string]string {
	labels := maps.Clone(object.GetLabels())
	if labels == nil {
		labels = make(map[string]string, 6)
	}

	labels[meta.WellKnownLabel.Name] = name
	labels[meta.WellKnownLabel.Instance] = object.GetName()
	labels[meta.WellKnownLabel.Version] = version
	labels[meta.WellKnownLabel.Component] = component
	labels[meta.WellKnownLabel.PartOf] = "supabase"
	labels[meta.WellKnownLabel.ManagedBy] = "supabase-operator"

	return labels
}
