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

type imgProxyEnvKeys struct {
	Bind                fixedEnv
	LocalFileSystemRoot stringEnv
	UseETag             fixedEnv
	EnableWebPDetection boolEnv
}

type imgProxyDefaults struct {
	ApiPort     int32
	ApiPortName string
	UID, GID    int64
}

func imgProxyServiceConfig() serviceConfig[imgProxyEnvKeys, imgProxyDefaults] {
	return serviceConfig[imgProxyEnvKeys, imgProxyDefaults]{
		Name: "imgproxy",
		EnvKeys: imgProxyEnvKeys{
			Bind:                fixedEnvOf("IMGPROXY_BIND", ":5001"),
			LocalFileSystemRoot: "IMGPROXY_LOCAL_FILESYSTEM_ROOT",
			UseETag:             fixedEnvOf("IMGPROXY_USE_ETAG", "true"),
			EnableWebPDetection: "IMGPROXY_ENABLE_WEBP_DETECTION",
		},
		Defaults: imgProxyDefaults{
			ApiPort:     5001,
			ApiPortName: "api",
			UID:         999,
			GID:         999,
		},
	}
}
