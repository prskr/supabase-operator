package controlplane

import (
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	corsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/cors/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
)

func Cors() *corsv3.Cors {
	return new(corsv3.Cors)
}

func CorsPolicy() *corsv3.CorsPolicy {
	return &corsv3.CorsPolicy{
		AllowMethods: "*",
		AllowHeaders: "*",
		AllowOriginStringMatch: []*matcherv3.StringMatcher{{
			MatchPattern: &matcherv3.StringMatcher_SafeRegex{
				SafeRegex: &matcherv3.RegexMatcher{
					Regex: `\*`,
				},
			},
		}},
		FilterEnabled: &corev3.RuntimeFractionalPercent{
			DefaultValue: &typev3.FractionalPercent{
				Numerator:   100,
				Denominator: typev3.FractionalPercent_HUNDRED,
			},
		},
	}
}
