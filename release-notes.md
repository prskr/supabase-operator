## [0.6.0] - 2026-03-31

### 🚀 Features

- Dev setup with Bazel
- Support collecting metrics from envoy
- *(postgrest)* Scrape metrics from PostgREST
- Update to latest Supabase release
- *(auth)* Prepare new auth routes
- *(ci)* Cache Bazel build caches
- *(generate)* Do not require to pass in  for code generation targets

### 🐛 Bug Fixes

- *(postgres)* Update Postgres extensions
- *(postgres)* Remove existing manifest before creating new one
- *(ci)* Don't delete architecture specific interim images
- *(postgres)* Do not build Postgres 18 for now
- *(ci)* Push Postgres images to GHCR
- *(ci)* Ignore test failures for now
- *(ci)* Use  for cache key
- *(ci)* Use  as cache key to re-use cache also for tag builds
- *(ci)* Declare necessary permissions for creating releases and packages
- *(ci)* Use correct github token for container registry login
- *(ci)* Use podman-login for container registry login
- *(ci)* Use git-cliff action to generate changelog
- *(ci)* Update image reference in kustomize edit

### 🚜 Refactor

- *(ci)* Use Podman & Buildah to build Postgres image
- Replace shell hacks with Go code generation

### 📚 Documentation

- Update URLs for GitHub
- Create developer docs and extend existing architecture doc
- Document database connection and auth setup
- Document storage API deployment

### ⚙️ Miscellaneous Tasks

- Updates of Bazel modules and Supabase resources
- Upgrade Go to 1.25
- Upgrade dependencies
- *(postgres)* Build for Postgres major 1{5,6,7,8}
- Migrate Go module name to GitHub and clean Bazel stuff
- *(ci)* Use github pages to host docs
- Remove Makefile and create Bazel shell targets to generate code
- *(supabase)* Upgrade images and migrations
- Cleanup of local development setup
## [0.5.1] - 2025-08-07

### 🚀 Features

- Build with bazel
- *(assets)* Fetch DB migrations with Bazel
- First working build
- *(ci)* Prepare release

### 🐛 Bug Fixes

- *(ci)* Re-configure setup-bazel
- *(ci)* Use pre-installed bazelisk instead of pipeline action
- *(ci)* Correct usage of --bazelrc flag
- Renew cert for control plane before it expires
- *(ci)* Use empty GH token for git-cliff action

### ⚙️ Miscellaneous Tasks

- Update deps
- Update Go toolchain
## [0.5.0] - 2025-04-01

### 🚀 Features

- *(ci)* Update Postgres minor version and delete temporary tags

### 🚜 Refactor

- Don't mount service account token into workloads

### ⚙️ Miscellaneous Tasks

- Update to Go 1.24
- Switch to Harbor registry
- Update migrations & images and switch to new registry
## [0.4.2] - 2025-02-13

### 🚀 Features

- *(apigateay)* Add OIDC and basic auth support
- *(webhook)* Validate dashboard auth spec
- Custom postgres images
- *(ci)* Configure image build caching

### 🐛 Bug Fixes

- *(db)* Track state of migrations and execute them again when necessary
- *(ci)* Remove cache for postgres images for now

### ⚙️ Miscellaneous Tasks

- *(postgres)* Schedule new image every week
## [0.4.1] - 2025-02-04

### 🚀 Features

- *(apigateway)* Allow to enable debug logging

### 📚 Documentation

- Update CRD docs
## [0.4.0] - 2025-02-03

### 🚀 Features

- *(dashboard)* PoC Oauth2 auth

### 🐛 Bug Fixes

- *(envoy)* Version not handled properly
- Route propagation

### 📚 Documentation

- Extend docs

### ⚙️ Miscellaneous Tasks

- Update deps
## [0.3.0] - 2025-01-24

### 🚜 Refactor

- *(apigateway)* Configure api & dashboard listeneres individually
## [0.2.0] - 2025-01-23

### 🚀 Features

- Prepare test execution with mage & update tools
- *(storage)* Finish initial basic implementation
## [0.1.0] - 2025-01-22

### 🚀 Features

- *(db)* Prepare migrations and core CRD
- Basic functionality implemented
- *(dashboard)* Initial support for studio & pg-meta services
- Prepare release
- *(storage)* Prepare custom resource for storage API

### 🐛 Bug Fixes

- *(ci)* Use arm64 version of kind if necessary
- *(ci)* Use rclone instead of AWS CLI
- Minor issues
- *(ci)* Unset GITHUB_TOKEN for release
- *(ci)* Force Gitea token
- *(ci)* Omit import path in container images
- *(ci)* Login to container registry
- *(ci)* Configure gitea URLs

### 🚜 Refactor

- *(db)* Extract Supabase migrations from release artifact
- *(gateway)* Make node name explicitly configurable in spec
- Implement control plane as controller-runtime manager

### 📚 Documentation

- Initial CI setup

### ⚙️ Miscellaneous Tasks

- Add husky config
- Update image versions
- Add license header to all files
- Setup some example schema to play around
- Cleanup examples
