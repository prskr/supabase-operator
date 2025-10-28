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
	"bytes"
	"context"
	"errors"
	"fmt"
	"iter"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"code.icb4dc0.de/prskr/supabase-operator/internal/supabase"
)

func init() {
	SchemeBuilder.Register(&Core{}, &CoreList{})
}

var (
	ErrNoSuchSecretValue = errors.New("no such secret value")
	ErrDSNNotSet         = errors.New("DSN not set")
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Core is the Schema for the cores API.
type Core struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CoreSpec   `json:"spec,omitempty"`
	Status CoreStatus `json:"status,omitempty"`
}

// CoreSpec defines the desired state of Core.
type CoreSpec struct {
	// APIExternalURL is referring to the URL where Supabase API will be available
	// Typically this is the ingress of the API gateway
	APIExternalURL string `json:"externalUrl"`
	// SiteURL is referring to the URL of the (frontend) application
	// In most Kubernetes scenarios this is the same as the APIExternalURL with a different path handler in the ingress
	SiteURL   string        `json:"siteUrl"`
	JWT       *CoreJwtSpec  `json:"jwt,omitempty"`
	Database  Database      `json:"database,omitzero"`
	Postgrest PostgrestSpec `json:"postgrest,omitzero"`
	Auth      *AuthSpec     `json:"auth,omitempty"`
}

type Database struct {
	DSN          *string                   `json:"dsn,omitempty"`
	DSNSecretRef *corev1.SecretKeySelector `json:"dsnSecretRef"`
	Roles        DatabaseRoles             `json:"roles,omitzero"`
}

func (d Database) GetDSN(ctx context.Context, client client.Client) (string, error) {
	if d.DSNSecretRef == nil {
		return "", ErrDSNNotSet
	}

	var secret corev1.Secret
	if err := client.Get(ctx, types.NamespacedName{Name: d.DSNSecretRef.Name}, &secret); err != nil {
		return "", err
	}

	data, ok := secret.Data[d.DSNSecretRef.Key]
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrNoSuchSecretValue, d.DSNSecretRef.Key)
	}

	return string(data), nil
}

func (d Database) DSNEnv(key string) corev1.EnvVar {
	return corev1.EnvVar{
		Name: key,
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: d.DSNSecretRef,
		},
	}
}

type DatabaseRoles struct {
	// SelfManaged - whether the database roles are managed externally
	// when enabled the operator does not attempt to create secrets, generate passwords or whatsoever for all database roles
	// i.e. all secrets need to be provided or the instance won't work
	SelfManaged bool `json:"selfManaged,omitempty"`
	// Secrets - typed 'map' of secrets for each database role that Supabase needs
	Secrets DatabaseRolesSecrets `json:"secrets,omitzero"`
}

type DatabaseRolesSecrets struct {
	Admin          string `json:"supabaseAdmin,omitempty"`
	Authenticator  string `json:"authenticator,omitempty"`
	AuthAdmin      string `json:"supabaseAuthAdmin,omitempty"`
	FunctionsAdmin string `json:"supabaseFunctionsAdmin,omitempty"`
	StorageAdmin   string `json:"supabaseStorageAdmin,omitempty"`
}

type CoreJwtSpec struct {
	JwtSpec `json:",inline"`
	// Secret - JWT HMAC secret in plain text
	// This is WRITE-ONLY and will be copied to the SecretRef by the defaulter
	Secret *string `json:"secret,omitempty"`
	// Expiry - expiration time in seconds for JWTs
	// +kubebuilder:default=3600
	Expiry int `json:"expiry,omitempty"`
}

func (s CoreJwtSpec) GetJWTSecret(ctx context.Context, client client.Client) ([]byte, error) {
	var secret corev1.Secret
	if err := client.Get(ctx, types.NamespacedName{Name: s.SecretName}, &secret); err != nil {
		return nil, nil
	}

	value, ok := secret.Data[s.SecretKey]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrNoSuchSecretValue, s.SecretKey)
	}

	return value, nil
}

func (s CoreJwtSpec) SecretKeySelector() *corev1.SecretKeySelector {
	return &corev1.SecretKeySelector{
		LocalObjectReference: corev1.LocalObjectReference{
			Name: s.SecretName,
		},
		Key: s.SecretKey,
	}
}

func (s CoreJwtSpec) JwksKeySelector() *corev1.SecretKeySelector {
	return &corev1.SecretKeySelector{
		LocalObjectReference: corev1.LocalObjectReference{
			Name: s.SecretName,
		},
		Key: s.JwksKey,
	}
}

