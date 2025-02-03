begin 'Short options do not allow equals to specify arguments'
options='ab:c::'
printf '  optstring: %s\n\n' "${options}"

set -- 'param' '-a' '-b=1'
resp="$(getopt -- "${options}" "${@}")"
printf '  input: %s\n' "${*}"
printf '  output: %s\n' "${resp## }"
end

begin 'But longopts in `only` longopts mode supports short options with arguments'
options=''
longopts='a,b:,c::'
printf '  optstring: %s\n' "${options}"
printf '  longopts: %s\n\n' "${longopts}"

set -- 'param' '-a' '-b' '-1' '-c=1'
resp="$(getopt -a -l "${longopts}" -- '' "${@}")"
printf '  input: %s\n' "${*}"
printf '  output: %s\n' "${resp## }"
end
