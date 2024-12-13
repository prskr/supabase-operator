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
	"errors"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Database struct {
	DSN     *string                   `json:"dsn,omitempty"`
	DSNFrom *corev1.SecretKeySelector `json:"dsnFrom,omitempty"`
}

func (d Database) GetDSN(ctx context.Context, client client.Client) (string, error) {
	if d.DSN != nil {
		return *d.DSN, nil
	}

	if d.DSNFrom == nil {
		return "", errors.New("DSN not set")
	}

	var secret corev1.Secret
	if err := client.Get(ctx, types.NamespacedName{Name: d.DSNFrom.Name}, &secret); err != nil {
		return "", err
	}

	data, ok := secret.Data[d.DSNFrom.Key]
	if !ok {
		return "", errors.New("key not found in secret")
	}

	return string(data), nil
}

// CoreSpec defines the desired state of Core.
type CoreSpec struct {
	// Important: Run "make" to regenerate code after modifying this file

	Database Database `json:"database,omitempty"`
}

type MigrationStatus map[string]int64

func (s MigrationStatus) IsApplied(name string) bool {
	_, ok := s[name]
	return ok
}

func (s MigrationStatus) Record(name string) {
	s[name] = time.Now().UTC().UnixMilli()
}

// CoreStatus defines the observed state of Core.
type CoreStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	AppliedMigrations MigrationStatus `json:"appliedMigrations,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Core is the Schema for the cores API.
type Core struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CoreSpec   `json:"spec,omitempty"`
	Status CoreStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CoreList contains a list of Core.
type CoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Core `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Core{}, &CoreList{})
}
