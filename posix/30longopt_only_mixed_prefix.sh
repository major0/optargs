begin 'LongOptOnly allows single or double hyphen for the same argument'
options=''
longopts='a,b:,c::'
printf '  optstring: %s\n' "${options}"
printf '  longopts: %s\n\n' "${longopts}"

set -- '-c=foo' '--c=bar'
resp="$(getopt -a -l "${longopts}" -- "${options}" "${@}")"
printf '  input: %s\n' "${*}"
printf '  output: %s\n' "${resp## }"
end
