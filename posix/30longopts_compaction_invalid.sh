begin 'Argument compaction is never supported for LongOptsOnly'
options=''
longopts='a,b,c'
printf '  optstring: %s\n' "${options}"
printf '  longopts: %s\n\n' "${longopts}"

set -- '-abc'
resp="$(getopt -a -l "${longopts}" -- "${options}" "${@}" 2>&1)" || :
printf '  input: %s\n' "${*}"
printf '  output: %s\n' "${resp## }"
end
