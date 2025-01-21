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

package supabase

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type serviceConfig[TEnvKeys, TDefaults any] struct {
	Name     string
	EnvKeys  TEnvKeys
	Defaults TDefaults
}

func (cfg serviceConfig[TEnvKeys, TDefaults]) ObjectName(obj metav1.Object) string {
	return fmt.Sprintf("%s-%s", obj.GetName(), cfg.Name)
}

func (cfg serviceConfig[TEnvKeys, TDefaults]) ObjectMeta(obj metav1.Object) metav1.ObjectMeta {
	return metav1.ObjectMeta{Name: cfg.ObjectName(obj), Namespace: obj.GetNamespace()}
}

type postgrestEnvKeys struct {
	Host                 fixedEnv
	DBUri                string
	Schemas              stringSliceEnv
	AnonRole             stringEnv
	JWTSecret            secretEnv
	UseLegacyGucs        boolEnv
	ExtraSearchPath      stringSliceEnv
	AppSettingsJWTSecret secretEnv
	AppSettingsJWTExpiry intEnv[int]
	AdminServerPort      intEnv[int32]
	MaxRows              intEnv[int]
	OpenAPIProxyURI      stringEnv
}

type postgrestConfigDefaults struct {
	AnonRole              string
	Schemas               []string
	ExtraSearchPath       []string
	UID, GID              int64
	ServerPort, AdminPort int32
}

type authEnvKeys struct {
	ApiHost                    fixedEnv
	ApiPort                    fixedEnv
	ApiExternalUrl             stringEnv
	DBDriver                   fixedEnv
	DatabaseUrl                string
	SiteUrl                    stringEnv
	AdditionalRedirectURLs     stringSliceEnv
	DisableSignup              boolEnv
	JWTIssuer                  fixedEnv
	JWTAdminRoles              fixedEnv
	JWTAudience                fixedEnv
	JwtDefaultGroup            fixedEnv
	JwtExpiry                  intEnv[int]
	JwtSecret                  secretEnv
	EmailSignupDisabled        boolEnv
	MailerUrlPathsInvite       stringEnv
	MailerUrlPathsConfirmation stringEnv
	MailerUrlPathsRecovery     stringEnv
	MailerUrlPathsEmailChange  stringEnv
	AnonymousUsersEnabled      boolEnv
}

type authConfigDefaults struct {
	MailerUrlPathsInvite       string
	MailerUrlPathsConfirmation string
	MailerUrlPathsRecovery     string
	MailerUrlPathsEmailChange  string
	APIPort                    int32
	UID, GID                   int64
}

type pgMetaEnvKeys struct {
	APIPort    intEnv[int32]
	DBHost     stringEnv
	DBPort     intEnv[int]
	DBName     stringEnv
	DBUser     secretEnv
	DBPassword secretEnv
}

type pgMetaDefaults struct {
	APIPort int32
	DBPort  string
	NodeUID int64
	NodeGID int64
}

type studioEnvKeys struct {
	PGMetaURL      stringEnv
	DBPassword     secretEnv
	ApiUrl         stringEnv
	APIExternalURL stringEnv
	JwtSecret      secretEnv
	AnonKey        secretEnv
	ServiceKey     secretEnv
	Host           fixedEnv
	LogsEnabled    fixedEnv
}

type studioDefaults struct {
	NodeUID int64
	NodeGID int64
	APIPort int32
}

type storageEnvApiKeys struct {
	AnonKey                     secretEnv
	ServiceKey                  secretEnv
	JwtSecret                   secretEnv
	JwtJwks                     secretEnv
	DatabaseDSN                 stringEnv
	FileSizeLimit               intEnv[uint64]
	UploadFileSizeLimit         intEnv[uint64]
	UploadFileSizeLimitStandard intEnv[uint64]
	StorageBackend              stringEnv
	TenantID                    fixedEnv
	StorageS3Region             stringEnv
	GlobalS3Bucket              fixedEnv
	EnableImaageTransformation  boolEnv
	ImgProxyURL                 stringEnv
	TusUrlPath                  fixedEnv
	S3AccessKeyId               secretEnv
	S3AccessKeySecret           secretEnv
	S3ProtocolPrefix            fixedEnv
	S3AllowForwardedHeader      boolEnv
}

type storageApiDefaults struct{}

type envoyDefaults struct {
	ConfigKey string
	UID, GID  int64
}

type envoyServiceConfig struct {
	Defaults envoyDefaults
}

