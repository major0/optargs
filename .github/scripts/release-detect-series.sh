#!/usr/bin/env bash
set -euo pipefail

# Determines the tag series (prefix and human-readable name) from GITHUB_REF.
# Outputs: tag, prefix, name

TAG="${GITHUB_REF#refs/tags/}"
echo "tag=$TAG" >> "$GITHUB_OUTPUT"

# Extract the series prefix
case "$TAG" in
  goarg/v*)  PREFIX="goarg/v" ;;
  pflags/v*) PREFIX="pflags/v" ;;
  v*)        PREFIX="v" ;;
  *)         echo "::error::Unrecognized tag: $TAG"; exit 1 ;;
esac
echo "prefix=$PREFIX" >> "$GITHUB_OUTPUT"

# Derive a human-readable name for the release title
case "$PREFIX" in
  goarg/v)  NAME="goarg" ;;
  pflags/v) NAME="pflags" ;;
  v)        NAME="optargs" ;;
esac
echo "name=$NAME" >> "$GITHUB_OUTPUT"
