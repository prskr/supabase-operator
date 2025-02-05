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
	"slices"

	accesslogv3 "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	routerv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

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
				s.Studio = &StudioCluster{Client: s.Client}
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

func (s *EnvoyServices) snapshot(
	ctx context.Context,
	instance, version string,
) (snapshot *cache.Snapshot, snapshotHash []byte, err error) {
	socket, err := s.apiTransportSocket(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to setup API TLS listener: %w", err)
	}

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
				TransportSocket: socket,
			},
		},
	}}

	if studioListener, err := s.Studio.Listener(ctx, instance, s.Gateway); err != nil {
		return nil, nil, err
	} else if studioListener != nil {
		listeners = append(listeners, studioListener)
	}

	routes := []types.Resource{s.apiRouteConfiguration(instance)}

	if studioRouteCfg := s.Studio.RouteConfiguration(instance); studioRouteCfg != nil {
		routes = append(routes, studioRouteCfg)
	}

	studioClusters, err := s.Studio.Cluster(instance, s.Gateway)
	if err != nil {
		return nil, nil, err
	}

	clusters := castResources(
		slices.Concat(
			s.Postgrest.Cluster(instance),
			s.GoTrue.Cluster(instance),
			s.StorageApi.Cluster(instance),
			s.PGMeta.Cluster(instance),
			studioClusters,
		)...)

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

func (s *EnvoyServices) apiTransportSocket(ctx context.Context) (*corev3.TransportSocket, error) {
	tlsSpec := s.Gateway.Spec.ApiEndpoint.TLSSpec()
	if tlsSpec == nil {
		return nil, nil
	}

	apiTlsSecret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tlsSpec.Cert.SecretName,
			Namespace: s.Gateway.Namespace,
		},
	}

	if err := s.Get(ctx, client.ObjectKeyFromObject(&apiTlsSecret), &apiTlsSecret); err != nil {
		if client.IgnoreNotFound(err) == nil {
			return nil, nil
		}

		return nil, err
	}

	return &corev3.TransportSocket{
		Name: SocketNameTLS,
		ConfigType: &corev3.TransportSocket_TypedConfig{
			TypedConfig: MustAny(&tlsv3.DownstreamTlsContext{
				CommonTlsContext: &tlsv3.CommonTlsContext{
					TlsCertificates: []*tlsv3.TlsCertificate{{
						CertificateChain: &corev3.DataSource{
							Specifier: &corev3.DataSource_InlineBytes{
								InlineBytes: apiTlsSecret.Data[corev1.TLSCertKey],
							},
						},
						PrivateKey: &corev3.DataSource{
							Specifier: &corev3.DataSource_InlineBytes{
								InlineBytes: apiTlsSecret.Data[corev1.TLSPrivateKeyKey],
							},
						},
					}},
					ValidationContextType: &tlsv3.CommonTlsContext_ValidationContext{
						ValidationContext: &tlsv3.CertificateValidationContext{
							TrustedCa: &corev3.DataSource{
								Specifier: &corev3.DataSource_InlineBytes{
									InlineBytes: apiTlsSecret.Data["ca.crt"],
								},
							},
						},
					},
				},
			}),
		},
	}, nil
}

func castResources[T types.Resource](from ...T) []types.Resource {
	result := make([]types.Resource, len(from))
	for idx := range from {
		result[idx] = from[idx]
	}

	return result
}
