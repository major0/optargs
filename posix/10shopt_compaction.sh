begin "short option compaction"
options='abc::'
printf '  optstring: %s\n\n' "${options}"

for args in 'abcarg' 'acbarg'; do
	resp="$(getopt -- "${options}" 'param' "-${args}")"
	printf '  %s = %s\n' "param -${args}" "${resp}"
done
end
