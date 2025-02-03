begin 'Required args to short options are greedy'
options='ab:c::'
printf '  optstring: %s\n\n' "${options}"

set -- 'param' '-b' '-1' '-a'
resp="$(getopt -- "${options}" "${@}")"
printf '  input: %s\n' "${*}"
printf '  output: %s\n' "${resp## }"
printf '\n  note: optional arguments to short options are _also_ greedy in `getopt(3)` but not `getopt(1)`\n'
end
