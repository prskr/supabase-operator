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
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"

	"code.icb4dc0.de/prskr/supabase-operator/internal/supabase"
)

type PostgrestCluster struct {
	ServiceCluster
}

func (c *PostgrestCluster) Cluster(instance string) []*clusterv3.Cluster {
	if c == nil {
		return nil
	}
	return []*clusterv3.Cluster{
		c.ServiceCluster.Cluster(fmt.Sprintf("%s@%s", supabase.ServiceConfig.Postgrest.Name, instance), 3000),
	}
}

func (c *PostgrestCluster) Routes(instance string) []*routev3.Route {
	if c == nil {
		return nil
	}

	return []*routev3.Route{
		{
			Name: "PostgREST: /rest/v1/* -> http://rest:3000/*",
			Match: &routev3.RouteMatch{
				PathSpecifier: &routev3.RouteMatch_Prefix{
					Prefix: "/rest/v1/",
				},
			},
			Action: &routev3.Route_Route{
				Route: &routev3.RouteAction{
					ClusterSpecifier: &routev3.RouteAction_Cluster{
						Cluster: fmt.Sprintf("%s@%s", supabase.ServiceConfig.Postgrest.Name, instance),
					},
					PrefixRewrite: "/",
				},
			},
		},
		{
			Name: "PostgREST: /graphql/v1/* -> http://rest:3000/rpc/graphql",
			Match: &routev3.RouteMatch{
				PathSpecifier: &routev3.RouteMatch_Prefix{
					Prefix: "/graphql/v1",
				},
			},
			Action: &routev3.Route_Route{
				Route: &routev3.RouteAction{
					ClusterSpecifier: &routev3.RouteAction_Cluster{
						Cluster: fmt.Sprintf("%s@%s", supabase.ServiceConfig.Postgrest.Name, instance),
					},
					PrefixRewrite: "/rpc/graphql",
				},
			},
			RequestHeadersToAdd: []*corev3.HeaderValueOption{{
				Header: &corev3.HeaderValue{
					Key:   "Content-Profile",
					Value: "graphql_public",
				},
			}},
		},
	}
}
