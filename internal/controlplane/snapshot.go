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
	"context"
	"slices"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	router "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"google.golang.org/protobuf/types/known/anypb"
	discoveryv1 "k8s.io/api/discovery/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"code.icb4dc0.de/prskr/supabase-operator/internal/supabase"
)

type EnvoyServices struct {
	ServiceLabelKey string             `json:"-"`
	Postgrest       *PostgrestCluster  `json:"postgrest,omitempty"`
	GoTrue          *GoTrueCluster     `json:"auth,omitempty"`
	StorageApi      *StorageApiCluster `json:"storageApi,omitempty"`
	PGMeta          *PGMetaCluster     `json:"pgmeta,omitempty"`
	Studio          *StudioCluster     `json:"studio,omitempty"`
}

func (s *EnvoyServices) UpsertEndpointSlices(endpointSlices ...discoveryv1.EndpointSlice) {
	for _, eps := range endpointSlices {
		switch eps.Labels[s.ServiceLabelKey] {
		case supabase.ServiceConfig.Postgrest.Name:
			if s.Postgrest == nil {
				s.Postgrest = new(PostgrestCluster)
			}
			s.Postgrest.AddOrUpdateEndpoints(eps)
		case supabase.ServiceConfig.Auth.Name:
			if s.GoTrue == nil {
				s.GoTrue = new(GoTrueCluster)
			}
			s.GoTrue.AddOrUpdateEndpoints(eps)
		case supabase.ServiceConfig.Storage.Name:
			if s.StorageApi == nil {
				s.StorageApi = new(StorageApiCluster)
			}
			s.StorageApi.AddOrUpdateEndpoints(eps)
		case supabase.ServiceConfig.PGMeta.Name:
			if s.PGMeta == nil {
				s.PGMeta = new(PGMetaCluster)
			}
			s.PGMeta.AddOrUpdateEndpoints(eps)
		}
	}
}

func (s EnvoyServices) Targets() map[string][]string {
	targets := make(map[string][]string)

	if s.Postgrest != nil {
		targets[supabase.ServiceConfig.Postgrest.Name] = s.Postgrest.Targets()
	}

	if s.GoTrue != nil {
		targets[supabase.ServiceConfig.Auth.Name] = s.GoTrue.Targets()
	}

	if s.StorageApi != nil {
		targets[supabase.ServiceConfig.Storage.Name] = s.StorageApi.Targets()
	}

	if s.PGMeta != nil {
		targets[supabase.ServiceConfig.PGMeta.Name] = s.PGMeta.Targets()
	}

	if s.Studio != nil {
		targets[supabase.ServiceConfig.Studio.Name] = s.Studio.Targets()
	}

	return targets
}

