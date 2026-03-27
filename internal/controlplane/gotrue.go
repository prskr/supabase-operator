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
	"fmt"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/prskr/supabase-operator/internal/supabase"
)

type GoTrueCluster struct {
	ServiceCluster
}

func (c *GoTrueCluster) Cluster(instance string) []*clusterv3.Cluster {
	if c == nil {
		return nil
	}

	serviceCfg := supabase.ServiceConfig.Auth

	return []*clusterv3.Cluster{c.ServiceCluster.Cluster(fmt.Sprintf("%s@%s", serviceCfg.Name, instance), uint32(serviceCfg.Defaults.APIPort))}
}

func (c *GoTrueCluster) Routes(instance string) []*routev3.Route {
	const openAuthPathsRegex = `/auth/v1/(callback|verify|authorize|.well-known/jwks.json|.well-known/openid-configuration)`
	if c == nil {
		return nil
	}

	return []*routev3.Route{
		{
			Name: "GoTrue (Open) /auth/v1/(callback|verify|authorize|.well-known/jwks.json|.well-known/openid-configuration) -> http://auth:9999/$1",
			Match: &routev3.RouteMatch{
				PathSpecifier: &routev3.RouteMatch_SafeRegex{
					SafeRegex: &matcherv3.RegexMatcher{
						Regex: openAuthPathsRegex,
					},
				},
			},
			Action: &routev3.Route_Route{
				Route: &routev3.RouteAction{
					ClusterSpecifier: &routev3.RouteAction_Cluster{
						Cluster: fmt.Sprintf("%s@%s", supabase.ServiceConfig.Auth.Name, instance),
					},
					RegexRewrite: &matcherv3.RegexMatchAndSubstitute{
						Pattern: &matcherv3.RegexMatcher{
							Regex: openAuthPathsRegex,
						},
						Substitution: `/\1`,
					},
				},
			},
			TypedPerFilterConfig: map[string]*anypb.Any{
				FilterNameRBAC:     MustAny(RBACPerRoute(RBACAllowAllConfig())),
				FilterNameJwtAuthn: MustAny(JWTAllowAll()),
			},
		},
		{
			Name: "Auth: /.well-known/oauth-authorization-server -> http://auth:9999/.well-known/oauth-authorization-server",
			Match: &routev3.RouteMatch{
				PathSpecifier: &routev3.RouteMatch_Path{
					Path: "/.well-known/oauth-authorization-server",
				},
			},
			Action: &routev3.Route_Route{
				Route: &routev3.RouteAction{
					ClusterSpecifier: &routev3.RouteAction_Cluster{
						Cluster: fmt.Sprintf("%s@%s", supabase.ServiceConfig.Auth.Name, instance),
					},
				},
			},
			TypedPerFilterConfig: map[string]*anypb.Any{
				FilterNameRBAC:     MustAny(RBACPerRoute(RBACAllowAllConfig())),
				FilterNameJwtAuthn: MustAny(JWTAllowAll()),
			},
		},
		{
			Name: "GoTrue: /auth/v1/* -> http://auth:9999/*",
			Match: &routev3.RouteMatch{
				PathSpecifier: &routev3.RouteMatch_Prefix{
					Prefix: "/auth/v1",
				},
			},
			Action: &routev3.Route_Route{
				Route: &routev3.RouteAction{
					ClusterSpecifier: &routev3.RouteAction_Cluster{
						Cluster: fmt.Sprintf("%s@%s", supabase.ServiceConfig.Auth.Name, instance),
					},
					PrefixRewrite: "/",
				},
			},
		},
	}
}
