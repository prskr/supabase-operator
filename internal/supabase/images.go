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

package supabase

import "fmt"

type ImageRef struct {
	// The repository of the image
	Repository string
	// The tag of the image
	Tag string
}

func (r ImageRef) String() string {
	return fmt.Sprintf("%s:%s", r.Repository, r.Tag)
}

var Images = struct {
	EdgeRuntime  ImageRef
	Envoy        ImageRef
	Gotrue       ImageRef
	ImgProxy     ImageRef
	PostgresMeta ImageRef
	Postgrest    ImageRef
	Realtime     ImageRef
	Storage      ImageRef
	Studio       ImageRef
}{
	EdgeRuntime: ImageRef{
		Repository: "supabase/edge-runtime",
		Tag:        "v1.66.4",
	},
	Envoy: ImageRef{
		Repository: "envoyproxy/envoy",
		Tag:        "distroless-v1.33.0",
	},
	Gotrue: ImageRef{
		Repository: "supabase/gotrue",
		Tag:        "v2.167.0",
	},
	ImgProxy: ImageRef{
		Repository: "darthsim/imgproxy",
		Tag:        "v3.8.0",
	},
	PostgresMeta: ImageRef{
		Repository: "supabase/postgres-meta",
		Tag:        "v0.84.2",
	},
	Postgrest: ImageRef{
		Repository: "postgrest/postgrest",
		Tag:        "v12.2.0",
	},
	Realtime: ImageRef{
		Repository: "supabase/realtime",
		Tag:        "v2.33.70",
	},
	Storage: ImageRef{
		Repository: "supabase/storage-api",
		Tag:        "v1.14.5",
	},
	Studio: ImageRef{
		Repository: "supabase/studio",
		Tag:        "20250113-83c9420",
	},
}
