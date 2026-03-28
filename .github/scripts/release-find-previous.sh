#!/usr/bin/env bash
set -euo pipefail

# Finds the previous tag in the same series for changelog generation.
#
# RC tags: find the nearest ancestor tag in the series.
# Stable tags: skip past RC tags to find the previous stable release.
#
# Assumes all tags are annotated (enforced by release-ensure-annotated.sh).
#
# Args: TAG PREFIX
# Outputs: tag

TAG="$1"
PREFIX="$2"

case "$TAG" in
  *-rc*)
    PREV=$(git describe --tags --abbrev=0 --match "${PREFIX}*" \
             "${TAG}^" 2>/dev/null || true)
    ;;
  *)
    PREV=""
    COMMIT="${TAG}^"
    for _ in $(seq 1 50); do
      CANDIDATE=$(git describe --tags --abbrev=0 --match "${PREFIX}*" \
                    "$COMMIT" 2>/dev/null || true)
      [ -z "$CANDIDATE" ] && break
      case "$CANDIDATE" in
        *-rc*) COMMIT="${CANDIDATE}^" ;;
        *)     PREV="$CANDIDATE"; break ;;
      esac
    done
    ;;
esac

if [ -z "$PREV" ]; then
  echo "::notice::No previous tag found for series ${PREFIX}*; changelog will cover full history"
else
  echo "::notice::Previous tag: $PREV"
fi
echo "tag=$PREV" >> "$GITHUB_OUTPUT"
