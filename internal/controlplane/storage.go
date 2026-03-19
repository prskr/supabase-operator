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
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/prskr/supabase-operator/internal/supabase"
)

type StorageApiCluster struct {
	ServiceCluster
}

func (c *StorageApiCluster) Cluster(instance string) []*clusterv3.Cluster {
	if c == nil {
		return nil
	}

	serviceCfg := supabase.ServiceConfig.Storage

	return []*clusterv3.Cluster{
		c.ServiceCluster.Cluster(fmt.Sprintf("%s@%s", serviceCfg.Name, instance), uint32(serviceCfg.Defaults.ApiPort)),
	}
}

func (c *StorageApiCluster) Routes(instance string) []*routev3.Route {
	if c == nil {
		return nil
	}

	serviceCfg := supabase.ServiceConfig.Storage

	return []*routev3.Route{{
		Name: "Storage: /storage/v1/* -> http://storage:5000/*",
		Match: &routev3.RouteMatch{
			PathSpecifier: &routev3.RouteMatch_Prefix{
				Prefix: "/storage/v1/",
			},
		},
		Action: &routev3.Route_Route{
			Route: &routev3.RouteAction{
				ClusterSpecifier: &routev3.RouteAction_Cluster{
					Cluster: fmt.Sprintf("%s@%s", serviceCfg.Name, instance),
				},
				PrefixRewrite: "/",
			},
		},
		TypedPerFilterConfig: map[string]*anypb.Any{
			FilterNameRBAC:     MustAny(RBACPerRoute(RBACAllowAllConfig())),
			FilterNameJwtAuthn: MustAny(JWTAllowAll()),
		},
	}}
}
