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

func studioServiceConfig() serviceConfig[studioEnvKeys, studioDefaults] {
	return serviceConfig[studioEnvKeys, studioDefaults]{
		Name:              "studio",
		LivenessProbePath: "/api/profile",
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
	}
}
