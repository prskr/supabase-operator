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

package meta

import supabasev1alpha1 "code.icb4dc0.de/prskr/supabase-operator/api/v1alpha1"

const (
	WellKnownMetaPrefix = "app.kubernetes.io/"
)

var WellKnownLabel = struct {
	Name      string
	Instance  string
	PartOf    string
	Version   string
	Component string
	ManagedBy string
}{
	Name:      WellKnownMetaPrefix + "name",
	Instance:  WellKnownMetaPrefix + "instance",
	PartOf:    WellKnownMetaPrefix + "part-of",
	Version:   WellKnownMetaPrefix + "version",
	Component: WellKnownMetaPrefix + "component",
	ManagedBy: WellKnownMetaPrefix + "managed-by",
}

var SupabaseLabel = struct {
	Reload           string
	ApiGatewayTarget string
}{
	Reload:           supabasev1alpha1.GroupVersion.Group + "/reload",
	ApiGatewayTarget: supabasev1alpha1.GroupVersion.Group + "/api-gateway-target",
}
