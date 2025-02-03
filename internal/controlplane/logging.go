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

package controlplane

import (
	accesslogv3 "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	streamv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/stream/v3"
	"google.golang.org/protobuf/types/known/structpb"
)

func AccessLog(name string) *accesslogv3.AccessLog {
	return &accesslogv3.AccessLog{
		Name: name,
		Filter: &accesslogv3.AccessLogFilter{
			FilterSpecifier: &accesslogv3.AccessLogFilter_NotHealthCheckFilter{
				NotHealthCheckFilter: new(accesslogv3.NotHealthCheckFilter),
			},
		},
		ConfigType: &accesslogv3.AccessLog_TypedConfig{
			TypedConfig: MustAny(&streamv3.StderrAccessLog{
				AccessLogFormat: &streamv3.StderrAccessLog_LogFormat{
					LogFormat: &corev3.SubstitutionFormatString{
						Format: &corev3.SubstitutionFormatString_JsonFormat{
							JsonFormat: &structpb.Struct{
								Fields: map[string]*structpb.Value{
									"type":                  {Kind: &structpb.Value_StringValue{StringValue: "access_log"}},
									"timestamp":             {Kind: &structpb.Value_StringValue{StringValue: "%START_TIME%"}},
									"protocol":              {Kind: &structpb.Value_StringValue{StringValue: "%PROTOCOL%"}},
									"http_method":           {Kind: &structpb.Value_StringValue{StringValue: "%REQ(:METHOD)%"}},
									"original_path":         {Kind: &structpb.Value_StringValue{StringValue: "%REQ(X-ENVOY-ORIGINAL-PATH?:PATH)%"}},
									"forwarded_for":         {Kind: &structpb.Value_StringValue{StringValue: "%REQ(X-FORWARDED-FOR)%"}},
									"user_agent":            {Kind: &structpb.Value_StringValue{StringValue: "%REQ(USER-AGENT)%"}},
									"request_id":            {Kind: &structpb.Value_StringValue{StringValue: "%REQ(X-REQUEST-ID)%"}},
									"authority":             {Kind: &structpb.Value_StringValue{StringValue: "%REQ(:AUTHORITY)%"}},
									"upstream_host":         {Kind: &structpb.Value_StringValue{StringValue: "%UPSTREAM_HOST%"}},
									"response_code":         {Kind: &structpb.Value_StringValue{StringValue: "%RESPONSE_CODE%"}},
									"response_code_details": {Kind: &structpb.Value_StringValue{StringValue: "%RESPONSE_CODE_DETAILS%"}},
									"trace_id":              {Kind: &structpb.Value_StringValue{StringValue: "%TRACE_ID%"}},
								},
							},
						},
					},
				},
			}),
		},
	}
}
