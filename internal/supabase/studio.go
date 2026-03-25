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

type studioEnvKeys struct {
	PGMetaURL         stringEnv
	DBHost            stringEnv
	DBPort            intEnv[int]
	DBName            stringEnv
	DBPassword        secretEnv
	DBSchemas         stringSliceEnv
	DBMaxRows         intEnv[int]
	DBExtraSearchPath stringSliceEnv
	ApiUrl            stringEnv
	APIExternalURL    stringEnv
	JwtSecret         secretEnv
	AnonKey           secretEnv
	ServiceKey        secretEnv
	Host              fixedEnv
	LogsEnabled       fixedEnv
	MetaCryptoKey     secretEnv
}

type studioDefaults struct {
	NodeUID         int64
	NodeGID         int64
	APIPort         int32
	MaxRows         int
	Schemas         []string
	ExtraSearchPath []string
}

func studioServiceConfig() serviceConfig[studioEnvKeys, studioDefaults] {
	return serviceConfig[studioEnvKeys, studioDefaults]{
		Name:              "studio",
		LivenessProbePath: "/api/platform/profile",
		EnvKeys: studioEnvKeys{
			PGMetaURL:         "STUDIO_PG_META_URL",
			DBHost:            "POSTGRES_HOST",
			DBPort:            "POSTGRES_PORT",
			DBName:            "POSTGRES_DB",
			DBPassword:        "POSTGRES_PASSWORD",
			DBSchemas:         stringSliceEnv{key: "PGRST_DB_SCHEMAS", separator: ","},
			DBMaxRows:         "PGRST_DB_MAX_ROWS",
			DBExtraSearchPath: stringSliceEnv{key: "PGRST_DB_EXTRA_SEARCH_PATH", separator: ","},
			ApiUrl:            "SUPABASE_URL",
			APIExternalURL:    "SUPABASE_PUBLIC_URL",
			JwtSecret:         "AUTH_JWT_SECRET",
			AnonKey:           "SUPABASE_ANON_KEY",
			ServiceKey:        "SUPABASE_SERVICE_KEY",
			Host:              fixedEnvOf("HOSTNAME", "0.0.0.0"),
			LogsEnabled:       fixedEnvOf("NEXT_PUBLIC_ENABLE_LOGS", "false"),
			MetaCryptoKey:     "PG_META_CRYPTO_KEY",
		},
		Defaults: studioDefaults{
			NodeUID:         1000,
			NodeGID:         1000,
			APIPort:         3000,
			MaxRows:         1000,
			Schemas:         []string{"public", "storage", "graphql_public"},
			ExtraSearchPath: []string{"public"},
		},
	}
}
