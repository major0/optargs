begin 'Any `isgraph()` character is a valid long option character'
options='ab:c::'
longopts='foo-bar,subcmd:param,{weird^options}'
printf '  optstring: %s\n' "${options}"
printf '  longopts: %s\n\n' "${longopts}"

set -- 'param' '--foo-bar' '--subcmd:param' '--{weird^options}'
resp="$(getopt -l "${longopts}" -- "${options}" "${@}")"
printf '  input: %s\n' "${*}"
printf '  output: %s\n' "${resp## }"
printf '\n  note: the `=` is technically also allowed in the long option name when using `getopt_long(3)` but not when using `getopt(1)`.\n'
end
