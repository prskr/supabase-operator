/*
Copyright 2024 Peter Kurfer.

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
	"errors"
	"fmt"
	"slices"
	"strconv"
	"sync"
	"time"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	router "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"google.golang.org/protobuf/types/known/anypb"
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/watch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"code.icb4dc0.de/prskr/supabase-operator/internal/meta"
	"code.icb4dc0.de/prskr/supabase-operator/internal/supabase"
)

var (
	ErrUnexpectedObject    = errors.New("unexpected object")
	ErrNoEnvoyClusterLabel = errors.New("no Envoy cluster label set")

	supabaseServices = []string{
		supabase.ServiceConfig.Postgrest.Name,
		supabase.ServiceConfig.Auth.Name,
		supabase.ServiceConfig.PGMeta.Name,
	}
)

type EndpointsController struct {
	lock          sync.Mutex
	Client        client.WithWatch
	Cache         cache.SnapshotCache
	envoyClusters map[string]*envoyClusterServices
}

func (c *EndpointsController) Run(ctx context.Context) error {
	var (
		logger         = ctrl.Log.WithName("endpoints-controller")
		endpointSlices discoveryv1.EndpointSliceList
	)

	selector := labels.NewSelector()

	partOfRequirement, err := labels.NewRequirement(meta.WellKnownLabel.PartOf, selection.Equals, []string{"supabase"})
	if err != nil {
		return fmt.Errorf("preparing watcher selectors: %w", err)
	}

	nameRequirement, err := labels.NewRequirement(meta.WellKnownLabel.Name, selection.In, supabaseServices)
	if err != nil {
		return fmt.Errorf("preparing watcher selectors: %w", err)
	}

	envoyClusterRequirement, err := labels.NewRequirement(meta.SupabaseLabel.EnvoyCluster, selection.Exists, nil)
	if err != nil {
		return fmt.Errorf("preparing watcher selectors: %w", err)
	}

	selector.Add(*partOfRequirement, *nameRequirement, *envoyClusterRequirement)

	watcher, err := c.Client.Watch(
		ctx,
		&endpointSlices,
		client.MatchingLabelsSelector{
			Selector: selector.Add(*partOfRequirement, *nameRequirement, *envoyClusterRequirement),
		},
	)
	if err != nil {
		return err
	}

	defer watcher.Stop()

	for {
		select {
		case ev, more := <-watcher.ResultChan():
			if !more {
				return nil
			}
			eventLogger := logger.WithValues("event_type", ev.Type)
			switch ev.Type {
			case watch.Added, watch.Modified:
				eps, ok := ev.Object.(*discoveryv1.EndpointSlice)
				if !ok {
					logger.Error(fmt.Errorf("%w: %T", ErrUnexpectedObject, ev.Object), "expected EndpointSlice but got a different object type")
					continue
				}

				if err := c.handleModificationEvent(log.IntoContext(ctx, eventLogger), eps); err != nil {
					logger.Error(err, "error occurred during event handling")
				}
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (c *EndpointsController) handleModificationEvent(ctx context.Context, epSlice *discoveryv1.EndpointSlice) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	var (
		logger      = log.FromContext(ctx)
		instanceKey string
		svc         *envoyClusterServices
	)

	logger.Info("Observed endpoint slice", "name", epSlice.Name)

	if c.envoyClusters == nil {
		c.envoyClusters = make(map[string]*envoyClusterServices)
	}

	envoyNodeName, ok := epSlice.Labels[meta.SupabaseLabel.EnvoyCluster]
	if !ok {
		return fmt.Errorf("%w: at object %s", ErrNoEnvoyClusterLabel, epSlice.Name)
	}

	instanceKey = fmt.Sprintf("%s:%s", envoyNodeName, epSlice.Namespace)

	if svc, ok = c.envoyClusters[instanceKey]; !ok {
		svc = new(envoyClusterServices)
	}

	svc.UpsertEndpoints(epSlice)

	c.envoyClusters[instanceKey] = svc

	return c.updateSnapshot(ctx, instanceKey)
}

func (c *EndpointsController) updateSnapshot(ctx context.Context, instance string) error {
	latestVersion := strconv.FormatInt(time.Now().UTC().UnixMilli(), 10)

	snapshot, err := c.envoyClusters[instance].snapshot(instance, latestVersion)
	if err != nil {
		return err
	}

	return c.Cache.SetSnapshot(ctx, instance, snapshot)
}

type envoyClusterServices struct {
	Postgrest *PostgrestCluster
	GoTrue    *GoTrueCluster
	PGMeta    *PGMetaCluster
}

func (s *envoyClusterServices) UpsertEndpoints(eps *discoveryv1.EndpointSlice) {
	switch eps.Labels[meta.WellKnownLabel.Name] {
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
	case supabase.ServiceConfig.PGMeta.Name:
		if s.PGMeta == nil {
			s.PGMeta = new(PGMetaCluster)
		}
		s.PGMeta.AddOrUpdateEndpoints(eps)
	}
}

func (s *envoyClusterServices) snapshot(instance, version string) (*cache.Snapshot, error) {
	const (
		routeName    = "supabase"
		vHostName    = "supabase"
		listenerName = "supabase"
	)

	manager := &hcm.HttpConnectionManager{
		CodecType:  hcm.HttpConnectionManager_AUTO,
		StatPrefix: "http",
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
				RouteConfigName: routeName,
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
				ConfigType: &hcm.HttpFilter_TypedConfig{TypedConfig: MustAny(new(router.Router))},
			},
		},
	}

	routeCfg := &route.RouteConfiguration{
		Name: routeName,
		VirtualHosts: []*route.VirtualHost{{
			Name:    "supabase",
			Domains: []string{"*"},
			TypedPerFilterConfig: map[string]*anypb.Any{
				FilterNameJwtAuthn: MustAny(JWTPerRouteConfig()),
				FilterNameRBAC:     MustAny(RBACPerRoute(RBACRequireAuthConfig())),
			},
			Routes: slices.Concat(
				s.Postgrest.Routes(instance),
				s.GoTrue.Routes(instance),
				s.PGMeta.Routes(instance),
			),
		}},
		TypedPerFilterConfig: map[string]*anypb.Any{
			FilterNameCORS: MustAny(CorsPolicy()),
		},
	}

	listener := &listener.Listener{
		Name: listenerName,
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
		FilterChains: []*listener.FilterChain{
			{
				Filters: []*listener.Filter{
					{
						Name: FilterNameHttpConnectionManager,
						ConfigType: &listener.Filter_TypedConfig{
							TypedConfig: MustAny(manager),
						},
					},
				},
			},
		},
	}

	rawSnapshot := map[resource.Type][]types.Resource{
		resource.ClusterType: castResources(
			slices.Concat(
				s.Postgrest.Cluster(instance),
				s.GoTrue.Cluster(instance),
				s.PGMeta.Cluster(instance),
			)...),
		resource.RouteType:    {routeCfg},
		resource.ListenerType: {listener},
	}

	snapshot, err := cache.NewSnapshot(
		version,
		rawSnapshot,
	)
	if err != nil {
		return nil, err
	}

	if err := snapshot.Consistent(); err != nil {
		return nil, err
	}

	return snapshot, nil
}

func castResources[T types.Resource](from ...T) []types.Resource {
	result := make([]types.Resource, len(from))
	for idx := range from {
		result[idx] = from[idx]
	}

	return result
}
