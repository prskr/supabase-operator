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
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	accesslogv3 "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	basic_authv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/basic_auth/v3"
	oauth2v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/oauth2/v3"
	routerv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	supabasev1alpha1 "github.com/prskr/supabase-operator/api/v1alpha1"
	"github.com/prskr/supabase-operator/internal/supabase"
)

var (
	errUnsupprtedEndpoointScheme = errors.New("unsupported token endpoint scheme")
	errInvalidOAuth2Config       = errors.New("invalid oauth2 config")
)

type StudioCluster struct {
	client.Client
	ServiceCluster
}

func (c *StudioCluster) Cluster(instance string, gateway *supabasev1alpha1.APIGateway) ([]*clusterv3.Cluster, error) {
	if c == nil {
		return nil, nil
	}

	serviceCfg := supabase.ServiceConfig.Studio

	clusters := []*clusterv3.Cluster{
		c.ServiceCluster.Cluster(fmt.Sprintf("%s@%s", serviceCfg.Name, instance), uint32(serviceCfg.Defaults.APIPort)),
	}

	if gateway.Spec.DashboardEndpoint.AuthType() == supabasev1alpha1.DashboardAuthTypeOAuth2 {
		if tokenEndpointCluster, err := c.oauth2TokenEndpointCluster(instance, gateway); err != nil {
			return nil, err
		} else {
			clusters = append(clusters, tokenEndpointCluster)
		}
	}

	return clusters, nil
}

func (s *StudioCluster) Listener(ctx context.Context, instance string, gateway *supabasev1alpha1.APIGateway) (*listenerv3.Listener, error) {
	if s == nil {
		return nil, nil
	}

	var httpFilters []*hcm.HttpFilter

	switch gateway.Spec.DashboardEndpoint.AuthType() {
	case supabasev1alpha1.DashboardAuthTypeOAuth2:
		if oauth2Filter := s.oauth2HttpFilter(instance, gateway); oauth2Filter != nil {
			httpFilters = append(httpFilters, oauth2Filter)
		}
	case supabasev1alpha1.DashboardAuthTypeBasic:
		if filter, err := s.basicAuthHttpFilter(ctx, gateway); err != nil {
			return nil, err
		} else if filter != nil {
			httpFilters = append(httpFilters, filter)
		}
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

	socket, err := s.dashboardTransportSocket(ctx, gateway)
	if err != nil {
		return nil, fmt.Errorf("failed to setup dashboard TLS listener: %w", err)
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
				TransportSocket: socket,
			},
		},
	}, nil
}

func (s *StudioCluster) RouteConfiguration(instance string) *routev3.RouteConfiguration {
	if s == nil {
		return nil
	}

	return &routev3.RouteConfiguration{
		Name: studioRouteName,
		VirtualHosts: []*routev3.VirtualHost{{
			Name:    "supabase-studio",
			Domains: []string{"*"},
			Routes: []*routev3.Route{{
				Name: "Studio: /* -> http://studio:3000/*",
				Match: &routev3.RouteMatch{
					PathSpecifier: &routev3.RouteMatch_Prefix{
						Prefix: "/",
					},
				},
				Action: &routev3.Route_Route{
					Route: &routev3.RouteAction{
						ClusterSpecifier: &routev3.RouteAction_Cluster{
							Cluster: fmt.Sprintf("%s@%s", supabase.ServiceConfig.Studio.Name, instance),
						},
						PrefixRewrite: "/",
					},
				},
			}},
		}},
	}
}

func (s *StudioCluster) oauth2TokenEndpointCluster(
	instance string,
	gateway *supabasev1alpha1.APIGateway,
) (*clusterv3.Cluster, error) {
	oauth2Spec := gateway.Spec.DashboardEndpoint.OAuth2()
	if oauth2Spec == nil {
		return nil, errInvalidOAuth2Config
	}

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
		return nil, fmt.Errorf("%w: %s", errUnsupprtedEndpoointScheme, parsedTokenEndpoint.Scheme)
	}

	if tokenEndpointPort := parsedTokenEndpoint.Port(); tokenEndpointPort != "" {
		if parsedPort, err := strconv.ParseUint(tokenEndpointPort, 10, 32); err != nil {
			return nil, fmt.Errorf("failed to parse token endpoint port: %w", err)
		} else {
			endpointPort = uint32(parsedPort)
		}
	}

	cluster := &clusterv3.Cluster{
		Name:           fmt.Sprintf("%s@%s", dashboardOAuth2ClusterName, instance),
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

	if gateway.Spec.Envoy != nil && gateway.Spec.Envoy.DisableIPv6 {
		cluster.DnsLookupFamily = clusterv3.Cluster_V4_ONLY
	}

	if tls {
		cluster.TransportSocket = &corev3.TransportSocket{
			Name: "envoy.transport_sockets.tls",
			ConfigType: &corev3.TransportSocket_TypedConfig{
				TypedConfig: MustAny(&tlsv3.UpstreamTlsContext{
					CommonTlsContext: new(tlsv3.CommonTlsContext),
				}),
			},
		}
	}

	return cluster, nil
}

