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
	"maps"

	corev1 "k8s.io/api/core/v1"
)

type ImageSpec struct {
	Image      string            `json:"image,omitempty"`
	PullPolicy corev1.PullPolicy `json:"pullPolicy,omitempty"`
}

type ContainerTemplate struct {
	ImageSpec        `json:",inline"`
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	// SecurityContext -
	SecurityContext *corev1.SecurityContext     `json:"securityContext,omitempty"`
	Resources       corev1.ResourceRequirements `json:"resources,omitempty"`
	VolumeMounts    []corev1.VolumeMount        `json:"volumeMounts,omitempty"`
	AdditionalEnv   []corev1.EnvVar             `json:"additionalEnv,omitempty"`
}

type WorkloadTemplate struct {
	Replicas         *int32                     `json:"replicas,omitempty"`
	SecurityContext  *corev1.PodSecurityContext `json:"securityContext"`
	AdditionalLabels map[string]string          `json:"additionalLabels,omitempty"`
	// Workload - customize the container template of the workload
	Workload *ContainerTemplate `json:"workload,omitempty"`
}

func (t *WorkloadTemplate) ReplicaCount() *int32 {
	if t != nil && t.Replicas != nil {
		return t.Replicas
	}

	return nil
}

func (t *WorkloadTemplate) MergeEnv(basicEnv []corev1.EnvVar) []corev1.EnvVar {
	if t == nil || t.Workload == nil || len(t.Workload.AdditionalEnv) == 0 {
		return basicEnv
	}

	existingKeys := make(map[string]bool, len(basicEnv)+len(t.Workload.AdditionalEnv))

	merged := append(make([]corev1.EnvVar, 0, len(basicEnv)+len(t.Workload.AdditionalEnv)), basicEnv...)

	for _, v := range basicEnv {
		existingKeys[v.Name] = true
	}

	for _, v := range t.Workload.AdditionalEnv {
		if _, alreadyPresent := existingKeys[v.Name]; alreadyPresent {
			continue
		}
		merged = append(merged, v)
		existingKeys[v.Name] = true
	}

	return merged
}

func (t *WorkloadTemplate) MergeLabels(initial map[string]string, toAppend ...map[string]string) map[string]string {
	result := make(map[string]string)

	maps.Copy(result, initial)

	var labelSets []map[string]string

	if t != nil && len(t.AdditionalLabels) > 0 {
		labelSets = append(labelSets, t.AdditionalLabels)
	}

	labelSets = append(labelSets, toAppend...)

	for _, lbls := range labelSets {
		for k, v := range lbls {
			if _, ok := result[k]; !ok {
				result[k] = v
			}
		}
	}

	return result
}

func (t *WorkloadTemplate) Image(defaultImage string) string {
	if t != nil && t.Workload != nil && t.Workload.Image != "" {
		return t.Workload.Image
	}

	return defaultImage
}

func (t *WorkloadTemplate) ImagePullPolicy() corev1.PullPolicy {
	if t != nil && t.Workload != nil && t.Workload.PullPolicy != "" {
		return t.Workload.PullPolicy
	}

	return corev1.PullIfNotPresent
}

func (t *WorkloadTemplate) PullSecrets() []corev1.LocalObjectReference {
	if t != nil && t.Workload != nil && len(t.Workload.ImagePullSecrets) > 0 {
		return t.Workload.ImagePullSecrets
	}

	return nil
}

func (t *WorkloadTemplate) Resources() corev1.ResourceRequirements {
	if t != nil && t.Workload != nil {
		return t.Workload.Resources
	}

	return corev1.ResourceRequirements{}
}

func (t *WorkloadTemplate) AdditionalVolumeMounts(defaultMounts ...corev1.VolumeMount) []corev1.VolumeMount {
	if t != nil && t.Workload != nil {
		return append(defaultMounts, t.Workload.VolumeMounts...)
	}

	return defaultMounts
}

func (t *WorkloadTemplate) PodSecurityContext() *corev1.PodSecurityContext {
	if t != nil && t.SecurityContext != nil {
		return t.SecurityContext
	}

	return &corev1.PodSecurityContext{
		RunAsNonRoot: ptrOf(true),
	}
}

func (t *WorkloadTemplate) ContainerSecurityContext(uid, gid int64) *corev1.SecurityContext {
	if t != nil && t.Workload != nil && t.Workload.SecurityContext != nil {
		return t.Workload.SecurityContext
	}

	return &corev1.SecurityContext{
		Privileged:               ptrOf(false),
		RunAsUser:                ptrOf(uid),
		RunAsGroup:               ptrOf(gid),
		RunAsNonRoot:             ptrOf(true),
		AllowPrivilegeEscalation: ptrOf(false),
		ReadOnlyRootFilesystem:   ptrOf(true),
		Capabilities: &corev1.Capabilities{
			Drop: []corev1.Capability{
				"ALL",
			},
		},
	}
}

func ptrOf[T any](val T) *T {
	return &val
}
