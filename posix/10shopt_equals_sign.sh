begin "The equals sign is a valid short option"
options='abc=:'
printf '  optstring: %s\n\n' "${options}"

set -- 'param' '-abc=1'
resp="$(getopt -- "${options}" "${@}")"
printf '  input: %s\n' "${*}"
printf '  output: %s\n' "${resp## }"
end
