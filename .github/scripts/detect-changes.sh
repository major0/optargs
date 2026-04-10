#!/usr/bin/env bash
set -euo pipefail

# Detects which modules have changed files.
# On push events, all modules are marked as changed.
# On PR events, only modules with changed files are marked.
#
# Args: EVENT_NAME
# Outputs (to GITHUB_OUTPUT): optargs, goarg, pflag

EVENT_NAME="$1"

if [ "$EVENT_NAME" = "push" ]; then
	echo "optargs=true"
	echo "goarg=true"
	echo "pflag=true"
else
	CHANGED=$(git diff --name-only HEAD~1 2>/dev/null || echo "")

	if echo "$CHANGED" | grep -qE '^[^/]+\.(go|mod|sum)$'; then
		echo "optargs=true"
	else
		echo "optargs=false"
	fi

	if echo "$CHANGED" | grep -qE '^(goarg/|[^/]+\.(go|mod|sum)$)'; then
		echo "goarg=true"
	else
		echo "goarg=false"
	fi

	if echo "$CHANGED" | grep -qE '^(pflag/|[^/]+\.(go|mod|sum)$)'; then
		echo "pflag=true"
	else
		echo "pflag=false"
	fi
fi
