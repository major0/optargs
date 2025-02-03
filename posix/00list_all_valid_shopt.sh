begin "Display all valid short options"
i=0
f=0
printf '  '
while test "${i}" -lt 128; do
	c="$(byte "${i}")"
	i="$((i+1))"

	# GNU/POSX explicilty disallows this characters
	! test "${c}" = '-' || continue
	! test "${c}" = ':' || continue
	! test "${c}" = ';' || continue

	# Everything else allowed by `isgraph(3)` that is not
	# `[[:space:]]` is allowed
	isgraph "${c}" || continue
	set -- $(getopt -- "a${c}" "-${c}") || :
	printf '%s ' "${1}"
        f=$((f + 1))
        ! test "$((f % 23))" -eq '0' || printf '\n  '
done
echo ''
unset i
unset c
unset f
end
