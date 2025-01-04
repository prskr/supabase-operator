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
