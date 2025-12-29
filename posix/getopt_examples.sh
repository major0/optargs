#!/bin/sh
# This is a simple script attempts to demonstrate valid GNU/POSIX uses of
# `getopt(3)`. This script relies on the util-linux version of the
# `getopt(1)` CLI utility which makes significant efforts to expose
# `getopt(3)`, `getopt_long(3)`, and `getopt_long_onlyy(3)` to the CLI.
#
# Some of the use of the `getopt(1)` utility may seem non-intuitive to
# anyone reviewing this script. From the `getopt(1)` man page:
#
#       getopt(3) can parse long options with optional arguments that are
#       given an empty optional argument (but cannot do this for short
#       options). This getopt(1) treats optional arguments that are empty
#       as if they were not present.
#
#       The syntax if you do not want any short option variables at all
#       is not very intuitive (you have to set them explicitly to the empty
#       string).
set -e


##
# Helper functions
error(){ echo "error: $*" >&2; }
die() { error "$*"; exit 1; }
begin() { printf '\n### %s\n' "${*}"; }
end() { printf '##\n';}

##
# Output the ASCII character of the given decimal value
byte() { printf "\\$(printf '%03o' "${1}")"; }
getopt_version()
{
	! test -f "$(which getopt)" && echo 127 && return 127 || :
	set +e
	getopt -T > /dev/null 2>&1
	set -- "$?"
	set -e
	echo "${1}" && return "${1}"
}
isgraph() {
	# expr's regexp is not reliable for these tests
	! printf '%s' "${1}" | grep -q '^[[:space:]]$' || return 1
	printf '%s' "${1}" | grep -q '^[[:graph:]]$' || return 1
	return 0
}

##
# Validate we have a usable version of `getopt(1)`
if test "$(getopt_version)" -ne '4'; then
	test "$?" -eq '4' || die 'no usable version of getopt(1) detected'
fi

for f in $(cd "${0%%/*}" && echo *.sh); do
	test "${f}" != "${0##*/}" || continue # don't rerun the entrypoint
	. "${0%/*}/${f}"
done
