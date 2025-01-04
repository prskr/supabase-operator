package controlplane

const (
	FilterNameJwtAuthn              = "envoy.filters.http.jwt_authn"
	FilterNameRBAC                  = "envoy.filters.http.rbac"
	FilterNameCORS                  = "envoy.filters.http.cors"
	FilterNameHttpRouter            = "envoy.filters.http.router"
	FilterNameHttpConnectionManager = "envoy.filters.network.http_connection_manager"
)
