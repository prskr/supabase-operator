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

	"code.icb4dc0.de/prskr/supabase-operator/internal/supabase"
)

type StudioCluster struct {
	ServiceCluster
}

func (c *StudioCluster) Cluster(instance string) []*clusterv3.Cluster {
	if c == nil {
		return nil
	}

	serviceCfg := supabase.ServiceConfig.Studio

	return []*clusterv3.Cluster{
		c.ServiceCluster.Cluster(fmt.Sprintf("%s@%s", serviceCfg.Name, instance), uint32(serviceCfg.Defaults.APIPort)),
	}
}

func (c *StudioCluster) Routes(instance string) []*routev3.Route {
	if c == nil {
		return nil
	}

	return []*routev3.Route{{
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
	}}
}
