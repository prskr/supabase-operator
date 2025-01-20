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
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	rbacv3cfg "github.com/envoyproxy/go-control-plane/envoy/config/rbac/v3"
	jwtauthnv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/jwt_authn/v3"
	rbacv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/rbac/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
)

const (
	JwtProviderName             = "supabase"
	JwtMetadataKey              = "supabase-jwt"
	JwtAuthenticatedRequirement = "supabase-jwt-authenticated"
)

type ForwardJwt bool

func (j ForwardJwt) Apply(opts *JwtOptions) {
	opts.ForwardJwt = bool(j)
}

type JwtFilterOption interface {
	Apply(opts *JwtOptions)
}

type JwtFilterOptionFunc func(opts *JwtOptions)

func (f JwtFilterOptionFunc) Apply(opts *JwtOptions) {
	f(opts)
}

type JwtOptions struct {
	ForwardJwt    bool
	ForwardHeader string
}

func JWTPerRouteConfig() *jwtauthnv3.PerRouteConfig {
	return &jwtauthnv3.PerRouteConfig{
		RequirementSpecifier: &jwtauthnv3.PerRouteConfig_RequirementName{
			RequirementName: JwtAuthenticatedRequirement,
		},
	}
}

func JWTAllowAll() *jwtauthnv3.PerRouteConfig {
	return &jwtauthnv3.PerRouteConfig{
		RequirementSpecifier: &jwtauthnv3.PerRouteConfig_Disabled{
			Disabled: true,
		},
	}
}

func JWTFilterConfig(opts ...JwtFilterOption) *jwtauthnv3.JwtAuthentication {
	const (
		issuerName             = "supabase"
		bearerTokenPrefix      = "Bearer "
		apiKeyParamKey         = "apikey"
		authorizationHeaderKey = "Authorization"
	)

	filterOpts := &JwtOptions{
		ForwardJwt: true,
	}

	for _, o := range opts {
		o.Apply(filterOpts)
	}

	return &jwtauthnv3.JwtAuthentication{
		Providers: map[string]*jwtauthnv3.JwtProvider{
			JwtProviderName: {
				Issuer:            issuerName,
				PayloadInMetadata: JwtMetadataKey,
				JwksSourceSpecifier: &jwtauthnv3.JwtProvider_LocalJwks{
					LocalJwks: &corev3.DataSource{
						Specifier: &corev3.DataSource_Filename{
							Filename: "/etc/envoy/jwks.json",
						},
						WatchedDirectory: &corev3.WatchedDirectory{
							Path: "/etc/envoy",
						},
					},
				},
				Forward: filterOpts.ForwardJwt,
				FromHeaders: []*jwtauthnv3.JwtHeader{
					{
						Name: apiKeyParamKey,
					},
					{
						Name:        authorizationHeaderKey,
						ValuePrefix: bearerTokenPrefix,
					},
				},
				FromParams:        []string{apiKeyParamKey},
				RequireExpiration: true,
			},
		},
		BypassCorsPreflight: true,
		RequirementMap: map[string]*jwtauthnv3.JwtRequirement{
			JwtAuthenticatedRequirement: {
				RequiresType: &jwtauthnv3.JwtRequirement_ProviderName{
					ProviderName: JwtProviderName,
				},
			},
		},
	}
}

func RBACPerRoute(cfg *rbacv3.RBAC) *rbacv3.RBACPerRoute {
	return &rbacv3.RBACPerRoute{Rbac: cfg}
}

func RBACAllowAllConfig() *rbacv3.RBAC {
	return &rbacv3.RBAC{
		Rules: &rbacv3cfg.RBAC{
			Action: rbacv3cfg.RBAC_ALLOW,
			Policies: map[string]*rbacv3cfg.Policy{
				"Allow anyone": {
					Permissions: []*rbacv3cfg.Permission{{
						Rule: &rbacv3cfg.Permission_Any{Any: true},
					}},
					Principals: []*rbacv3cfg.Principal{{
						Identifier: &rbacv3cfg.Principal_Any{
							Any: true,
						},
					}},
				},
			},
		},
	}
}

func RBACRequireAuthConfig() *rbacv3.RBAC {
	return &rbacv3.RBAC{
		Rules: &rbacv3cfg.RBAC{
			Action: rbacv3cfg.RBAC_ALLOW,
			Policies: map[string]*rbacv3cfg.Policy{
				"allow admin and anon roles": {
					Permissions: []*rbacv3cfg.Permission{{
						Rule: &rbacv3cfg.Permission_Any{Any: true},
					}},
					Principals: []*rbacv3cfg.Principal{{
						Identifier: &rbacv3cfg.Principal_OrIds{
							OrIds: &rbacv3cfg.Principal_Set{
								Ids: []*rbacv3cfg.Principal{
									{
										Identifier: &rbacv3cfg.Principal_Metadata{
											Metadata: &matcherv3.MetadataMatcher{
												Filter: FilterNameJwtAuthn,
												Path: []*matcherv3.MetadataMatcher_PathSegment{
													{
														Segment: &matcherv3.MetadataMatcher_PathSegment_Key{
															Key: "jwt_payload",
														},
													},
													{
														Segment: &matcherv3.MetadataMatcher_PathSegment_Key{
															Key: "role",
														},
													},
												},
												Value: &matcherv3.ValueMatcher{
													MatchPattern: &matcherv3.ValueMatcher_OrMatch{
														OrMatch: &matcherv3.OrMatcher{
															ValueMatchers: []*matcherv3.ValueMatcher{
																{
																	MatchPattern: &matcherv3.ValueMatcher_StringMatch{
																		StringMatch: &matcherv3.StringMatcher{
																			MatchPattern: &matcherv3.StringMatcher_Exact{
																				Exact: "anon",
																			},
																		},
																	},
																},
																{
																	MatchPattern: &matcherv3.ValueMatcher_StringMatch{
																		StringMatch: &matcherv3.StringMatcher{
																			MatchPattern: &matcherv3.StringMatcher_Exact{
																				Exact: "authenticated",
																			},
																		},
																	},
																},
																{
																	MatchPattern: &matcherv3.ValueMatcher_StringMatch{
																		StringMatch: &matcherv3.StringMatcher{
																			MatchPattern: &matcherv3.StringMatcher_Exact{
																				Exact: "admin",
																			},
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					}},
				},
			},
		},
	}
}