func (s CoreJwtSpec) SecretAsEnv(key string) corev1.EnvVar {
	return corev1.EnvVar{
		Name: key,
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: s.SecretName,
				},
				Key: s.SecretKey,
			},
		},
	}
}

func (s CoreJwtSpec) ExpiryAsEnv(key string) corev1.EnvVar {
	return corev1.EnvVar{
		Name:  key,
		Value: strconv.Itoa(s.Expiry),
	}
}

type PostgrestSpec struct {
	// Schemas - schema where PostgREST is looking for objects (tables, views, functions, ...)
	// +kubebuilder:default={"public","graphql_public"}
	Schemas []string `json:"schemas,omitempty"`
	// ExtraSearchPath - Extra schemas to add to the search_path of every request.
	// These schemas tables, views and functions don’t get API endpoints, they can only be referred from the database objects inside your db-schemas.
	// +kubebuilder:default={"public","extensions"}
	ExtraSearchPath []string `json:"extraSearchPath,omitempty"`
	// AnonRole - name of the anon role
	// +kubebuilder:default=anon
	AnonRole string `json:"anonRole,omitempty"`
	// MaxRows - maximum number of rows PostgREST will load at a time
	// +kubebuilder:default=1000
	MaxRows int `json:"maxRows,omitempty"`
	// WorkloadSpec - customize the PostgREST workload
	WorkloadSpec *WorkloadSpec `json:"workloadSpec,omitempty"`
	// ObservabilitySpec - customize the PostgREST observability
	Observability *PostgrestObservabilitySpec `json:"observability,omitempty"`
}

type PostgrestObservabilitySpec struct {
	Metrics *PostgrestMetricsSpec `json:"metrics,omitempty"`
}

type PostgrestMetricsSpec struct {
	Enabled bool `json:"enabled,omitempty"`
}

type AuthSpec struct {
	AdditionalRedirectUrls []string                 `json:"additionalRedirectUrls,omitempty"`
	DisableSignup          *bool                    `json:"disableSignup,omitempty"`
	AnonymousUsersEnabled  *bool                    `json:"anonymousUsersEnabled,omitempty"`
	Providers              *AuthProviders           `json:"providers,omitempty"`
	WorkloadTemplate       *WorkloadSpec            `json:"workloadTemplate,omitempty"`
	EmailSignupDisabled    *bool                    `json:"emailSignupDisabled,omitempty"`
	Observability          *GoTrueObservabilitySpec `json:"observability,omitempty"`
}

type AuthProviders struct {
	Email  *EmailAuthProvider  `json:"email,omitempty"`
	Azure  *AzureAuthProvider  `json:"azure,omitempty"`
	Github *GithubAuthProvider `json:"github,omitempty"`
	Phone  *PhoneAuthProvider  `json:"phone,omitempty"`
}

func (p *AuthProviders) Vars(apiExternalURL string) []corev1.EnvVar {
	if p == nil {
		return nil
	}

	return slices.Concat(
		p.Email.Vars(apiExternalURL),
		p.Azure.Vars(apiExternalURL),
		p.Github.Vars(apiExternalURL),
		p.Phone.Vars(),
	)
}

type AuthProviderMeta struct {
	// Enabled - whether the authentication provider is enabled or not
	Enabled bool `json:"enabled,omitempty"`
}

func (p *AuthProviderMeta) Vars(provider string) []corev1.EnvVar {
	if p == nil {
		return nil
	}

	return []corev1.EnvVar{{
		Name:  fmt.Sprintf("GOTRUE_EXTERNAL_%s_ENABLED", strings.ToUpper(provider)),
		Value: strconv.FormatBool(p.Enabled),
	}}
}

type GithubAuthProvider struct {
	AuthProviderMeta `json:",inline"`
	OAuthProvider    `json:",inline"`
}

func (p *GithubAuthProvider) Vars(apiExternalURL string) []corev1.EnvVar {
	const providerName = "GITHUB"
	if p == nil {
		return nil
	}

	return slices.Concat(
		p.AuthProviderMeta.Vars(providerName),
		p.OAuthProvider.Vars(providerName, apiExternalURL),
	)
}

type AzureAuthProvider struct {
	AuthProviderMeta `json:",inline"`
	OAuthProvider    `json:",inline"`
}

