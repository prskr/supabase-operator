#!/bin/bash

supabase db push \
    --include-seed \
    --db-url "postgresql://supabase_admin:1n1t-R00t!@localhost:5432/app"