func (s *EnvoyServices) snapshot(ctx context.Context, instance, version string) (*cache.Snapshot, error) {
	const (
		apiRouteName    = "supabase"
		studioRouteName = "supabas-studio"
		vHostName       = "supabase"
		listenerName    = "supabase"
	)

	logger := log.FromContext(ctx)

	apiConnectionManager := &hcm.HttpConnectionManager{
		CodecType:  hcm.HttpConnectionManager_AUTO,
		StatPrefix: "http",
		RouteSpecifier: &hcm.HttpConnectionManager_Rds{
			Rds: &hcm.Rds{
				ConfigSource: &corev3.ConfigSource{
					ResourceApiVersion: resource.DefaultAPIVersion,
					ConfigSourceSpecifier: &corev3.ConfigSource_ApiConfigSource{
						ApiConfigSource: &corev3.ApiConfigSource{
							TransportApiVersion:       resource.DefaultAPIVersion,
							ApiType:                   corev3.ApiConfigSource_GRPC,
							SetNodeOnFirstMessageOnly: true,
							GrpcServices: []*corev3.GrpcService{{
								TargetSpecifier: &corev3.GrpcService_EnvoyGrpc_{
									EnvoyGrpc: &corev3.GrpcService_EnvoyGrpc{ClusterName: "supabase-control-plane"},
								},
							}},
						},
					},
				},
				RouteConfigName: apiRouteName,
			},
		},
		HttpFilters: []*hcm.HttpFilter{
			{
				Name:       FilterNameJwtAuthn,
				ConfigType: &hcm.HttpFilter_TypedConfig{TypedConfig: MustAny(JWTFilterConfig())},
			},
			{
				Name:       FilterNameCORS,
				ConfigType: &hcm.HttpFilter_TypedConfig{TypedConfig: MustAny(Cors())},
			},
			{
				Name:       FilterNameHttpRouter,
				ConfigType: &hcm.HttpFilter_TypedConfig{TypedConfig: MustAny(new(router.Router))},
			},
		},
	}

	studioConnetionManager := &hcm.HttpConnectionManager{
		CodecType:  hcm.HttpConnectionManager_AUTO,
		StatPrefix: "http",
		RouteSpecifier: &hcm.HttpConnectionManager_Rds{
			Rds: &hcm.Rds{
				ConfigSource: &corev3.ConfigSource{
					ResourceApiVersion: resource.DefaultAPIVersion,
					ConfigSourceSpecifier: &corev3.ConfigSource_ApiConfigSource{
						ApiConfigSource: &corev3.ApiConfigSource{
							TransportApiVersion:       resource.DefaultAPIVersion,
							ApiType:                   corev3.ApiConfigSource_GRPC,
							SetNodeOnFirstMessageOnly: true,
							GrpcServices: []*corev3.GrpcService{{
								TargetSpecifier: &corev3.GrpcService_EnvoyGrpc_{
									EnvoyGrpc: &corev3.GrpcService_EnvoyGrpc{ClusterName: "supabase-control-plane"},
								},
							}},
						},
					},
				},
				RouteConfigName: studioRouteName,
			},
		},
		HttpFilters: []*hcm.HttpFilter{
			{
				Name:       FilterNameHttpRouter,
				ConfigType: &hcm.HttpFilter_TypedConfig{TypedConfig: MustAny(new(router.Router))},
			},
		},
	}

	apiRouteCfg := &route.RouteConfiguration{
		Name: apiRouteName,
		VirtualHosts: []*route.VirtualHost{{
			Name:    "supabase",
			Domains: []string{"*"},
			TypedPerFilterConfig: map[string]*anypb.Any{
				FilterNameJwtAuthn: MustAny(JWTPerRouteConfig()),
				FilterNameRBAC:     MustAny(RBACPerRoute(RBACRequireAuthConfig())),
			},
			Routes: slices.Concat(
				s.Postgrest.Routes(instance),
				s.GoTrue.Routes(instance),
				s.StorageApi.Routes(instance),
				s.PGMeta.Routes(instance),
			),
		}},
		TypedPerFilterConfig: map[string]*anypb.Any{
			FilterNameCORS: MustAny(CorsPolicy()),
		},
	}

	// TODO add studio route config

	listeners := []*listenerv3.Listener{{
		Name: listenerName,
		Address: &corev3.Address{
			Address: &corev3.Address_SocketAddress{
				SocketAddress: &corev3.SocketAddress{
					Protocol: corev3.SocketAddress_TCP,
					Address:  "0.0.0.0",
					PortSpecifier: &corev3.SocketAddress_PortValue{
						PortValue: 8000,
					},
				},
			},
		},
		FilterChains: []*listenerv3.FilterChain{
			{
				Filters: []*listenerv3.Filter{
					{
						Name: FilterNameHttpConnectionManager,
						ConfigType: &listenerv3.Filter_TypedConfig{
							TypedConfig: MustAny(apiConnectionManager),
						},
					},
				},
			},
		},
	}}

	if s.Studio != nil {
		logger.Info("Adding studio listener")

		listeners = append(listeners, &listenerv3.Listener{
			Name: "studio",
			Address: &corev3.Address{
				Address: &corev3.Address_SocketAddress{
					SocketAddress: &corev3.SocketAddress{
						Protocol: corev3.SocketAddress_TCP,
						Address:  "0.0.0.0",
						PortSpecifier: &corev3.SocketAddress_PortValue{
							PortValue: 3000,
						},
					},
				},
			},
			FilterChains: []*listenerv3.FilterChain{
				{
					Filters: []*listenerv3.Filter{
						{
							Name: FilterNameHttpConnectionManager,
							ConfigType: &listenerv3.Filter_TypedConfig{
								TypedConfig: MustAny(studioConnetionManager),
							},
						},
					},
				},
			},
		})
	}

	rawSnapshot := map[resource.Type][]types.Resource{
		resource.ClusterType: castResources(
			slices.Concat(
				s.Postgrest.Cluster(instance),
				s.GoTrue.Cluster(instance),
				s.StorageApi.Cluster(instance),
				s.PGMeta.Cluster(instance),
			)...),
		resource.RouteType:    {apiRouteCfg},
		resource.ListenerType: castResources(listeners...),
	}

	snapshot, err := cache.NewSnapshot(
		version,
		rawSnapshot,
	)
	if err != nil {
		return nil, err
	}

	if err := snapshot.Consistent(); err != nil {
		return nil, err
	}

	return snapshot, nil
}

func castResources[T types.Resource](from ...T) []types.Resource {
	result := make([]types.Resource, len(from))
	for idx := range from {
		result[idx] = from[idx]
	}

	return result
}