func (p *AzureAuthProvider) Vars(apiExternalURL string) []corev1.EnvVar {
	const providerName = "AZURE"
	if p == nil {
		return nil
	}

	return slices.Concat(
		p.AuthProviderMeta.Vars(providerName),
		p.OAuthProvider.Vars(providerName, apiExternalURL),
	)
}

type OAuthProvider struct {
	ClientID        string                    `json:"clientID"`
	ClientSecretRef *corev1.SecretKeySelector `json:"clientSecretRef"`
	URL             string                    `json:"url,omitempty"`
}

func (p *OAuthProvider) Vars(provider, apiExternalURL string) []corev1.EnvVar {
	if p == nil {
		return nil
	}

	vars := []corev1.EnvVar{
		{
			Name:  fmt.Sprintf("GOTRUE_EXTERNAL_%s_CLIENT_ID", strings.ToUpper(provider)),
			Value: p.ClientID,
		},
		{
			Name:  fmt.Sprintf("GOTRUE_EXTERNAL_%s_REDIRECT_URI", strings.ToUpper(provider)),
			Value: path.Join(apiExternalURL, "/auth/v1/callback"),
		},
		{
			Name: fmt.Sprintf("GOTRUE_EXTERNAL_%s_SECRET", strings.ToUpper(provider)),
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: p.ClientSecretRef,
			},
		},
	}

	if p.URL != "" {
		vars = append(vars, corev1.EnvVar{
			Name:  fmt.Sprintf("GOTRUE_EXTERNAL_%s_URL", strings.ToUpper(provider)),
			Value: p.URL,
		})
	}

	return vars
}

type PhoneAuthProvider struct {
	AuthProviderMeta `json:",inline"`
}

func (p *PhoneAuthProvider) Vars() []corev1.EnvVar {
	if p == nil {
		return nil
	}

	return []corev1.EnvVar{}
}

type EmailAuthProvider struct {
	AuthProviderMeta     `json:",inline"`
	AdminEmail           string             `json:"adminEmail"`
	SenderName           *string            `json:"senderName,omitempty"`
	Autoconfirm          *bool              `json:"autoconfirmEmail,omitempty"`
	SubjectsInvite       string             `json:"subjectsInvite,omitempty"`
	SubjectsConfirmation string             `json:"subjectsConfirmation,omitempty"`
	SMTPSpec             *EmailAuthSMTPSpec `json:"smtpSpec"`
}

func (p *EmailAuthProvider) Vars(apiExternalURL string) []corev1.EnvVar {
	if p == nil || p.SMTPSpec == nil {
		return nil
	}

	svcDefaults := supabase.ServiceConfig.Auth.Defaults

	vars := []corev1.EnvVar{
		{Name: "GOTRUE_SMTP_HOST", Value: p.SMTPSpec.Host},
		{Name: "GOTRUE_SMTP_PORT", Value: strconv.FormatUint(uint64(p.SMTPSpec.Port), 10)},
		{
			Name: "GOTRUE_SMTP_USER",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: p.SMTPSpec.CredentialsRef.SecretName,
					},
					Key: p.SMTPSpec.CredentialsRef.UsernameKey,
				},
			},
		},
		{
			Name: "GOTRUE_SMTP_PASS",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: p.SMTPSpec.CredentialsRef.SecretName,
					},
					Key: p.SMTPSpec.CredentialsRef.PasswordKey,
				},
			},
		},
		{Name: "GOTRUE_SMTP_ADMIN_EMAIL", Value: p.AdminEmail},
		{Name: "MAILER_URLPATHS_INVITE", Value: path.Join(apiExternalURL, svcDefaults.MailerUrlPathsInvite)},
		{Name: "MAILER_URLPATHS_CONFIRMATION", Value: path.Join(apiExternalURL, svcDefaults.MailerUrlPathsConfirmation)},
		{Name: "MAILER_URLPATHS_RECOVERY", Value: path.Join(apiExternalURL, svcDefaults.MailerUrlPathsRecovery)},
		{Name: "MAILER_URLPATHS_EMAIL_CHANGE", Value: path.Join(apiExternalURL, svcDefaults.MailerUrlPathsEmailChange)},
	}

	if p.SubjectsInvite != "" {
		vars = append(vars, corev1.EnvVar{Name: "MAILER_SUBJECTS_INVITE", Value: p.SubjectsInvite})
	}

	return vars
}

