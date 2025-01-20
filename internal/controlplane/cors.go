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
