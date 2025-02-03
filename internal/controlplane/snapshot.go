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
	"crypto/sha256"
	"fmt"
	"net/url"
	"slices"
	"strconv"
	"time"

	accesslogv3 "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	oauth2v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/oauth2/v3"
	routerv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	supabasev1alpha1 "code.icb4dc0.de/prskr/supabase-operator/api/v1alpha1"
	"code.icb4dc0.de/prskr/supabase-operator/internal/supabase"
)

const (
	studioRouteName            = "supabase-studio"
	dashboardOAuth2ClusterName = "dashboard-oauth"
	apiRouteName               = "supabase-api"
	apilistenerName            = "supabase-api"
)

type EnvoyServices struct {
	client.Client
	ServiceLabelKey string
	Gateway         *supabasev1alpha1.APIGateway
	Postgrest       *PostgrestCluster
	GoTrue          *GoTrueCluster
	StorageApi      *StorageApiCluster
	PGMeta          *PGMetaCluster
	Studio          *StudioCluster
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
		case supabase.ServiceConfig.Studio.Name:
			if s.Studio == nil {
				s.Studio = new(StudioCluster)
			}

			s.Studio.AddOrUpdateEndpoints(eps)
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

func (s *EnvoyServices) snapshot(ctx context.Context, instance, version string) (snapshot *cache.Snapshot, snapshotHash []byte, err error) {
	logger := log.FromContext(ctx)

	listeners := []*listenerv3.Listener{{
		Name: apilistenerName,
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
							TypedConfig: MustAny(s.apiConnectionManager()),
						},
					},
				},
			},
		},
	}}

	if studioListener := s.studioListener(); studioListener != nil {
		logger.Info("Adding studio listener")
		listeners = append(listeners, studioListener)
	}

	routes := []types.Resource{s.apiRouteConfiguration(instance)}

	if studioRouteCfg := s.studioRoute(instance); studioRouteCfg != nil {
		logger.Info("Adding studio route")
		routes = append(routes, studioRouteCfg)
	}

	clusters := castResources(
		slices.Concat(
			s.Postgrest.Cluster(instance),
			s.GoTrue.Cluster(instance),
			s.StorageApi.Cluster(instance),
			s.PGMeta.Cluster(instance),
			s.Studio.Cluster(instance),
		)...)

	if oauth2Spec := s.Gateway.Spec.DashboardEndpoint.OAuth2(); oauth2Spec != nil {
		if oauth2TokenEndpointCluster, err := s.oauth2TokenEndpointCluster(); err != nil {
			return nil, nil, err
		} else {
			logger.Info("Adding OAuth2 token endpoint cluster")
			clusters = append(clusters, oauth2TokenEndpointCluster)
		}
	}

	sdsSecrets, err := s.secrets(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to collect dynamic secrets: %w", err)
	}

	rawSnapshot := map[resource.Type][]types.Resource{
		resource.ClusterType:  clusters,
		resource.RouteType:    routes,
		resource.ListenerType: castResources(listeners...),
		resource.SecretType:   sdsSecrets,
	}

	hash := sha256.New()

	for _, resType := range []resource.Type{resource.ClusterType, resource.RouteType, resource.ListenerType, resource.SecretType} {
		for _, resource := range rawSnapshot[resType] {
			if raw, err := proto.Marshal(resource); err != nil {
				return nil, nil, err
			} else {
				_, _ = hash.Write(raw)
			}
		}
	}

	snapshot, err = cache.NewSnapshot(
		version,
		rawSnapshot,
	)
	if err != nil {
		return nil, nil, err
	}

	if err := snapshot.Consistent(); err != nil {
		return nil, nil, err
	}

	return snapshot, hash.Sum(nil), nil
}

func (s *EnvoyServices) secrets(ctx context.Context) ([]types.Resource, error) {
	var (
		serviceConfig = supabase.ServiceConfig.Envoy
		resources     = make([]types.Resource, 0, 2)
	)

	hmacSecret, err := s.k8sSecretKeyToSecret(
		ctx,
		serviceConfig.HmacSecretName(s.Gateway),
		serviceConfig.Defaults.HmacSecretKey,
		serviceConfig.Defaults.HmacSecretKey,
	)

	if err == nil {
		resources = append(resources, hmacSecret)
	} else if client.IgnoreNotFound(err) != nil {
		return nil, err
	}

	if oauth2Spec := s.Gateway.Spec.DashboardEndpoint.OAuth2(); oauth2Spec != nil {
		oauth2ClientSecret, err := s.k8sSecretKeyToSecret(
			ctx,
			oauth2Spec.ClientSecretRef.Name,
			oauth2Spec.ClientSecretRef.Key,
			serviceConfig.Defaults.OAuth2ClientSecretKey,
		)

		if err == nil {
			resources = append(resources, oauth2ClientSecret)
		} else if client.IgnoreNotFound(err) != nil {
			return nil, err
		}
	}

	return resources, nil
}

