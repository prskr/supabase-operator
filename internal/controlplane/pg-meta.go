package controlplane

import (
	"fmt"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"

	"code.icb4dc0.de/prskr/supabase-operator/internal/supabase"
)

type PGMetaCluster struct {
	ServiceCluster
}

func (c *PGMetaCluster) Cluster(instance string) []*clusterv3.Cluster {
	if c == nil {
		return nil
	}
	return []*clusterv3.Cluster{
		c.ServiceCluster.Cluster(fmt.Sprintf("%s@%s", supabase.ServiceConfig.PGMeta.Name, instance), uint32(supabase.ServiceConfig.PGMeta.Defaults.APIPort)),
	}
}

func (c *PGMetaCluster) Routes(instance string) []*routev3.Route {
	if c == nil {
		return nil
	}

	return []*routev3.Route{{
		Name: "pg-meta: /pg/* -> http://pg-meta:8080/*",
		Match: &routev3.RouteMatch{
			PathSpecifier: &routev3.RouteMatch_Prefix{
				Prefix: "/pg/",
			},
		},
		Action: &routev3.Route_Route{
			Route: &routev3.RouteAction{
				ClusterSpecifier: &routev3.RouteAction_Cluster{
					Cluster: fmt.Sprintf("%s@%s", supabase.ServiceConfig.PGMeta.Name, instance),
				},
				PrefixRewrite: "/",
			},
		},
	}}
}
