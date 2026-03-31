#!/usr/bin/env bash

current_commit=$(git rev-parse HEAD)
current_tag="$(git describe --exact-match "${current_commit}" 2>/dev/null || printf 'dev')"

echo STABLE_GIT_COMMIT "${current_commit}"
echo STABLE_GIT_TAG "${current_tag#v}"
