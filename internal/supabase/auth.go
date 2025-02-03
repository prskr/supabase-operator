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

func authServiceConfig() serviceConfig[authEnvKeys, authConfigDefaults] {
	return serviceConfig[authEnvKeys, authConfigDefaults]{
		Name:              "auth",
		LivenessProbePath: "/health",
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
	}
}
