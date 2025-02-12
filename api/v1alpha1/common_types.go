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

type JwtSpec struct {
	// SecretRef - object reference to the Secret where JWT values are stored
	SecretName string `json:"secretName,omitempty"`
	// SecretKey - key in secret where to read the JWT HMAC secret from
	// +kubebuilder:default=secret
	SecretKey string `json:"secretKey,omitempty"`
	// JwksKey - key in secret where to read the JWKS from
	// +kubebuilder:default=jwks.json
	JwksKey string `json:"jwksKey,omitempty"`
	// AnonKey - key in secret where to read the anon JWT from
	// +kubebuilder:default=anon_key
	AnonKey string `json:"anonKey,omitempty"`
	// ServiceKey - key in secret where to read the service JWT from
	// +kubebuilder:default=service_key
	ServiceKey string `json:"serviceKey,omitempty"`
}

func (s JwtSpec) SecretKeySelector() *corev1.SecretKeySelector {
	return &corev1.SecretKeySelector{
		LocalObjectReference: corev1.LocalObjectReference{
			Name: s.SecretName,
		},
		Key: s.SecretKey,
	}
}

func (s JwtSpec) JwksKeySelector() *corev1.SecretKeySelector {
	return &corev1.SecretKeySelector{
		LocalObjectReference: corev1.LocalObjectReference{
			Name: s.SecretName,
		},
		Key: s.JwksKey,
	}
}

func (s JwtSpec) AnonKeySelector() *corev1.SecretKeySelector {
	return &corev1.SecretKeySelector{
		LocalObjectReference: corev1.LocalObjectReference{
			Name: s.SecretName,
		},
		Key: s.AnonKey,
	}
}

func (s JwtSpec) ServiceKeySelector() *corev1.SecretKeySelector {
	return &corev1.SecretKeySelector{
		LocalObjectReference: corev1.LocalObjectReference{
			Name: s.SecretName,
		},
		Key: s.ServiceKey,
	}
}

type ImageSpec struct {
	Image      string            `json:"image,omitempty"`
	PullPolicy corev1.PullPolicy `json:"pullPolicy,omitempty"`
}

type ContainerTemplate struct {
	ImageSpec        `json:",inline"`
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	// SecurityContext - override the container SecurityContext
	// use with caution, by default the operator already uses sane defaults
	SecurityContext *corev1.SecurityContext     `json:"securityContext,omitempty"`
	Resources       corev1.ResourceRequirements `json:"resources,omitempty"`
	VolumeMounts    []corev1.VolumeMount        `json:"volumeMounts,omitempty"`
	AdditionalEnv   []corev1.EnvVar             `json:"additionalEnv,omitempty"`
}

type WorkloadSpec struct {
	Replicas         *int32                     `json:"replicas,omitempty"`
	SecurityContext  *corev1.PodSecurityContext `json:"securityContext,omitempty"`
	AdditionalLabels map[string]string          `json:"additionalLabels,omitempty"`
	// ContainerSpec - customize the container template of the workload
	ContainerSpec     *ContainerTemplate `json:"container,omitempty"`
	AdditionalVolumes []corev1.Volume    `json:"additionalVolumes,omitempty"`
}

func (t *WorkloadSpec) ReplicaCount() *int32 {
	if t != nil && t.Replicas != nil {
		return t.Replicas
	}

	return nil
}

func (t *WorkloadSpec) MergeEnv(basicEnv []corev1.EnvVar) []corev1.EnvVar {
	if t == nil || t.ContainerSpec == nil || len(t.ContainerSpec.AdditionalEnv) == 0 {
		return basicEnv
	}

	existingKeys := make(map[string]bool, len(basicEnv)+len(t.ContainerSpec.AdditionalEnv))

	merged := append(make([]corev1.EnvVar, 0, len(basicEnv)+len(t.ContainerSpec.AdditionalEnv)), basicEnv...)

	for _, v := range basicEnv {
		existingKeys[v.Name] = true
	}

	for _, v := range t.ContainerSpec.AdditionalEnv {
		if _, alreadyPresent := existingKeys[v.Name]; alreadyPresent {
			continue
		}
		merged = append(merged, v)
		existingKeys[v.Name] = true
	}

	return merged
}

func (t *WorkloadSpec) MergeLabels(initial map[string]string, toAppend ...map[string]string) map[string]string {
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

func (t *WorkloadSpec) Image(defaultImage string) string {
	if t != nil && t.ContainerSpec != nil && t.ContainerSpec.Image != "" {
		return t.ContainerSpec.Image
	}

	return defaultImage
}

func (t *WorkloadSpec) ImagePullPolicy() corev1.PullPolicy {
	if t != nil && t.ContainerSpec != nil && t.ContainerSpec.PullPolicy != "" {
		return t.ContainerSpec.PullPolicy
	}

	return corev1.PullIfNotPresent
}

func (t *WorkloadSpec) PullSecrets() []corev1.LocalObjectReference {
	if t != nil && t.ContainerSpec != nil && len(t.ContainerSpec.ImagePullSecrets) > 0 {
		return t.ContainerSpec.ImagePullSecrets
	}

	return nil
}

func (t *WorkloadSpec) Resources() corev1.ResourceRequirements {
	if t != nil && t.ContainerSpec != nil {
		return t.ContainerSpec.Resources
	}

	return corev1.ResourceRequirements{}
}

func (t *WorkloadSpec) AdditionalVolumeMounts(defaultMounts ...corev1.VolumeMount) []corev1.VolumeMount {
	if t != nil && t.ContainerSpec != nil {
		return append(defaultMounts, t.ContainerSpec.VolumeMounts...)
	}

	return defaultMounts
}

func (t *WorkloadSpec) Volumes(defaultVolumes ...corev1.Volume) []corev1.Volume {
	if t == nil {
		return defaultVolumes
	}

	return append(defaultVolumes, t.AdditionalVolumes...)
}

func (t *WorkloadSpec) PodSecurityContext() *corev1.PodSecurityContext {
	if t != nil && t.SecurityContext != nil {
		return t.SecurityContext
	}

	return &corev1.PodSecurityContext{
		RunAsNonRoot: ptrOf(true),
	}
}

func (t *WorkloadSpec) ContainerSecurityContext(uid, gid int64) *corev1.SecurityContext {
	if t != nil && t.ContainerSpec != nil && t.ContainerSpec.SecurityContext != nil {
		return t.ContainerSpec.SecurityContext
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
