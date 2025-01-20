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
