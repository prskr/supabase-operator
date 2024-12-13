#!/usr/bin/env bash

set -euo pipefail

export DATABASE_URL="postgres://supabase_admin:1n1t-R00t!@localhost:5432/app"

go run mage.go Migrate
