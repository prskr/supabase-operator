package supabase

type ImageRef struct {
	// The repository of the image
	Repository string
	// The tag of the image
	Tag string
}

var Images = map[string]ImageRef{
	"analytics": {
		Repository: "supabase/logflare",
		Tag:        "1.4.0",
	},
	"auth": {
		Repository: "supabase/gotrue",
		Tag:        "v2.164.0",
	},
	"db": {
		Repository: "supabase/postgres",
		Tag:        "15.6.1.146",
	},
	"functions": {
		Repository: "supabase/edge-runtime",
		Tag:        "v1.65.3",
	},
	"imgproxy": {
		Repository: "darthsim/imgproxy",
		Tag:        "v3.8.0",
	},
	"kong": {
		Repository: "kong",
		Tag:        "2.8.1",
	},
	"meta": {
		Repository: "supabase/postgres-meta",
		Tag:        "v0.84.2",
	},
	"realtime": {
		Repository: "supabase/realtime",
		Tag:        "v2.33.58",
	},
	"rest": {
		Repository: "postgrest/postgrest",
		Tag:        "v12.2.0",
	},
	"storage": {
		Repository: "supabase/storage-api",
		Tag:        "v1.11.13",
	},
	"studio": {
		Repository: "supabase/studio",
		Tag:        "20241202-71e5240",
	},
	"supavisor": {
		Repository: "supabase/supavisor",
		Tag:        "1.1.56",
	},
	"vector": {
		Repository: "timberio/vector",
		Tag:        "0.28.1-alpine",
	},
}
