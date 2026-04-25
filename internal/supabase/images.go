/*
Copyright 2026 Peter Kurfer.

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
		Tag:        "v1.70.3",
	},
	Envoy: ImageRef{
		Repository: "envoyproxy/envoy",
		Tag:        "distroless-v1.37.1",
	},
	Gotrue: ImageRef{
		Repository: "supabase/gotrue",
		Tag:        "v2.186.0",
	},
	ImgProxy: ImageRef{
		Repository: "darthsim/imgproxy",
		Tag:        "v3.30.1",
	},
	PostgresMeta: ImageRef{
		Repository: "supabase/postgres-meta",
		Tag:        "v0.95.2",
	},
	Postgrest: ImageRef{
		Repository: "postgrest/postgrest",
		Tag:        "v14.5",
	},
	Realtime: ImageRef{
		Repository: "supabase/realtime",
		Tag:        "v2.76.5",
	},
	Storage: ImageRef{
		Repository: "supabase/storage-api",
		Tag:        "v1.37.8",
	},
	Studio: ImageRef{
		Repository: "supabase/studio",
		Tag:        "2026.02.16-sha-26c615c",
	},
}
