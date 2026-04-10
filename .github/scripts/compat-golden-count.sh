#!/usr/bin/env bash
set -euo pipefail

# Verifies golden files exist for compat test modules.

echo "=== goarg golden files ==="
find goarg/compat/testdata -name '*.golden.json' | wc -l

echo "=== pflag golden files ==="
find pflag/compat/testdata -name '*.golden.json' | wc -l
