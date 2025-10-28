#!/usr/bin/env bash
# See https://github.com/bazelbuild/rules_go/wiki/Editor-setup#3-editor-setup
exec bazel --output_base /tmp/bazel/output/supbase-operator run -- @rules_go//go/tools/gopackagesdriver "${@}"
