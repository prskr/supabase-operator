package supabase

import (
	"fmt"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type stringEnv string

func (e stringEnv) Var(value string) corev1.EnvVar {
	return corev1.EnvVar{
		Name:  string(e),
		Value: value,
	}
}

type stringSliceEnv struct {
	key       string
	separator string
}

func (e stringSliceEnv) Var(value []string) corev1.EnvVar {
	return corev1.EnvVar{
		Name:  e.key,
		Value: strings.Join(value, e.separator),
	}
}

type intEnv string

func (e intEnv) Var(value int) corev1.EnvVar {
	return corev1.EnvVar{
		Name:  string(e),
		Value: strconv.Itoa(value),
	}
}

type boolEnv string

func (e boolEnv) Var(value bool) corev1.EnvVar {
	return corev1.EnvVar{
		Name:  string(e),
		Value: strconv.FormatBool(value),
	}
}

type secretEnv string

func (e secretEnv) Var(sel *corev1.SecretKeySelector) corev1.EnvVar {
	return corev1.EnvVar{
		Name: string(e),
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: sel,
		},
	}
}

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
	DBUri                string
	Schemas              stringSliceEnv
	AnonRole             stringEnv
	JWTSecret            secretEnv
	UseLegacyGucs        boolEnv
	ExtraSearchPath      stringSliceEnv
	AppSettingsJWTSecret secretEnv
	AppSettingsJWTExpiry intEnv
	AdminServerPort      intEnv
	MaxRows              intEnv
}

type postgrestConfigDefaults struct {
	AnonRole        string
	Schemas         []string
	ExtraSearchPath []string
}

type authEnvKeys struct {
	ApiHost                    stringEnv
	ApiPort                    intEnv
	ApiExternalUrl             stringEnv
	DBDriver                   stringEnv
	DatabaseUrl                string
	SiteUrl                    stringEnv
	AdditionalRedirectURLs     stringSliceEnv
	DisableSignup              boolEnv
	JWTIssuer                  stringEnv
	JWTAdminRoles              stringEnv
	JWTAudience                stringEnv
	JwtDefaultGroup            stringEnv
	JwtExpiry                  intEnv
	JwtSecret                  secretEnv
	EmailSignupDisabled        boolEnv
	MailerUrlPathsInvite       stringEnv
	MailerUrlPathsConfirmation stringEnv
	MailerUrlPathsRecovery     stringEnv
	MailerUrlPathsEmailChange  stringEnv
	AnonymousUsersEnabled      boolEnv
}

type authConfigDefaults struct {
	ApiHost                    string
	ApiPort                    int
	DbDriver                   string
	JwtIssuer                  string
	JwtAdminRoles              string
	JwtAudience                string
	JwtDefaultGroupName        string
	MailerUrlPathsInvite       string
	MailerUrlPathsConfirmation string
	MailerUrlPathsRecovery     string
	MailerUrlPathsEmailChange  string
}

type pgMetaEnvKeys struct {
	APIPort    intEnv
	DBHost     stringEnv
	DBPort     intEnv
	DBName     stringEnv
	DBUser     secretEnv
	DBPassword secretEnv
}

type pgMetaDefaults struct {
	APIPort int
	DBPort  string
}

type envoyDefaults struct {
	ConfigKey string
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
	Envoy     envoyServiceConfig
	JWT       jwtConfig
}{
	Postgrest: serviceConfig[postgrestEnvKeys, postgrestConfigDefaults]{
		Name: "postgrest",
		EnvKeys: postgrestEnvKeys{
			DBUri:                "PGRST_DB_URI",
			Schemas:              stringSliceEnv{key: "PGRST_DB_SCHEMAS", separator: ","},
			AnonRole:             "PGRST_DB_ANON_ROLE",
			JWTSecret:            "PGRST_JWT_SECRET",
			UseLegacyGucs:        "PGRST_DB_USE_LEGACY_GUCS",
			AppSettingsJWTSecret: "PGRST_APP_SETTINGS_JWT_SECRET",
			AppSettingsJWTExpiry: "PGRST_APP_SETTINGS_JWT_EXP",
			AdminServerPort:      "PGRST_ADMIN_SERVER_PORT",
			ExtraSearchPath:      stringSliceEnv{key: "PGRST_DB_EXTRA_SEARCH_PATH", separator: ","},
		},
		Defaults: postgrestConfigDefaults{
			AnonRole:        "anon",
			Schemas:         []string{"public", "graphql_public"},
			ExtraSearchPath: []string{"public", "extensions"},
		},
	},
	Auth: serviceConfig[authEnvKeys, authConfigDefaults]{
		Name: "auth",
		EnvKeys: authEnvKeys{
			ApiHost:                    "GOTRUE_API_HOST",
			ApiPort:                    "GOTRUE_API_PORT",
			ApiExternalUrl:             "API_EXTERNAL_URL",
			DBDriver:                   "GOTRUE_DB_DRIVER",
			DatabaseUrl:                "GOTRUE_DB_DATABASE_URL",
			SiteUrl:                    "GOTRUE_SITE_URL",
			AdditionalRedirectURLs:     stringSliceEnv{key: "GOTRUE_URI_ALLOW_LIST", separator: ","},
			DisableSignup:              "GOTRUE_DISABLE_SIGNUP",
			JWTIssuer:                  "GOTRUE_JWT_ISSUER",
			JWTAdminRoles:              "GOTRUE_JWT_ADMIN_ROLES",
			JWTAudience:                "GOTRUE_JWT_AUD",
			JwtDefaultGroup:            "GOTRUE_JWT_DEFAULT_GROUP_NAME",
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
			ApiHost:                    "0.0.0.0",
			ApiPort:                    9999,
			DbDriver:                   "postgres",
			JwtIssuer:                  "supabase",
			JwtAdminRoles:              "service_role",
			JwtAudience:                "authenticated",
			JwtDefaultGroupName:        "authenticated",
			MailerUrlPathsInvite:       "/auth/v1/verify",
			MailerUrlPathsConfirmation: "/auth/v1/verify",
			MailerUrlPathsRecovery:     "/auth/v1/verify",
			MailerUrlPathsEmailChange:  "/auth/v1/verify",
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
		},
	},
	Envoy: envoyServiceConfig{
		Defaults: envoyDefaults{
			"config.yaml",
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