func (s *EnvoyServices) k8sSecretKeyToSecret(ctx context.Context, secretName, secretKey, envoySecretName string) (*tlsv3.Secret, error) {
	k8sSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: s.Gateway.Namespace,
		},
	}

	if err := s.Get(ctx, client.ObjectKeyFromObject(k8sSecret), k8sSecret); err != nil {
		return nil, err
	}

	return &tlsv3.Secret{
		Name: envoySecretName,
		Type: &tlsv3.Secret_GenericSecret{
			GenericSecret: &tlsv3.GenericSecret{
				Secret: &corev3.DataSource{
					Specifier: &corev3.DataSource_InlineBytes{
						InlineBytes: k8sSecret.Data[secretKey],
					},
				},
			},
		},
	}, nil
}

func (s *EnvoyServices) apiConnectionManager() *hcm.HttpConnectionManager {
	return &hcm.HttpConnectionManager{
		CodecType:  hcm.HttpConnectionManager_AUTO,
		StatPrefix: "supabase_rest",
		AccessLog:  []*accesslogv3.AccessLog{AccessLog("supabase-rest-access-log")},
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
				ConfigType: &hcm.HttpFilter_TypedConfig{TypedConfig: MustAny(new(routerv3.Router))},
			},
		},
	}
}

func (s *EnvoyServices) apiRouteConfiguration(instance string) *routev3.RouteConfiguration {
	return &routev3.RouteConfiguration{
		Name: apiRouteName,
		VirtualHosts: []*routev3.VirtualHost{{
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
}

func (s *EnvoyServices) studioListener() *listenerv3.Listener {
	if s.Studio == nil {
		return nil
	}

	var (
		httpFilters []*hcm.HttpFilter
		serviceCfg  = supabase.ServiceConfig.Envoy
	)

	if oauth2Spec := s.Gateway.Spec.DashboardEndpoint.OAuth2(); oauth2Spec != nil {
		httpFilters = append(httpFilters, &hcm.HttpFilter{
			Name: FilterNameOAuth2,
			ConfigType: &hcm.HttpFilter_TypedConfig{TypedConfig: MustAny(&oauth2v3.OAuth2{
				Config: &oauth2v3.OAuth2Config{
					TokenEndpoint: &corev3.HttpUri{
						HttpUpstreamType: &corev3.HttpUri_Cluster{
							Cluster: dashboardOAuth2ClusterName,
						},
						Uri:     s.Gateway.Spec.DashboardEndpoint.Auth.OAuth2.TokenEndpoint,
						Timeout: durationpb.New(3 * time.Second),
					},
					AuthorizationEndpoint: s.Gateway.Spec.DashboardEndpoint.Auth.OAuth2.AuthorizationEndpoint,
					RedirectUri:           "%REQ(x-forwarded-proto)%://%REQ(:authority)%/callback",
					RedirectPathMatcher: &matcherv3.PathMatcher{
						Rule: &matcherv3.PathMatcher_Path{
							Path: &matcherv3.StringMatcher{
								MatchPattern: &matcherv3.StringMatcher_Exact{
									Exact: "/callback",
								},
							},
						},
					},
					SignoutPath: &matcherv3.PathMatcher{
						Rule: &matcherv3.PathMatcher_Path{
							Path: &matcherv3.StringMatcher{
								MatchPattern: &matcherv3.StringMatcher_Exact{
									Exact: "/signout",
								},
							},
						},
					},
					Credentials: &oauth2v3.OAuth2Credentials{
						ClientId: oauth2Spec.ClientID,
						TokenSecret: &tlsv3.SdsSecretConfig{
							Name: serviceCfg.Defaults.OAuth2ClientSecretKey,
							SdsConfig: &corev3.ConfigSource{
								ConfigSourceSpecifier: &corev3.ConfigSource_Ads{
									Ads: new(corev3.AggregatedConfigSource),
								},
							},
						},
						TokenFormation: &oauth2v3.OAuth2Credentials_HmacSecret{
							HmacSecret: &tlsv3.SdsSecretConfig{
								Name: serviceCfg.Defaults.HmacSecretKey,
								SdsConfig: &corev3.ConfigSource{
									ConfigSourceSpecifier: &corev3.ConfigSource_Ads{
										Ads: new(corev3.AggregatedConfigSource),
									},
								},
							},
						},
					},
					AuthScopes: oauth2Spec.Scopes,
					Resources:  oauth2Spec.Resources,
				},
			})},
		})
	}

	studioConnetionManager := &hcm.HttpConnectionManager{
		CodecType:  hcm.HttpConnectionManager_AUTO,
		StatPrefix: "supbase_studio",
		AccessLog:  []*accesslogv3.AccessLog{AccessLog("supbase_studio_access_log")},
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
		HttpFilters: append(httpFilters, &hcm.HttpFilter{
			Name:       FilterNameHttpRouter,
			ConfigType: &hcm.HttpFilter_TypedConfig{TypedConfig: MustAny(new(routerv3.Router))},
		}),
	}

	return &listenerv3.Listener{
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
	}
}

func (s *EnvoyServices) studioRoute(instance string) *routev3.RouteConfiguration {
	if s.Studio == nil {
		return nil
	}

	return &routev3.RouteConfiguration{
		Name: studioRouteName,
		VirtualHosts: []*routev3.VirtualHost{{
			Name:    "supabase-studio",
			Domains: []string{"*"},
			Routes:  s.Studio.Routes(instance),
		}},
	}
}

func (s *EnvoyServices) oauth2TokenEndpointCluster() (*clusterv3.Cluster, error) {
	oauth2Spec := s.Gateway.Spec.DashboardEndpoint.OAuth2()
	parsedTokenEndpoint, err := url.Parse(oauth2Spec.TokenEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token endpoint: %w", err)
	}

	var (
		endpointPort uint32
		tls          bool
	)
	switch parsedTokenEndpoint.Scheme {
	case "http":
		endpointPort = 80
	case "https":
		endpointPort = 443
		tls = true
	default:
		return nil, fmt.Errorf("unsupported token endpoint scheme: %s", parsedTokenEndpoint.Scheme)
	}

	if tokenEndpointPort := parsedTokenEndpoint.Port(); tokenEndpointPort != "" {
		if parsedPort, err := strconv.ParseUint(tokenEndpointPort, 10, 32); err != nil {
			return nil, fmt.Errorf("failed to parse token endpoint port: %w", err)
		} else {
			endpointPort = uint32(parsedPort)
		}
	}

	cluster := &clusterv3.Cluster{
		Name:           dashboardOAuth2ClusterName,
		ConnectTimeout: durationpb.New(3 * time.Second),
		ClusterDiscoveryType: &clusterv3.Cluster_Type{
			Type: clusterv3.Cluster_LOGICAL_DNS,
		},
		LbPolicy: clusterv3.Cluster_ROUND_ROBIN,
		LoadAssignment: &endpointv3.ClusterLoadAssignment{
			ClusterName: dashboardOAuth2ClusterName,
			Endpoints: []*endpointv3.LocalityLbEndpoints{{
				LbEndpoints: []*endpointv3.LbEndpoint{{
					HostIdentifier: &endpointv3.LbEndpoint_Endpoint{
						Endpoint: &endpointv3.Endpoint{
							Address: &corev3.Address{
								Address: &corev3.Address_SocketAddress{
									SocketAddress: &corev3.SocketAddress{
										Address: parsedTokenEndpoint.Hostname(),
										PortSpecifier: &corev3.SocketAddress_PortValue{
											PortValue: endpointPort,
										},
										Protocol: corev3.SocketAddress_TCP,
									},
								},
							},
						},
					},
				}},
			}},
		},
	}

	if s.Gateway.Spec.Envoy != nil && s.Gateway.Spec.Envoy.DisableIPv6 {
		cluster.DnsLookupFamily = clusterv3.Cluster_V4_ONLY
	}

	if tls {
		cluster.TransportSocket = &corev3.TransportSocket{
			Name: "envoy.transport_sockets.tls",
			ConfigType: &corev3.TransportSocket_TypedConfig{
				TypedConfig: MustAny(&tlsv3.UpstreamTlsContext{
					Sni:                parsedTokenEndpoint.Hostname(),
					AllowRenegotiation: true,
					CommonTlsContext: &tlsv3.CommonTlsContext{
						TlsParams: &tlsv3.TlsParameters{
							TlsMinimumProtocolVersion: tlsv3.TlsParameters_TLSv1_2,
						},
					},
				}),
			},
		}
	}

	return cluster, nil
}

func castResources[T types.Resource](from ...T) []types.Resource {
	result := make([]types.Resource, len(from))
	for idx := range from {
		result[idx] = from[idx]
	}

	return result
}