type EmailAuthSMTPSpec struct {
	Host           string                    `json:"host"`
	Port           uint16                    `json:"port"`
	MaxFrequency   *uint                     `json:"maxFrequency,omitempty"`
	CredentialsRef *SMTPCredentialsReference `json:"credentialsRef"`
}

type SMTPCredentialsReference struct {
	SecretName string `json:"secretName"`
	// UsernameKey
	// +kubebuilder:default="username"
	UsernameKey string `json:"usernameKey"`
	// PasswordKey
	// +kubebuilder:default="password"
	PasswordKey string `json:"passwordKey"`
}

type GoTrueObservabilitySpec struct {
	Metrics *GoTrueMetricsSpec `json:"metrics,omitempty"`
	Tracing *GoTrueTracingSpec `json:"tracing,omitempty"`
}

func (s *GoTrueObservabilitySpec) Vars() []corev1.EnvVar {
	if s == nil {
		return nil
	}

	var vars []corev1.EnvVar

	if s.Metrics != nil && s.Metrics.Enabled {
		vars = append(vars, corev1.EnvVar{
			Name:  "GOTRUE_METRICS_ENABLED",
			Value: strconv.FormatBool(true),
		},
		)
	}

	return vars
}

type GoTrueMetricsSpec struct {
	Enabled bool `json:"enabled"`
}

type GoTrueTracingSpec struct{}

type CoreConditionType string

// CoreStatus defines the observed state of Core.
type CoreStatus struct {
	Database DatabaseStatus `json:"database,omitempty"`
}

type DatabaseStatus struct {
	MigrationConditions []MigrationScriptCondition `json:"migrationConditions,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
	Roles               map[string][]byte          `json:"roles,omitempty"`
}

func (s DatabaseStatus) IsMigrationUpToDate(name string, hash []byte) (found bool, upToDate bool) {
	for _, cond := range s.MigrationConditions {
		if cond.Name == name {
			return true, bytes.Equal(cond.Hash, hash)
		}
	}

	return false, false
}

func (s *DatabaseStatus) RecordMigrationCondition(name string, hash []byte, err error) error {
	var (
		now                = time.Now()
		newStatus          = MigrationConditionStatusApplied
		lastProbeTime      = metav1.NewTime(now)
		lastTransitionTime = metav1.NewTime(now)
		message            string
	)

	if err != nil {
		newStatus = MigrationConditionStatusFailed
		message = err.Error()
	}

	for idx, cond := range s.MigrationConditions {
		if cond.Name == name {
			lastTransitionTime = cond.LastTransitionTime
			if cond.Status != newStatus {
				lastTransitionTime = metav1.NewTime(now)
			}

			cond.Hash = hash
			cond.Status = newStatus
			cond.LastProbeTime = lastProbeTime
			cond.LastTransitionTime = lastTransitionTime
			cond.Reason = "Outdated"
			cond.Message = message

			s.MigrationConditions[idx] = cond
			return err
		}
	}

	s.MigrationConditions = append(s.MigrationConditions, MigrationScriptCondition{
		Name:               name,
		Hash:               hash,
		Status:             newStatus,
		LastProbeTime:      lastProbeTime,
		LastTransitionTime: lastTransitionTime,
		Message:            message,
	})

	return err
}

type MigrationConditionStatus string

const (
	MigrationConditionStatusApplied MigrationConditionStatus = "Applied"
	MigrationConditionStatusFailed  MigrationConditionStatus = "Failed"
)

type MigrationScriptCondition struct {
	// Name - file name of the migration script
	Name string `json:"name"`
	// Hash - SHA256 hash of the script when it was last successfully applied
	Hash []byte `json:"hash"`
	// Status - whether the migration was applied or not
	// +kubebuilder:validation:Enum=Applied;Failed
	Status MigrationConditionStatus `json:"status"`
	// LastProbeTime - last time the operator tried to execute the migration script
	LastProbeTime metav1.Time `json:"lastProbeTime,omitzero"`
	// LastTransitionTime - last time the condition transitioned from one status to another
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitzero"`
	// Reason - one-word, CamcelCase reason for the condition's last transition
	Reason string `json:"reason,omitempty"`
	// Message - human-readable message indicating details about the last transition
	Message string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true

// CoreList contains a list of Core.
type CoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Core `json:"items"`
}

func (l CoreList) Iter() iter.Seq[*Core] {
	return func(yield func(*Core) bool) {
		for _, c := range l.Items {
			if !yield(&c) {
				return
			}
		}
	}
}
