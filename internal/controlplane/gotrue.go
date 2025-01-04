package controlplane

import (
	"fmt"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"google.golang.org/protobuf/types/known/anypb"

	"code.icb4dc0.de/prskr/supabase-operator/internal/supabase"
)

type GoTrueCluster struct {
	ServiceCluster
}

func (c *GoTrueCluster) Cluster(instance string) []*clusterv3.Cluster {
	if c == nil {
		return nil
	}

	return []*clusterv3.Cluster{c.ServiceCluster.Cluster(fmt.Sprintf("auth@%s", instance), 9999)}
}

func (c *GoTrueCluster) Routes(instance string) []*routev3.Route {
	if c == nil {
		return nil
	}

	return []*routev3.Route{
		{
			Name: "GoTrue (Open) /auth/v1/(callback|verify) -> http://auth:9999/$1",
			Match: &routev3.RouteMatch{
				PathSpecifier: &routev3.RouteMatch_SafeRegex{
					SafeRegex: &matcherv3.RegexMatcher{
						Regex: `/auth/v1/(callback|verify|authorize)`,
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
							Regex: `/auth/v1/(callback|verify|authorize)`,
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
