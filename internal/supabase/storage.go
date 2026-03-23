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

import "strconv"

type storageEnvApiKeys struct {
	AnonKey                        secretEnv
	ServiceKey                     secretEnv
	JwtSecret                      secretEnv
	JwtJwks                        secretEnv
	DatabaseDSN                    stringEnv
	FileSizeLimit                  intEnv[uint64]
	UploadFileSizeLimit            intEnv[uint64]
	UploadFileSizeLimitStandard    intEnv[uint64]
	StorageBackend                 stringEnv
	FileStorageBackendPath         stringEnv
	FileStorageRegion              stringEnv
	TenantID                       stringEnv
	StorageS3Bucket                stringEnv
	StorageS3MaxSockets            intEnv[uint8]
	StorageS3Endpoint              stringEnv
	StorageS3ForcePathStyle        boolEnv
	StorageS3Region                stringEnv
	StorageS3AccessKeyID           secretEnv
	StorageS3AccessSecretKey       secretEnv
	EnableImaageTransformation     boolEnv
	ImgProxyURL                    stringEnv
	TusURLPath                     fixedEnv
	S3ProtocolAccessKeyID          secretEnv
	S3ProtocolAccessKeySecret      secretEnv
	S3ProtocolAllowForwardedHeader boolEnv
	S3ProtocolPrefix               fixedEnv
	AllowForwardedPathHeader       fixedEnv
}

type storageApiDefaults struct {
	APIPort     int32
	APIPortName string
	UID, GID    int64
}

func storageServiceConfig() serviceConfig[storageEnvApiKeys, storageApiDefaults] {
	return serviceConfig[storageEnvApiKeys, storageApiDefaults]{
		Name:              "storage-api",
		LivenessProbePath: "/status",
		EnvKeys: storageEnvApiKeys{
			AnonKey:                        "ANON_KEY",
			ServiceKey:                     "SERVICE_KEY",
			JwtSecret:                      "AUTH_JWT_SECRET",
			JwtJwks:                        "AUTH_JWT_JWKS",
			StorageBackend:                 "STORAGE_BACKEND",
			FileStorageBackendPath:         "FILE_STORAGE_BACKEND_PATH",
			FileStorageRegion:              "REGION",
			DatabaseDSN:                    "DATABASE_URL",
			FileSizeLimit:                  "FILE_SIZE_LIMIT",
			UploadFileSizeLimit:            "UPLOAD_FILE_SIZE_LIMIT",
			UploadFileSizeLimitStandard:    "UPLOAD_FILE_SIZE_LIMIT_STANDARD",
			TenantID:                       "TENANT_ID",
			StorageS3Bucket:                "STORAGE_S3_BUCKET",
			StorageS3MaxSockets:            "STORAGE_S3_MAX_SOCKETS",
			StorageS3Endpoint:              "STORAGE_S3_ENDPOINT",
			StorageS3ForcePathStyle:        "STORAGE_S3_FORCE_PATH_STYLE",
			StorageS3Region:                "STORAGE_S3_REGION",
			StorageS3AccessKeyID:           "AWS_ACCESS_KEY_ID",
			StorageS3AccessSecretKey:       "AWS_SECRET_ACCESS_KEY",
			EnableImaageTransformation:     "ENABLE_IMAGE_TRANSFORMATION",
			ImgProxyURL:                    "IMGPROXY_URL",
			TusURLPath:                     fixedEnvOf("TUS_URL_PATH", "/storage/v1/upload/resumable"),
			S3ProtocolAccessKeyID:          "S3_PROTOCOL_ACCESS_KEY_ID",
			S3ProtocolAccessKeySecret:      "S3_PROTOCOL_ACCESS_KEY_SECRET",
			S3ProtocolPrefix:               fixedEnvOf("S3_PROTOCOL_PREFIX", "/storage/v1"),
			S3ProtocolAllowForwardedHeader: "S3_ALLOW_FORWARDED_HEADER",
			AllowForwardedPathHeader:       fixedEnvOf("REQUEST_ALLOW_X_FORWARDED_PATH", strconv.FormatBool(true)),
		},
		Defaults: storageApiDefaults{
			APIPort:     5000,
			APIPortName: "api",
			UID:         1000,
			GID:         1000,
		},
	}
}
