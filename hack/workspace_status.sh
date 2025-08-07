#!/usr/bin/env bash

current_tag=$(git rev-parse HEAD)

echo STABLE_GIT_COMMIT "${current_tag}"
echo STABLE_GIT_TAG "$(git describe --exact-match "${current_tag}" 2> /dev/null || printf 'dev')"