func (envoyServiceConfig) ObjectName(obj metav1.Object) string {
	return fmt.Sprintf("%s-envoy", obj.GetName())
}

type jwtDefaults struct {
	SecretKey    string
	JwksKey      string
	AnonKey      string
	ServiceKey   string
	SecretLength int
	Expiry       int
}

type jwtConfig struct {
	Defaults jwtDefaults
}

func (jwtConfig) ObjectName(obj metav1.Object) string {
	return fmt.Sprintf("%s-jwt", obj.GetName())
}

var ServiceConfig = struct {
	Postgrest serviceConfig[postgrestEnvKeys, postgrestConfigDefaults]
	Auth      serviceConfig[authEnvKeys, authConfigDefaults]
	PGMeta    serviceConfig[pgMetaEnvKeys, pgMetaDefaults]
	Studio    serviceConfig[studioEnvKeys, studioDefaults]
	Storage   serviceConfig[storageEnvApiKeys, storageApiDefaults]
	Envoy     envoyServiceConfig
	JWT       jwtConfig
}{
	Postgrest: serviceConfig[postgrestEnvKeys, postgrestConfigDefaults]{
		Name: "postgrest",
		EnvKeys: postgrestEnvKeys{
			Host:                 fixedEnvOf("PGRST_SERVER_HOST", "*"),
			DBUri:                "PGRST_DB_URI",
			Schemas:              stringSliceEnv{key: "PGRST_DB_SCHEMAS", separator: ","},
			AnonRole:             "PGRST_DB_ANON_ROLE",
			JWTSecret:            "PGRST_JWT_SECRET",
			UseLegacyGucs:        "PGRST_DB_USE_LEGACY_GUCS",
			AppSettingsJWTSecret: "PGRST_APP_SETTINGS_JWT_SECRET",
			AppSettingsJWTExpiry: "PGRST_APP_SETTINGS_JWT_EXP",
			AdminServerPort:      "PGRST_ADMIN_SERVER_PORT",
			ExtraSearchPath:      stringSliceEnv{key: "PGRST_DB_EXTRA_SEARCH_PATH", separator: ","},
			MaxRows:              "PGRST_DB_MAX_ROWS",
			OpenAPIProxyURI:      "PGRST_OPENAPI_SERVER_PROXY_URI",
		},
		Defaults: postgrestConfigDefaults{
			AnonRole:        "anon",
			Schemas:         []string{"public", "graphql_public"},
			ExtraSearchPath: []string{"public", "extensions"},
			UID:             1000,
			GID:             1000,
			ServerPort:      3000,
			AdminPort:       3001,
		},
	},
	Auth: serviceConfig[authEnvKeys, authConfigDefaults]{
		Name: "auth",
		EnvKeys: authEnvKeys{
			ApiHost:                    fixedEnvOf("GOTRUE_API_HOST", "0.0.0.0"),
			ApiPort:                    fixedEnvOf("GOTRUE_API_PORT", "9999"),
			ApiExternalUrl:             "API_EXTERNAL_URL",
			DBDriver:                   fixedEnvOf("GOTRUE_DB_DRIVER", "postgres"),
			DatabaseUrl:                "GOTRUE_DB_DATABASE_URL",
			SiteUrl:                    "GOTRUE_SITE_URL",
			AdditionalRedirectURLs:     stringSliceEnv{key: "GOTRUE_URI_ALLOW_LIST", separator: ","},
			DisableSignup:              "GOTRUE_DISABLE_SIGNUP",
			JWTIssuer:                  fixedEnvOf("GOTRUE_JWT_ISSUER", "supabase"),
			JWTAdminRoles:              fixedEnvOf("GOTRUE_JWT_ADMIN_ROLES", "service_role"),
			JWTAudience:                fixedEnvOf("GOTRUE_JWT_AUD", "authenticated"),
			JwtDefaultGroup:            fixedEnvOf("GOTRUE_JWT_DEFAULT_GROUP_NAME", "authenticated"),
			JwtExpiry:                  "GOTRUE_JWT_EXP",
			JwtSecret:                  "GOTRUE_JWT_SECRET",
			EmailSignupDisabled:        "GOTRUE_EXTERNAL_EMAIL_ENABLED",
			MailerUrlPathsInvite:       "MAILER_URLPATHS_INVITE",
			MailerUrlPathsConfirmation: "MAILER_URLPATHS_CONFIRMATION",
			MailerUrlPathsRecovery:     "MAILER_URLPATHS_RECOVERY",
			MailerUrlPathsEmailChange:  "MAILER_URLPATHS_EMAIL_CHANGE",
			AnonymousUsersEnabled:      "GOTRUE_EXTERNAL_ANONYMOUS_USERS_ENABLED",
		},
		Defaults: authConfigDefaults{
			MailerUrlPathsInvite:       "/auth/v1/verify",
			MailerUrlPathsConfirmation: "/auth/v1/verify",
			MailerUrlPathsRecovery:     "/auth/v1/verify",
			MailerUrlPathsEmailChange:  "/auth/v1/verify",
			APIPort:                    9999,
			UID:                        1000,
			GID:                        1000,
		},
	},
	PGMeta: serviceConfig[pgMetaEnvKeys, pgMetaDefaults]{
		Name: "pg-meta",
		EnvKeys: pgMetaEnvKeys{
			APIPort:    "PG_META_PORT",
			DBHost:     "PG_META_DB_HOST",
			DBPort:     "PG_META_DB_PORT",
			DBName:     "PG_META_DB_NAME",
			DBUser:     "PG_META_DB_USER",
			DBPassword: "PG_META_DB_PASSWORD",
		},
		Defaults: pgMetaDefaults{
			APIPort: 8080,
			DBPort:  "5432",
			NodeUID: 1000,
			NodeGID: 1000,
		},
	},
	Studio: serviceConfig[studioEnvKeys, studioDefaults]{
		Name: "studio",
		EnvKeys: studioEnvKeys{
			PGMetaURL:      "STUDIO_PG_META_URL",
			DBPassword:     "POSTGRES_PASSWORD",
			ApiUrl:         "SUPABASE_URL",
			APIExternalURL: "SUPABASE_PUBLIC_URL",
			JwtSecret:      "AUTH_JWT_SECRET",
			AnonKey:        "SUPABASE_ANON_KEY",
			ServiceKey:     "SUPABASE_SERVICE_KEY",
			Host:           fixedEnvOf("HOSTNAME", "0.0.0.0"),
			LogsEnabled:    fixedEnvOf("NEXT_PUBLIC_ENABLE_LOGS", "true"),
		},
		Defaults: studioDefaults{
			NodeUID: 1000,
			NodeGID: 1000,
			APIPort: 3000,
		},
	},
	Storage: serviceConfig[storageEnvApiKeys, storageApiDefaults]{
		Name: "storage-api",
		EnvKeys: storageEnvApiKeys{
			AnonKey:                     "ANON_KEY",
			ServiceKey:                  "SERVICE_KEY",
			JwtSecret:                   "AUTH_JWT_SECRET",
			JwtJwks:                     "AUTH_JWT_JWKS",
			StorageBackend:              "STORAGE_BACKEND",
			DatabaseDSN:                 "DATABASE_URL",
			FileSizeLimit:               "FILE_SIZE_LIMIT",
			UploadFileSizeLimit:         "UPLOAD_FILE_SIZE_LIMIT",
			UploadFileSizeLimitStandard: "UPLOAD_FILE_SIZE_LIMIT_STANDARD",
			TenantID:                    fixedEnvOf("TENANT_ID", "stub"),
			StorageS3Region:             "STORAGE_S3_REGION",
			GlobalS3Bucket:              fixedEnvOf("GLOBAL_S3_BUCKET", "stub"),
			EnableImaageTransformation:  "ENABLE_IMAGE_TRANSFORMATION",
			ImgProxyURL:                 "IMGPROXY_URL",
			TusUrlPath:                  fixedEnvOf("TUS_URL_PATH", "/storage/v1/upload/resumable"),
			S3AccessKeyId:               "S3_PROTOCOL_ACCESS_KEY_ID",
			S3AccessKeySecret:           "S3_PROTOCOL_ACCESS_KEY_SECRET",
			S3ProtocolPrefix:            fixedEnvOf("S3_PROTOCOL_PREFIX", "/storage/v1"),
			S3AllowForwardedHeader:      "S3_ALLOW_FORWARDED_HEADER",
		},
		Defaults: storageApiDefaults{},
	},
	Envoy: envoyServiceConfig{
		Defaults: envoyDefaults{
			ConfigKey: "config.yaml",
			UID:       65532,
			GID:       65532,
		},
	},
	JWT: jwtConfig{
		Defaults: jwtDefaults{
			SecretKey:    "secret",
			JwksKey:      "jwks.json",
			AnonKey:      "anon_key",
			ServiceKey:   "service_key",
			SecretLength: 40,
			Expiry:       3600,
		},
	},
}
