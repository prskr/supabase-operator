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
	AnonRole                      string
	Schemas                       []string
	ExtraSearchPath               []string
	UID, GID                      int64
	ServerPort, AdminPort         int32
	ServerPortName, AdminPortName string
}

func postgrestServiceConfig() serviceConfig[postgrestEnvKeys, postgrestConfigDefaults] {
	return serviceConfig[postgrestEnvKeys, postgrestConfigDefaults]{
		Name:              "postgrest",
		LivenessProbePath: "/ready",
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
			ServerPortName:  "rest",
			AdminPortName:   "admin",
		},
	}
}
