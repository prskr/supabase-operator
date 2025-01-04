package controller

import (
	"maps"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"code.icb4dc0.de/prskr/supabase-operator/internal/meta"
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
