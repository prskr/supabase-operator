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
| `jwks` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#secretkeyselector-v1-core)_ | JWKSSelector - selector where the JWKS can be retrieved from to enable the API gateway to validate JWTs |  |  |
| `serviceSelector` _[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#labelselector-v1-meta)_ | ServiceSelector - selector to match all Supabase services (or in fact EndpointSlices) that should be considered for this APIGateway | \{ matchExpressions:[map[key:app.kubernetes.io/part-of operator:In values:[supabase]] map[key:supabase.k8s.icb4dc0.de/api-gateway-target operator:Exists]] \} |  |
| `componentTypeLabel` _string_ | ComponentTypeLabel - Label to identify which Supabase component a Service represents (e.g. auth, postgrest, ...) | app.kubernetes.io/name |  |




#### AuthProviderMeta







_Appears in:_
- [AzureAuthProvider](#azureauthprovider)
- [EmailAuthProvider](#emailauthprovider)
- [GithubAuthProvider](#githubauthprovider)
- [PhoneAuthProvider](#phoneauthprovider)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `enabled` _boolean_ | Enabled - whether the authentication provider is enabled or not |  |  |


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
| `workloadTemplate` _[WorkloadTemplate](#workloadtemplate)_ |  |  |  |
| `emailSignupDisabled` _boolean_ |  |  |  |


#### AzureAuthProvider







_Appears in:_
- [AuthProviders](#authproviders)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `enabled` _boolean_ | Enabled - whether the authentication provider is enabled or not |  |  |
| `clientID` _string_ |  |  |  |
| `clientSecretRef` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#secretkeyselector-v1-core)_ |  |  |  |
| `url` _string_ |  |  |  |


#### ContainerTemplate







_Appears in:_
- [WorkloadTemplate](#workloadtemplate)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `image` _string_ |  |  |  |
| `pullPolicy` _[PullPolicy](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#pullpolicy-v1-core)_ |  |  |  |
| `imagePullSecrets` _[LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#localobjectreference-v1-core) array_ |  |  |  |
| `securityContext` _[SecurityContext](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#securitycontext-v1-core)_ | SecurityContext - |  |  |
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
| `secret` _string_ | Secret - JWT HMAC secret in plain text<br />This is WRITE-ONLY and will be copied to the SecretRef by the defaulter |  |  |
| `secretRef` _[LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#localobjectreference-v1-core)_ | SecretRef - object reference to the Secret where JWT values are stored |  |  |
| `secretKey` _string_ | SecretKey - key in secret where to read the JWT HMAC secret from | secret |  |
| `jwksKey` _string_ | JwksKey - key in secret where to read the JWKS from | jwks.json |  |
| `anonKey` _string_ | AnonKey - key in secret where to read the anon JWT from | anon_key |  |
| `serviceKey` _string_ | ServiceKey - key in secret where to read the service JWT from | service_key |  |
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


#### DashboardDbSpec







_Appears in:_
- [DashboardSpec](#dashboardspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `host` _string_ |  |  |  |
| `port` _integer_ | Port - Database port, typically 5432 | 5432 |  |
| `dbName` _string_ |  |  |  |
| `dbCredentialsRef` _[LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#localobjectreference-v1-core)_ | DBCredentialsRef - reference to a Secret key where the DB credentials can be retrieved from<br />Credentials need to be stored in basic auth form |  |  |


#### DashboardJwtSpec







_Appears in:_
- [StudioSpec](#studiospec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `secretRef` _[LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#localobjectreference-v1-core)_ | SecretRef - object reference to the Secret where JWT values are stored |  |  |
| `secretKey` _string_ | SecretKey - key in secret where to read the JWT HMAC secret from | secret |  |
| `anonKey` _string_ | AnonKey - key in secret where to read the anon JWT from | anon_key |  |
| `serviceKey` _string_ | ServiceKey - key in secret where to read the service JWT from | service_key |  |


#### DashboardList



DashboardList contains a list of Dashboard.





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `supabase.k8s.icb4dc0.de/v1alpha1` | | |
| `kind` _string_ | `DashboardList` | | |
| `metadata` _[ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#listmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `items` _[Dashboard](#dashboard) array_ |  |  |  |


#### DashboardSpec



DashboardSpec defines the desired state of Dashboard.



_Appears in:_
- [Dashboard](#dashboard)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `db` _[DashboardDbSpec](#dashboarddbspec)_ |  |  |  |
| `pgMeta` _[PGMetaSpec](#pgmetaspec)_ | PGMeta |  |  |
| `studio` _[StudioSpec](#studiospec)_ | Studio |  |  |




#### Database







_Appears in:_
- [CoreSpec](#corespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `dsn` _string_ |  |  |  |
| `dsnFrom` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#secretkeyselector-v1-core)_ |  |  |  |
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
| `supabaseAdmin` _[LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#localobjectreference-v1-core)_ |  |  |  |
| `authenticator` _[LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#localobjectreference-v1-core)_ |  |  |  |
| `supabaseAuthAdmin` _[LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#localobjectreference-v1-core)_ |  |  |  |
| `supabaseFunctionsAdmin` _[LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#localobjectreference-v1-core)_ |  |  |  |
| `supabaseStorageAdmin` _[LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#localobjectreference-v1-core)_ |  |  |  |


#### DatabaseStatus







_Appears in:_
- [CoreStatus](#corestatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `appliedMigrations` _[MigrationStatus](#migrationstatus)_ |  |  |  |
| `roles` _object (keys:string, values:integer array)_ |  |  |  |


#### EmailAuthProvider







_Appears in:_
- [AuthProviders](#authproviders)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `enabled` _boolean_ | Enabled - whether the authentication provider is enabled or not |  |  |
| `adminEmail` _string_ |  |  |  |
| `senderName` _string_ |  |  |  |
| `autoconfirmEmail` _boolean_ |  |  |  |
| `subjectsInvite` _string_ |  |  |  |
| `subjectsConfirmation` _string_ |  |  |  |
| `smtpSpec` _[EmailAuthSmtpSpec](#emailauthsmtpspec)_ |  |  |  |


#### EmailAuthSmtpSpec







_Appears in:_
- [EmailAuthProvider](#emailauthprovider)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `host` _string_ |  |  |  |
| `port` _integer_ |  |  |  |
| `maxFrequency` _integer_ |  |  |  |
| `credentialsFrom` _[LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#localobjectreference-v1-core)_ |  |  |  |


#### EnvoySpec







_Appears in:_
- [APIGatewaySpec](#apigatewayspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `nodeName` _string_ | NodeName - identifies the Envoy cluster within the current namespace<br />if not set, the name of the APIGateway resource will be used<br />The primary use case is to make the assignment of multiple supabase instances in a single namespace explicit. |  |  |
| `controlPlane` _[ControlPlaneSpec](#controlplanespec)_ | ControlPlane - configure the control plane where Envoy will retrieve its configuration from |  |  |
| `workloadTemplate` _[WorkloadTemplate](#workloadtemplate)_ | WorkloadTemplate - customize the Envoy deployment |  |  |


#### EnvoyStatus







_Appears in:_
- [APIGatewayStatus](#apigatewaystatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `configVersion` _string_ |  |  |  |
| `resourceHash` _integer array_ |  |  |  |


#### GithubAuthProvider







_Appears in:_
- [AuthProviders](#authproviders)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `enabled` _boolean_ | Enabled - whether the authentication provider is enabled or not |  |  |
| `clientID` _string_ |  |  |  |
| `clientSecretRef` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#secretkeyselector-v1-core)_ |  |  |  |
| `url` _string_ |  |  |  |


#### ImageSpec







_Appears in:_
- [ContainerTemplate](#containertemplate)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `image` _string_ |  |  |  |
| `pullPolicy` _[PullPolicy](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#pullpolicy-v1-core)_ |  |  |  |


#### MigrationStatus

_Underlying type:_ _[Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#time-v1-meta)_





_Appears in:_
- [DatabaseStatus](#databasestatus)



#### OAuthProvider







_Appears in:_
- [AzureAuthProvider](#azureauthprovider)
- [GithubAuthProvider](#githubauthprovider)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `clientID` _string_ |  |  |  |
| `clientSecretRef` _[SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#secretkeyselector-v1-core)_ |  |  |  |
| `url` _string_ |  |  |  |


#### PGMetaSpec







_Appears in:_
- [DashboardSpec](#dashboardspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `workloadTemplate` _[WorkloadTemplate](#workloadtemplate)_ | WorkloadTemplate - customize the pg-meta deployment |  |  |


#### PhoneAuthProvider







_Appears in:_
- [AuthProviders](#authproviders)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `enabled` _boolean_ | Enabled - whether the authentication provider is enabled or not |  |  |


#### PostgrestSpec







_Appears in:_
- [CoreSpec](#corespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `schemas` _string array_ | Schemas - schema where PostgREST is looking for objects (tables, views, functions, ...) | [public graphql_public] |  |
| `extraSearchPath` _string array_ | ExtraSearchPath - Extra schemas to add to the search_path of every request.<br />These schemas tables, views and functions don’t get API endpoints, they can only be referred from the database objects inside your db-schemas. | [public extensions] |  |
| `anonRole` _string_ | AnonRole - name of the anon role | anon |  |
| `maxRows` _integer_ | MaxRows - maximum number of rows PostgREST will load at a time | 1000 |  |
| `workloadTemplate` _[WorkloadTemplate](#workloadtemplate)_ | WorkloadTemplate - customize the PostgREST workload |  |  |


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
| `foo` _string_ | Foo is an example field of Storage. Edit storage_types.go to remove/update |  |  |




#### StudioSpec







_Appears in:_
- [DashboardSpec](#dashboardspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `jwt` _[DashboardJwtSpec](#dashboardjwtspec)_ |  |  |  |
| `workloadTemplate` _[WorkloadTemplate](#workloadtemplate)_ | WorkloadTemplate - customize the studio deployment |  |  |
| `gatewayServiceSelector` _object (keys:string, values:string)_ | GatewayServiceSelector - selector to find the service for the API gateway<br />Required to configure the API URL in the studio deployment<br />If you don't run multiple APIGateway instances in the same namespaces, the default will be fine | \{ app.kubernetes.io/component:api-gateway app.kubernetes.io/name:envoy \} |  |
| `externalUrl` _string_ | APIExternalURL is referring to the URL where Supabase API will be available<br />Typically this is the ingress of the API gateway |  |  |


#### WorkloadTemplate







_Appears in:_
- [AuthSpec](#authspec)
- [EnvoySpec](#envoyspec)
- [PGMetaSpec](#pgmetaspec)
- [PostgrestSpec](#postgrestspec)
- [StudioSpec](#studiospec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `replicas` _integer_ |  |  |  |
| `securityContext` _[PodSecurityContext](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#podsecuritycontext-v1-core)_ |  |  |  |
| `additionalLabels` _object (keys:string, values:string)_ |  |  |  |
| `workload` _[ContainerTemplate](#containertemplate)_ | Workload - customize the container template of the workload |  |  |


