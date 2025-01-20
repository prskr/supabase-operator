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
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	"google.golang.org/protobuf/types/known/durationpb"
	discoveryv1 "k8s.io/api/discovery/v1"
)

var _ json.Marshaler = (*ServiceCluster)(nil)

type ServiceCluster struct {
	ServiceEndpoints map[string]Endpoints
}

// MarshalJSON implements json.Marshaler.
func (c *ServiceCluster) MarshalJSON() ([]byte, error) {
	tmp := struct {
		Endpoints []string `json:"endpoints"`
	}{}

	for _, endpoints := range c.ServiceEndpoints {
		tmp.Endpoints = append(tmp.Endpoints, endpoints.Targets...)
	}

	slices.Sort(tmp.Endpoints)

	return json.Marshal(tmp)
}

func (c *ServiceCluster) AddOrUpdateEndpoints(eps discoveryv1.EndpointSlice) {
	if c.ServiceEndpoints == nil {
		c.ServiceEndpoints = make(map[string]Endpoints)
	}

	c.ServiceEndpoints[eps.Name] = newEndpointsFromSlice(eps)
}

func (c ServiceCluster) Targets() []string {
	var targets []string

	for _, ep := range c.ServiceEndpoints {
		targets = append(targets, ep.Targets...)
	}

	return targets
}

func (c ServiceCluster) Cluster(name string, port uint32) *clusterv3.Cluster {
	return &clusterv3.Cluster{
		Name:                 name,
		ConnectTimeout:       durationpb.New(5 * time.Second),
		ClusterDiscoveryType: &clusterv3.Cluster_Type{Type: clusterv3.Cluster_STATIC},
		LbPolicy:             clusterv3.Cluster_ROUND_ROBIN,
		LoadAssignment: &endpointv3.ClusterLoadAssignment{
			ClusterName: name,
			Endpoints:   c.endpoints(port),
		},
	}
}

func (c ServiceCluster) endpoints(port uint32) []*endpointv3.LocalityLbEndpoints {
	eps := make([]*endpointv3.LocalityLbEndpoints, 0, len(c.ServiceEndpoints))

	for _, sep := range c.ServiceEndpoints {
		eps = append(eps, &endpointv3.LocalityLbEndpoints{
			LbEndpoints: sep.LBEndpoints(port),
		})
	}

	return eps
}

func newEndpointsFromSlice(eps discoveryv1.EndpointSlice) Endpoints {
	var result Endpoints

	for _, ep := range eps.Endpoints {
		if ep.Conditions.Ready != nil && *ep.Conditions.Ready {
			result.Addresses = append(result.Addresses, ep.Addresses...)
			result.Targets = append(result.Targets, strings.ToLower(fmt.Sprintf("%s/%s", ep.TargetRef.Kind, ep.TargetRef.Name)))
		}
	}

	return result
}

type Endpoints struct {
	Addresses []string
	Targets   []string
}

func (e Endpoints) LBEndpoints(port uint32) []*endpointv3.LbEndpoint {
	endpoints := make([]*endpointv3.LbEndpoint, 0, len(e.Addresses))

	for _, ep := range e.Addresses {
		endpoints = append(endpoints, &endpointv3.LbEndpoint{
			HostIdentifier: &endpointv3.LbEndpoint_Endpoint{
				Endpoint: &endpointv3.Endpoint{
					Address: &corev3.Address{
						Address: &corev3.Address_SocketAddress{
							SocketAddress: &corev3.SocketAddress{
								Address:  ep,
								Protocol: corev3.SocketAddress_TCP,
								PortSpecifier: &corev3.SocketAddress_PortValue{
									PortValue: port,
								},
							},
						},
					},
				},
			},
		})
	}

	return endpoints
}
