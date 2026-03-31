# API Reference

## Packages
- [supabase.k8s.icb4dc0.de/v1alpha1](#supabasek8sicb4dc0dev1alpha1)


## supabase.k8s.icb4dc0.de/v1alpha1

Package v1alpha1 contains API Schema definitions for the supabase v1alpha1 API group.

### Resource Types
- [APIGateway](#apigateway)
- [APIGatewayList](#apigatewaylist)
- [Core](#core)
- [CoreList](#corelist)
- [Dashboard](#dashboard)
- [DashboardList](#dashboardlist)
- [Storage](#storage)
- [StorageList](#storagelist)



#### AISpec







_Appears in:_
- [StudioSpec](#studiospec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `openai` _[OpenAISpec](#openaispec)_ |  |  |  |


#### APIGateway



APIGateway is the Schema for the apigateways API.



_Appears in:_
- [APIGatewayList](#apigatewaylist)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `supabase.k8s.icb4dc0.de/v1alpha1` | | |
| `kind` _string_ | `APIGateway` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[APIGatewaySpec](#apigatewayspec)_ |  |  |  |


#### APIGatewayList



APIGatewayList contains a list of APIGateway.





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `supabase.k8s.icb4dc0.de/v1alpha1` | | |
| `kind` _string_ | `APIGatewayList` | | |
| `metadata` _[ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#listmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `items` _[APIGateway](#apigateway) array_ |  |  |  |


#### APIGatewaySpec



APIGatewaySpec defines the desired state of APIGateway.



_Appears in:_
- [APIGateway](#apigateway)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `envoy` _[EnvoySpec](#envoyspec)_ | Envoy - configure the envoy instance and most importantly the control-plane |  |  |
| `apiEndpoint` _[ApiEndpointSpec](#apiendpointspec)_ | ApiEndpoint - Configure the endpoint for all API routes<br />this includes the JWT configuration |  |  |
| `dashboardEndpoint` _[DashboardEndpointSpec](#dashboardendpointspec)_ | DashboardEndpoint - Configure the endpoint for the Supabase dashboard (studio)<br />this includes optional authentication (basic or Oauth2) for the dashboard |  |  |
| `serviceSelector` _[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#labelselector-v1-meta)_ | ServiceSelector - selector to match all Supabase services (or in fact EndpointSlices) that should be considered for this APIGateway | \{ matchExpressions:[map[key:app.kubernetes.io/part-of operator:In values:[supabase]] map[key:supabase.k8s.icb4dc0.de/api-gateway-target operator:Exists]] \} |  |
| `componentTypeLabel` _string_ | ComponentTypeLabel - Label to identify which Supabase component a Service represents (e.g. auth, postgrest, ...) | app.kubernetes.io/name |  |




#### ApiEndpointSpec







_Appears in:_
- [APIGatewaySpec](#apigatewayspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `jwks` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#secretkeyselector-v1-core)_ | JWKSSelector - selector where the JWKS can be retrieved from to enable the API gateway to validate JWTs |  |  |
| `tls` _[EndpointTlsSpec](#endpointtlsspec)_ | TLS - enable and configure TLS for the API endpoint |  |  |


#### AuthProviderMeta







_Appears in:_
- [AzureAuthProvider](#azureauthprovider)
- [EmailAuthProvider](#emailauthprovider)
- [GithubAuthProvider](#githubauthprovider)
- [PhoneAuthProvider](#phoneauthprovider)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `enabled` _boolean_ | Enabled - whether the authentication provider is enabled or not | true |  |


#### AuthProviders







_Appears in:_
- [AuthSpec](#authspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `email` _[EmailAuthProvider](#emailauthprovider)_ |  |  |  |
| `azure` _[AzureAuthProvider](#azureauthprovider)_ |  |  |  |
| `github` _[GithubAuthProvider](#githubauthprovider)_ |  |  |  |
| `phone` _[PhoneAuthProvider](#phoneauthprovider)_ |  |  |  |


#### AuthSpec







_Appears in:_
- [CoreSpec](#corespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `additionalRedirectUrls` _string array_ |  |  |  |
| `disableSignup` _boolean_ |  |  |  |
| `anonymousUsersEnabled` _boolean_ |  |  |  |
| `providers` _[AuthProviders](#authproviders)_ |  |  |  |
| `workloadTemplate` _[WorkloadSpec](#workloadspec)_ |  |  |  |
| `emailSignupDisabled` _boolean_ |  |  |  |
| `observability` _[GoTrueObservabilitySpec](#gotrueobservabilityspec)_ |  |  |  |


#### AzureAuthProvider







_Appears in:_
- [AuthProviders](#authproviders)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `enabled` _boolean_ | Enabled - whether the authentication provider is enabled or not | true |  |
| `clientID` _string_ | ClientID - OAuth2 client ID |  |  |
| `clientSecretRef` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#secretkeyselector-v1-core)_ | ClientSecretRef - reference to a secret containing the OAuth2 client secret in the same namespace as the `Core` resource |  |  |
| `url` _string_ | URL - OAuth2 provider URL. Only necessary for Azure, Gitlab and Keycloak where custom base URLs are used |  |  |




#### ContainerTemplate







_Appears in:_
- [StorageWorkloadSpec](#storageworkloadspec)
- [WorkloadSpec](#workloadspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `image` _string_ |  |  |  |
| `pullPolicy` _[PullPolicy](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#pullpolicy-v1-core)_ |  |  |  |
| `imagePullSecrets` _[LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#localobjectreference-v1-core) array_ |  |  |  |
| `securityContext` _[SecurityContext](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#securitycontext-v1-core)_ | SecurityContext - override the container SecurityContext<br />use with caution, by default the operator already uses sane defaults |  |  |
| `resources` _[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#resourcerequirements-v1-core)_ |  |  |  |
| `volumeMounts` _[VolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#volumemount-v1-core) array_ |  |  |  |
| `additionalEnv` _[EnvVar](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#envvar-v1-core) array_ |  |  |  |


#### ControlPlaneSpec







_Appears in:_
- [EnvoySpec](#envoyspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `host` _string_ | Host is the hostname of the envoy control plane endpoint |  |  |
| `port` _integer_ | Port is the port number of the envoy control plane endpoint - typically this is 18000 | 18000 | Maximum: 65535 <br /> |


#### Core



Core is the Schema for the cores API.



_Appears in:_
- [CoreList](#corelist)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `supabase.k8s.icb4dc0.de/v1alpha1` | | |
| `kind` _string_ | `Core` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[CoreSpec](#corespec)_ |  |  |  |




#### CoreJwtSpec







_Appears in:_
- [CoreSpec](#corespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `secretName` _string_ | SecretRef - object reference to the Secret where JWT values are stored |  |  |
| `secretKey` _string_ | SecretKey - key in secret where to read the JWT HMAC secret from | secret |  |
| `jwksKey` _string_ | JwksKey - key in secret where to read the JWKS from | jwks.json |  |
| `anonKey` _string_ | AnonKey - key in secret where to read the anon JWT from | anon_key |  |
| `serviceKey` _string_ | ServiceKey - key in secret where to read the service JWT from | service_key |  |
| `secret` _string_ | Secret - JWT HMAC secret in plain text<br />This is WRITE-ONLY and will be copied to the SecretRef by the defaulter |  |  |
| `expiry` _integer_ | Expiry - expiration time in seconds for JWTs | 3600 |  |


#### CoreList



CoreList contains a list of Core.





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `supabase.k8s.icb4dc0.de/v1alpha1` | | |
| `kind` _string_ | `CoreList` | | |
| `metadata` _[ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#listmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `items` _[Core](#core) array_ |  |  |  |


#### CoreSpec



CoreSpec defines the desired state of Core.



_Appears in:_
- [Core](#core)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `externalUrl` _string_ | APIExternalURL is referring to the URL where Supabase API will be available<br />Typically this is the ingress of the API gateway |  |  |
| `siteUrl` _string_ | SiteURL is referring to the URL of the (frontend) application<br />In most Kubernetes scenarios this is the same as the APIExternalURL with a different path handler in the ingress |  |  |
| `jwt` _[CoreJwtSpec](#corejwtspec)_ |  |  |  |
| `database` _[Database](#database)_ |  |  |  |
| `postgrest` _[PostgrestSpec](#postgrestspec)_ |  |  |  |
| `auth` _[AuthSpec](#authspec)_ |  |  |  |




#### DBCredentialsReference







_Appears in:_
- [DashboardDBSpec](#dashboarddbspec)
- [StorageAPIDBSpec](#storageapidbspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `secretName` _string_ |  |  |  |
| `usernameKey` _string_ | UsernameKey | username |  |
| `passwordKey` _string_ | PasswordKey | password |  |


#### Dashboard



Dashboard is the Schema for the dashboards API.



_Appears in:_
- [DashboardList](#dashboardlist)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `supabase.k8s.icb4dc0.de/v1alpha1` | | |
| `kind` _string_ | `Dashboard` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[DashboardSpec](#dashboardspec)_ |  |  |  |


#### DashboardAuthSpec







_Appears in:_
- [DashboardEndpointSpec](#dashboardendpointspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `oauth2` _[DashboardOAuth2Spec](#dashboardoauth2spec)_ | OAuth2 - configure oauth2 authentication for the dashhboard listener<br />if configured, will be preferred over Basic authentication configuration<br />effectively disabling basic auth |  |  |
| `basic` _[DashboardBasicAuthSpec](#dashboardbasicauthspec)_ | Basic - HTTP basic auth configuration, this should only be used in exceptions<br />e.g. during evaluations or for local development<br />only used if no other authentication is configured |  |  |




#### DashboardBasicAuthSpec







_Appears in:_
- [DashboardAuthSpec](#dashboardauthspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `usersInline` _string array_ | UsersInline - [htpasswd format](https://httpd.apache.org/docs/2.4/programs/htpasswd.html) |  | items:Pattern: ^[\w_.]+:\\{SHA\\}[A-z0-9]+=*$ <br /> |
| `plaintextUsersSecretRef` _string_ | PlaintextUsersSecretRef - name of a secret that contains plaintext credentials in key-value form<br />if not empty, credentials will be merged with inline users |  |  |


#### DashboardDBSpec







_Appears in:_
- [DashboardSpec](#dashboardspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `host` _string_ |  |  |  |
| `port` _integer_ | Port - Database port, typically 5432 | 5432 |  |
| `dbName` _string_ |  |  |  |
| `dbCredentialsRef` _[DBCredentialsReference](#dbcredentialsreference)_ | DBCredentialsRef - reference to a Secret key where the DB credentials can be retrieved from<br />Credentials need to be stored in basic auth form |  |  |
| `schemas` _string array_ | Schemas - schema where PostgREST is looking for objects (tables, views, functions, ...) | [public storage graphql_public] |  |
| `extraSearchPath` _string array_ | ExtraSearchPath - Extra schemas to add to the search_path of every request.<br />These schemas tables, views and functions don’t get API endpoints, they can only be referred from the database objects inside your db-schemas. | [public] |  |


#### DashboardEndpointSpec







_Appears in:_
- [APIGatewaySpec](#apigatewayspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `auth` _[DashboardAuthSpec](#dashboardauthspec)_ | Auth - configure authentication for the dashboard endpoint |  |  |
| `tls` _[EndpointTlsSpec](#endpointtlsspec)_ | TLS - enable and configure TLS for the Dashboard endpoint |  |  |


#### DashboardList



DashboardList contains a list of Dashboard.





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `supabase.k8s.icb4dc0.de/v1alpha1` | | |
| `kind` _string_ | `DashboardList` | | |
| `metadata` _[ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#listmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `items` _[Dashboard](#dashboard) array_ |  |  |  |


#### DashboardOAuth2Spec







_Appears in:_
- [DashboardAuthSpec](#dashboardauthspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `openIdIssuer` _string_ | OpenIDIssuer - if set the defaulter will fetch the discovery document and fill<br />TokenEndpoint and AuthorizationEndpoint based on the discovery document |  |  |
| `tokenEndpoint` _string_ | TokenEndpoint - endpoint where Envoy will retrieve the OAuth2 access and identity token from |  |  |
| `authorizationEndpoint` _string_ | AuthorizationEndpoint - endpoint where the user will be redirected to authenticate |  |  |
| `clientId` _string_ | ClientID - client ID to authenticate with the OAuth2 provider |  |  |
| `scopes` _string array_ | Scopes - scopes to request from the OAuth2 provider (e.g. "openid", "profile", ...) - optional |  |  |
| `resources` _string array_ | Resources - resources to request from the OAuth2 provider (e.g. "user", "email", ...) - optional |  |  |
| `clientSecretRef` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#secretkeyselector-v1-core)_ | ClientSecretRef - reference to the secret that contains the client secret |  |  |


#### DashboardSpec



DashboardSpec defines the desired state of Dashboard.



_Appears in:_
- [Dashboard](#dashboard)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `db` _[DashboardDBSpec](#dashboarddbspec)_ |  |  |  |
| `pgMeta` _[PGMetaSpec](#pgmetaspec)_ | PGMeta |  |  |
| `studio` _[StudioSpec](#studiospec)_ | Studio |  |  |




#### Database







_Appears in:_
- [CoreSpec](#corespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `dsn` _string_ |  |  |  |
| `dsnSecretRef` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#secretkeyselector-v1-core)_ |  |  |  |
| `roles` _[DatabaseRoles](#databaseroles)_ |  |  |  |


#### DatabaseRoles







_Appears in:_
- [Database](#database)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `selfManaged` _boolean_ | SelfManaged - whether the database roles are managed externally<br />when enabled the operator does not attempt to create secrets, generate passwords or whatsoever for all database roles<br />i.e. all secrets need to be provided or the instance won't work |  |  |
| `secrets` _[DatabaseRolesSecrets](#databaserolessecrets)_ | Secrets - typed 'map' of secrets for each database role that Supabase needs |  |  |


#### DatabaseRolesSecrets







_Appears in:_
- [DatabaseRoles](#databaseroles)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `supabaseAdmin` _string_ |  |  |  |
| `authenticator` _string_ |  |  |  |
| `supabaseAuthAdmin` _string_ |  |  |  |
| `supabaseFunctionsAdmin` _string_ |  |  |  |
| `supabaseStorageAdmin` _string_ |  |  |  |


#### DatabaseStatus







_Appears in:_
- [CoreStatus](#corestatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `migrationConditions` _[MigrationScriptCondition](#migrationscriptcondition) array_ |  |  |  |
| `roles` _object (keys:string, values:integer array)_ |  |  |  |


#### EmailAuthProvider







_Appears in:_
- [AuthProviders](#authproviders)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `enabled` _boolean_ | Enabled - whether the authentication provider is enabled or not | true |  |
| `adminEmail` _string_ |  |  |  |
| `senderName` _string_ |  |  |  |
| `autoconfirmEmail` _boolean_ |  |  |  |
| `subjectsInvite` _string_ |  |  |  |
| `subjectsConfirmation` _string_ |  |  |  |
| `smtpSpec` _[EmailAuthSMTPSpec](#emailauthsmtpspec)_ |  |  |  |


#### EmailAuthSMTPSpec







_Appears in:_
- [EmailAuthProvider](#emailauthprovider)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `host` _string_ |  |  |  |
| `port` _integer_ |  |  |  |
| `maxFrequency` _[uint](#uint)_ |  |  |  |
| `credentialsRef` _[SMTPCredentialsReference](#smtpcredentialsreference)_ |  |  |  |


#### EndpointTlsSpec







_Appears in:_
- [ApiEndpointSpec](#apiendpointspec)
- [DashboardEndpointSpec](#dashboardendpointspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `cert` _[TlsCertRef](#tlscertref)_ |  |  |  |


#### EnvoyComponentLogLevel







_Appears in:_
- [EnvoyDebuggingOptions](#envoydebuggingoptions)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `component` _string_ | Component - the component to set the log level for<br />the component IDs can be found [here](https://github.com/envoyproxy/envoy/blob/main/source/common/common/logger.h#L36) |  |  |
| `level` _[EnvoyLogLevel](#envoyloglevel)_ | Level - the log level to set for the component |  | Enum: [trace debug info warning error critical off] <br /> |


#### EnvoyDebuggingOptions







_Appears in:_
- [EnvoySpec](#envoyspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `componentLogLevels` _[EnvoyComponentLogLevel](#envoycomponentloglevel) array_ |  |  |  |


#### EnvoyLogLevel

_Underlying type:_ _string_





_Appears in:_
- [EnvoyComponentLogLevel](#envoycomponentloglevel)



#### EnvoyMetricsSpec







_Appears in:_
- [EnvoyObservabilitySpec](#envoyobservabilityspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `enabled` _boolean_ |  |  |  |


#### EnvoyObservabilitySpec







_Appears in:_
- [EnvoySpec](#envoyspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `metrics` _[EnvoyMetricsSpec](#envoymetricsspec)_ |  |  |  |
| `traces` _[EnvoyTracesSpec](#envoytracesspec)_ |  |  |  |


#### EnvoySpec







_Appears in:_
- [APIGatewaySpec](#apigatewayspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `nodeName` _string_ | NodeName - identifies the Envoy cluster within the current namespace<br />if not set, the name of the APIGateway resource will be used<br />The primary use case is to make the assignment of multiple supabase instances in a single namespace explicit. |  |  |
| `controlPlane` _[ControlPlaneSpec](#controlplanespec)_ | ControlPlane - configure the control plane where Envoy will retrieve its configuration from |  |  |
| `workloadSpec` _[WorkloadSpec](#workloadspec)_ | WorkloadTemplate - customize the Envoy deployment |  |  |
| `disableIPv6` _boolean_ | DisableIPv6 - disable IPv6 for the Envoy instance<br />this will force Envoy to use IPv4 for upstream hosts (mostly for the OAuth2 token endpoint) |  |  |
| `debugging` _[EnvoyDebuggingOptions](#envoydebuggingoptions)_ |  |  |  |
| `observability` _[EnvoyObservabilitySpec](#envoyobservabilityspec)_ |  |  |  |


#### EnvoyStatus







_Appears in:_
- [APIGatewayStatus](#apigatewaystatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `resourceHash` _integer array_ |  |  |  |


#### EnvoyTracesSpec







_Appears in:_
- [EnvoyObservabilitySpec](#envoyobservabilityspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `otel` _[EnvoyTracingOTELSpec](#envoytracingotelspec)_ |  |  |  |


#### EnvoyTracingOTELSpec







_Appears in:_
- [EnvoyTracesSpec](#envoytracesspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `serviceName` _string_ | The name for the service. This will be populated in the ResourceSpan Resource attributes. If it is not provided, it will default to “<name-of-the-api-gateway>:envoy”. |  |  |
| `endpoint` _string_ |  |  |  |


#### FileBackendSpec







_Appears in:_
- [StorageAPISpec](#storageapispec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `path` _string_ | Path - path to where files will be stored |  |  |


#### GithubAuthProvider







_Appears in:_
- [AuthProviders](#authproviders)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `enabled` _boolean_ | Enabled - whether the authentication provider is enabled or not | true |  |
| `clientID` _string_ | ClientID - OAuth2 client ID |  |  |
| `clientSecretRef` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#secretkeyselector-v1-core)_ | ClientSecretRef - reference to a secret containing the OAuth2 client secret in the same namespace as the `Core` resource |  |  |
| `url` _string_ | URL - OAuth2 provider URL. Only necessary for Azure, Gitlab and Keycloak where custom base URLs are used |  |  |


#### GoTrueMetricsSpec







_Appears in:_
- [GoTrueObservabilitySpec](#gotrueobservabilityspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `enabled` _boolean_ |  |  |  |


#### GoTrueObservabilitySpec







_Appears in:_
- [AuthSpec](#authspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `metrics` _[GoTrueMetricsSpec](#gotruemetricsspec)_ |  |  |  |
| `tracing` _[GoTrueTracingSpec](#gotruetracingspec)_ |  |  |  |


#### GoTrueTracingSpec







_Appears in:_
- [GoTrueObservabilitySpec](#gotrueobservabilityspec)



#### ImageProxySpec







_Appears in:_
- [StorageSpec](#storagespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `enable` _boolean_ | Enable - whether to deploy the image proxy or not |  |  |
| `enableWebPDetection` _boolean_ |  |  |  |
| `workloadSpec` _[WorkloadSpec](#workloadspec)_ | WorkloadTemplate - customize the image proxy workload |  |  |


#### ImageSpec







_Appears in:_
- [ContainerTemplate](#containertemplate)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `image` _string_ |  |  |  |
| `pullPolicy` _[PullPolicy](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#pullpolicy-v1-core)_ |  |  |  |


#### JwtSpec







_Appears in:_
- [CoreJwtSpec](#corejwtspec)
- [StorageAPISpec](#storageapispec)
- [StudioSpec](#studiospec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `secretName` _string_ | SecretRef - object reference to the Secret where JWT values are stored |  |  |
| `secretKey` _string_ | SecretKey - key in secret where to read the JWT HMAC secret from | secret |  |
| `jwksKey` _string_ | JwksKey - key in secret where to read the JWKS from | jwks.json |  |
| `anonKey` _string_ | AnonKey - key in secret where to read the anon JWT from | anon_key |  |
| `serviceKey` _string_ | ServiceKey - key in secret where to read the service JWT from | service_key |  |




#### MigrationScriptCondition







_Appears in:_
- [DatabaseStatus](#databasestatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name - file name of the migration script |  |  |
| `hash` _integer array_ | Hash - SHA256 hash of the script when it was last successfully applied |  |  |
| `lastProbeTime` _[Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#time-v1-meta)_ | LastProbeTime - last time the operator tried to execute the migration script |  |  |
| `lastTransitionTime` _[Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#time-v1-meta)_ | LastTransitionTime - last time the condition transitioned from one status to another |  |  |
| `reason` _string_ | Reason - one-word, CamcelCase reason for the condition's last transition |  |  |
| `message` _string_ | Message - human-readable message indicating details about the last transition |  |  |


#### OAuthProvider







_Appears in:_
- [AzureAuthProvider](#azureauthprovider)
- [GithubAuthProvider](#githubauthprovider)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `clientID` _string_ | ClientID - OAuth2 client ID |  |  |
| `clientSecretRef` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#secretkeyselector-v1-core)_ | ClientSecretRef - reference to a secret containing the OAuth2 client secret in the same namespace as the `Core` resource |  |  |
| `url` _string_ | URL - OAuth2 provider URL. Only necessary for Azure, Gitlab and Keycloak where custom base URLs are used |  |  |


#### OpenAISpec







_Appears in:_
- [AISpec](#aispec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiKey` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#secretkeyselector-v1-core)_ |  |  |  |


#### PGMetaSpec







_Appears in:_
- [DashboardSpec](#dashboardspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `workloadSpec` _[WorkloadSpec](#workloadspec)_ | WorkloadTemplate - customize the pg-meta deployment |  |  |


#### PhoneAuthProvider







_Appears in:_
- [AuthProviders](#authproviders)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `enabled` _boolean_ | Enabled - whether the authentication provider is enabled or not | true |  |


#### PostgrestMetricsSpec







_Appears in:_
- [PostgrestObservabilitySpec](#postgrestobservabilityspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `enabled` _boolean_ |  |  |  |


#### PostgrestObservabilitySpec







_Appears in:_
- [PostgrestSpec](#postgrestspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `metrics` _[PostgrestMetricsSpec](#postgrestmetricsspec)_ |  |  |  |


#### PostgrestSpec







_Appears in:_
- [CoreSpec](#corespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `schemas` _string array_ | Schemas - schema where PostgREST is looking for objects (tables, views, functions, ...) | [public storage graphql_public] |  |
| `extraSearchPath` _string array_ | ExtraSearchPath - Extra schemas to add to the search_path of every request.<br />These schemas tables, views and functions don’t get API endpoints, they can only be referred from the database objects inside your db-schemas. | [public] |  |
| `anonRole` _string_ | AnonRole - name of the anon role | anon |  |
| `maxRows` _integer_ | MaxRows - maximum number of rows PostgREST will load at a time | 1000 |  |
| `workloadSpec` _[WorkloadSpec](#workloadspec)_ | WorkloadSpec - customize the PostgREST workload |  |  |
| `observability` _[PostgrestObservabilitySpec](#postgrestobservabilityspec)_ | ObservabilitySpec - customize the PostgREST observability |  |  |


#### S3BackendSpec







_Appears in:_
- [StorageAPISpec](#storageapispec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `region` _string_ | Region - S3 region of the backend |  |  |
| `endpoint` _string_ | Endpoint - hostname and port **with** http/https |  |  |
| `forcePathStyle` _boolean_ | ForcePathStyle - whether to use path style (e.g. for MinIO) or domain style<br />for bucket addressing |  |  |
| `bucket` _string_ | Bucket - bucke to use, if file backend is used, default value is sufficient | stub |  |
| `credentialsSecretRef` _[S3CredentialsRef](#s3credentialsref)_ | CredentialsSecretRef - reference to the Secret where access key id and access secret key are stored |  |  |


#### S3CredentialsRef







_Appears in:_
- [S3BackendSpec](#s3backendspec)
- [S3ProtocolSpec](#s3protocolspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `secretName` _string_ |  |  |  |
| `accessKeyIdKey` _string_ | AccessKeyIDKey - key in Secret where access key id will be referenced from | accessKeyId |  |
| `accessSecretKeyKey` _string_ | AccessSecretKeyKey - key in Secret where access secret key will be referenced from | secretAccessKey |  |


#### S3ProtocolSpec







_Appears in:_
- [StorageAPISpec](#storageapispec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `allowForwardedHeader` _boolean_ | AllowForwardedHeader | true |  |
| `credentialsSecretRef` _[S3CredentialsRef](#s3credentialsref)_ | CredentialsSecretRef - reference to the Secret where access key id and access secret key are stored |  |  |




#### Storage



Storage is the Schema for the storages API.



_Appears in:_
- [StorageList](#storagelist)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `supabase.k8s.icb4dc0.de/v1alpha1` | | |
| `kind` _string_ | `Storage` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[StorageSpec](#storagespec)_ |  |  |  |


#### StorageAPIDBSpec







_Appears in:_
- [StorageAPISpec](#storageapispec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `host` _string_ |  |  |  |
| `port` _integer_ | Port - Database port, typically 5432 | 5432 |  |
| `dbName` _string_ |  |  |  |
| `dbCredentialsRef` _[DBCredentialsReference](#dbcredentialsreference)_ | DBCredentialsRef - reference to a Secret key where the DB credentials can be retrieved from<br />Credentials need to be stored in basic auth form |  |  |


#### StorageAPISpec







_Appears in:_
- [StorageSpec](#storagespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `tenantId` _string_ |  | stub |  |
| `region` _string_ |  | stub |  |
| `s3Backend` _[S3BackendSpec](#s3backendspec)_ |  |  |  |
| `fileBackend` _[FileBackendSpec](#filebackendspec)_ | FileBackend - configure the file backend<br />either S3 or file backend **MUST** be configured |  |  |
| `fileSizeLimit` _integer_ | FileSizeLimit - maximum file upload size in bytes | 52428800 |  |
| `jwtAuth` _[JwtSpec](#jwtspec)_ | JwtAuth - Configure the JWT authentication parameters.<br />This includes where to retrieve anon and service key from as well as JWT secret and JWKS references<br />needed to validate JWTs send to the API |  |  |
| `db` _[StorageAPIDBSpec](#storageapidbspec)_ | DBSpec - Configure access to the Postgres database<br />In most cases this will reference the supabase-storage-admin credentials secret provided by the Core resource |  |  |
| `s3` _[S3ProtocolSpec](#s3protocolspec)_ | S3Protocol - Configure S3 access to the Storage API allowing clients to use any S3 client |  |  |
| `uploadTemp` _[UploadTempSpec](#uploadtempspec)_ | UploadTemp - configure the emptyDir for storing intermediate files during uploads |  |  |
| `workloadSpec` _[StorageWorkloadSpec](#storageworkloadspec)_ | WorkloadTemplate - customize the Storage API workload |  |  |
| `postgRESTServiceSelector` _object (keys:string, values:string)_ | PostgrestServiceSelector - selector to find the service for the PostgREST API<br />Required to configure the API URL in the studio deployment<br />If you don't run multiple PostgREST instances in the same namespaces, the default will be fine | \{ app.kubernetes.io/component:core app.kubernetes.io/name:postgrest \} |  |


#### StorageList



StorageList contains a list of Storage.





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `supabase.k8s.icb4dc0.de/v1alpha1` | | |
| `kind` _string_ | `StorageList` | | |
| `metadata` _[ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#listmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `items` _[Storage](#storage) array_ |  |  |  |


#### StorageSpec



StorageSpec defines the desired state of Storage.



_Appears in:_
- [Storage](#storage)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `api` _[StorageAPISpec](#storageapispec)_ | API - configure the Storage API |  |  |
| `imageProxy` _[ImageProxySpec](#imageproxyspec)_ | ImageProxy - optionally enable and configure the image proxy<br />the image proxy scale images to lower resolutions on demand to reduce traffic for instance for mobile devices |  |  |




#### StorageWorkloadSpec







_Appears in:_
- [StorageAPISpec](#storageapispec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `replicas` _integer_ |  |  |  |
| `securityContext` _[PodSecurityContext](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#podsecuritycontext-v1-core)_ |  |  |  |
| `additionalLabels` _object (keys:string, values:string)_ |  |  |  |
| `container` _[ContainerTemplate](#containertemplate)_ | ContainerSpec - customize the container template of the workload |  |  |
| `additionalVolumes` _[Volume](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#volume-v1-core) array_ |  |  |  |
| `strategy` _[DeploymentStrategy](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#deploymentstrategy-v1-apps)_ |  |  |  |


#### StudioSpec







_Appears in:_
- [DashboardSpec](#dashboardspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `ai` _[AISpec](#aispec)_ | AI / LLM integration<br />configure OpenAI API key and optionally base URL |  |  |
| `jwt` _[JwtSpec](#jwtspec)_ |  |  |  |
| `workloadSpec` _[WorkloadSpec](#workloadspec)_ | WorkloadTemplate - customize the studio deployment |  |  |
| `gatewayServiceSelector` _object (keys:string, values:string)_ | GatewayServiceSelector - selector to find the service for the API gateway<br />Required to configure the API URL in the studio deployment<br />If you don't run multiple APIGateway instances in the same namespaces, the default will be fine | \{ app.kubernetes.io/component:api-gateway app.kubernetes.io/name:envoy \} |  |
| `externalUrl` _string_ | APIExternalURL is referring to the URL where Supabase API will be available<br />Typically this is the ingress of the API gateway |  |  |


#### TlsCertRef







_Appears in:_
- [EndpointTlsSpec](#endpointtlsspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `secretName` _string_ |  |  |  |
| `serverCertKey` _string_ | ServerCertKey - key in the secret that contains the server certificate | tls.crt |  |
| `serverKeyKey` _string_ | ServerKeyKey - key in the secret that contains the server private key | tls.key |  |
| `caCertKey` _string_ | CaCertKey - key in the secret that contains the CA certificate | ca.crt |  |


#### UploadTempSpec







_Appears in:_
- [StorageAPISpec](#storageapispec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `medium` _[StorageMedium](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#storagemedium-v1-core)_ | Medium of the empty dir to cache uploads |  |  |
| `sizeLimit` _[Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#quantity-resource-api)_ |  |  |  |


#### WorkloadSpec







_Appears in:_
- [AuthSpec](#authspec)
- [EnvoySpec](#envoyspec)
- [ImageProxySpec](#imageproxyspec)
- [PGMetaSpec](#pgmetaspec)
- [PostgrestSpec](#postgrestspec)
- [StorageWorkloadSpec](#storageworkloadspec)
- [StudioSpec](#studiospec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `replicas` _integer_ |  |  |  |
| `securityContext` _[PodSecurityContext](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#podsecuritycontext-v1-core)_ |  |  |  |
| `additionalLabels` _object (keys:string, values:string)_ |  |  |  |
| `container` _[ContainerTemplate](#containertemplate)_ | ContainerSpec - customize the container template of the workload |  |  |
| `additionalVolumes` _[Volume](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#volume-v1-core) array_ |  |  |  |
