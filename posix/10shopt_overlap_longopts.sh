begin 'Long and Short options can look similar, but can behave differently'
options='ab:c::'
longopts='a,b:,c::'
printf '  optstring: %s\n' "${options}"
printf '  longopts: %s\n\n' "${longopts}"

set -- '-a' '1' '--a' '1'
resp="$(getopt -l "${longopts}" -- "${options}" "$@")"
printf '  input: %s\n' "${*}"
printf '  output: %s\n' "${resp## }"
echo ''
set -- '-b' '-1' '-b=-1' '--b=-1'
resp="$(getopt -l "${longopts}" -- "${options}" "$@")"
printf '  input: %s\n' "${*}"
printf '  output: %s\n' "${resp## }"
echo ''
set -- '-c' '1' '-c=1' '--c=1'
resp="$(getopt -l "${longopts}" -- "${options}" "$@")"
printf '  input: %s\n' "${*}"
printf '  output: %s\n' "${resp## }"
end
unset options
unset resp
