begin 'Can emulate passing an empty string to a long opt using the equals sign'
options='ab:c::'
printf '  optstring: %s\n\n' "${options}"

set -- 'param' '--b='
resp="$(getopt -l a,b:,c:: -- "${options}" "${@}")"
printf '  input: %s\n' "${*}"
printf '  output: %s\n' "${resp## }"
end