func (s *StudioCluster) dashboardTransportSocket(
	ctx context.Context,
	gateway *supabasev1alpha1.APIGateway,
) (*corev3.TransportSocket, error) {
	tlsSpec := gateway.Spec.DashboardEndpoint.TLSSpec()
	if tlsSpec == nil {
		return nil, nil
	}

	dashboardTlsSecret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tlsSpec.Cert.SecretName,
			Namespace: gateway.Namespace,
		},
	}

	if err := s.Get(ctx, client.ObjectKeyFromObject(&dashboardTlsSecret), &dashboardTlsSecret); err != nil {
		if client.IgnoreNotFound(err) == nil {
			return nil, nil
		}

		return nil, err
	}

	return &corev3.TransportSocket{
		Name: SocketNameTLS,
		ConfigType: &corev3.TransportSocket_TypedConfig{
			TypedConfig: MustAny(&tlsv3.DownstreamTlsContext{
				RequireClientCertificate: &wrapperspb.BoolValue{Value: false},
				CommonTlsContext: &tlsv3.CommonTlsContext{
					TlsCertificates: []*tlsv3.TlsCertificate{{
						CertificateChain: &corev3.DataSource{
							Specifier: &corev3.DataSource_InlineBytes{
								InlineBytes: dashboardTlsSecret.Data[corev1.TLSCertKey],
							},
						},
						PrivateKey: &corev3.DataSource{
							Specifier: &corev3.DataSource_InlineBytes{
								InlineBytes: dashboardTlsSecret.Data[corev1.TLSPrivateKeyKey],
							},
						},
					}},
					ValidationContextType: &tlsv3.CommonTlsContext_ValidationContext{
						ValidationContext: &tlsv3.CertificateValidationContext{
							TrustedCa: &corev3.DataSource{
								Specifier: &corev3.DataSource_InlineBytes{
									InlineBytes: dashboardTlsSecret.Data["ca.crt"],
								},
							},
						},
					},
				},
			}),
		},
	}, nil
}

func (s *StudioCluster) basicAuthHttpFilter(ctx context.Context, gateway *supabasev1alpha1.APIGateway) (*hcm.HttpFilter, error) {
	users := gateway.Spec.DashboardEndpoint.Auth.Basic.UsersInline

	usersSecret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      gateway.Spec.DashboardEndpoint.Auth.Basic.PlaintextUsersSecretRef,
			Namespace: gateway.Namespace,
		},
	}

	if err := s.Get(ctx, client.ObjectKeyFromObject(&usersSecret), &usersSecret); err != nil {
		if client.IgnoreNotFound(err) != nil {
			return nil, fmt.Errorf("failed to fetch credentials secret: %w", err)
		}
	}

	hash := sha1.New()
	for username, passwd := range usersSecret.Data {
		_, _ = hash.Write(passwd)
		pwHash := base64.StdEncoding.EncodeToString(hash.Sum(nil))
		users = append(users, fmt.Sprintf("%s:{SHA}%s", username, pwHash))
		hash.Reset()
	}

	if len(users) == 0 {
		return nil, nil
	}

	slices.Sort(users)

	return &hcm.HttpFilter{
		Name: FilterNameBasicAuth,
		ConfigType: &hcm.HttpFilter_TypedConfig{
			TypedConfig: MustAny(&basic_authv3.BasicAuth{
				Users: &corev3.DataSource{
					Specifier: &corev3.DataSource_InlineString{
						InlineString: strings.Join(users, "\n"),
					},
				},
			}),
		},
	}, nil
}

func (s *StudioCluster) oauth2HttpFilter(instance string, gateway *supabasev1alpha1.APIGateway) *hcm.HttpFilter {
	var (
		serviceCfg = supabase.ServiceConfig.Envoy
		oauth2Spec = gateway.Spec.DashboardEndpoint.OAuth2()
	)

	if oauth2Spec == nil {
		return nil
	}

	return &hcm.HttpFilter{
		Name: FilterNameOAuth2,
		ConfigType: &hcm.HttpFilter_TypedConfig{TypedConfig: MustAny(&oauth2v3.OAuth2{
			Config: &oauth2v3.OAuth2Config{
				TokenEndpoint: &corev3.HttpUri{
					HttpUpstreamType: &corev3.HttpUri_Cluster{
						Cluster: fmt.Sprintf("%s@%s", dashboardOAuth2ClusterName, instance),
					},
					Uri:     gateway.Spec.DashboardEndpoint.Auth.OAuth2.TokenEndpoint,
					Timeout: durationpb.New(3 * time.Second),
				},
				AuthorizationEndpoint: gateway.Spec.DashboardEndpoint.Auth.OAuth2.AuthorizationEndpoint,
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
	}
}
