#!/usr/bin/env bash
set -euo pipefail

# Generates release notes using Copilot CLI.
#
# Args: TAG PREVIOUS_TAG
# Requires: COPILOT_GITHUB_TOKEN (env)
# Output: release-notes.md

TAG="$1"
PREV="$2"

# Derive the module name for context.
case "$TAG" in
  goarg/v*)  MODULE="goarg" ;;
  pflag/v*) MODULE="pflag" ;;
  v*)        MODULE="optargs (core)" ;;
  *)         MODULE="optargs" ;;
esac

# Use printf with %s tokens to safely interpolate tag names without
# risk of shell expansion or special character interpretation.
PROMPT="$(printf 'You are writing release notes for the %s module of the OptArgs project, a Go implementation of POSIX/GNU getopt(3), getopt_long(3), and getopt_long_only(3).

Review the git log between tags %s and %s in this repository. Write concise, user-facing release notes in Markdown format:

1. Start with a 2-3 sentence overview summarizing the most important changes.
2. Then list changes grouped under these headings (omit empty sections):
   - **Features** — new capabilities
   - **Bug Fixes** — corrections to existing behavior
   - **Performance** — speed or memory improvements
   - **Documentation** — doc updates
   - **Other** — refactoring, CI, dependencies

Each item should be one line describing the user-facing impact, not the implementation detail. Do not include commit hashes or PR numbers. Write the output to release-notes.md.' \
  "$MODULE" "$PREV" "$TAG")"

copilot -p "$PROMPT" \
  --allow-tool='shell(git:*)' \
  --allow-tool=write \
  --no-ask-user

if [ ! -f release-notes.md ]; then
  echo "::error::Copilot did not generate release-notes.md"
  exit 1
fi

echo "::notice::Generated release notes for ${TAG}"
cat release-notes.md
