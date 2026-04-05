begin 'LongOptOnly: short fallback with compaction when long match fails'
# Question: when getopt_long_only(3) sees -abc and there is no long option
# "abc", does it fall back to short option compaction (-a -b -c)?
#
# The man page says: "If [a single-dash option] does not match a long option,
# but does match a short option, it is parsed as a short option instead."
#
# Test 1: -abc with short opts a,b,c and long opt "verbose"
# Expect: short compaction → -a -b -c
options='abc'
longopts='verbose'
printf '  optstring: %s\n' "${options}"
printf '  longopts: %s\n\n' "${longopts}"

set -- '-abc'
resp="$(getopt -a -l "${longopts}" -- "${options}" "${@}" 2>&1)" || :
printf '  input: %s\n' "${*}"
printf '  output: %s\n\n' "${resp## }"

# Test 2: -verbose with same config
# Expect: long match → --verbose
set -- '-verbose'
resp="$(getopt -a -l "${longopts}" -- "${options}" "${@}" 2>&1)" || :
printf '  input: %s\n' "${*}"
printf '  output: %s\n\n' "${resp## }"

# Test 3: -abc with short opts a,b,c and long opt "abc"
# Expect: long match wins → --abc (not compaction)
options='abc'
longopts='abc,verbose'
printf '  optstring: %s\n' "${options}"
printf '  longopts: %s\n\n' "${longopts}"

set -- '-abc'
resp="$(getopt -a -l "${longopts}" -- "${options}" "${@}" 2>&1)" || :
printf '  input: %s\n' "${*}"
printf '  output: %s\n\n' "${resp## }"

# Test 4: -ab with short opts a,b,c and long opt "able"
# Ambiguous: -ab could be long prefix for "able" or short compaction -a -b
# What does getopt_long_only choose?
options='abc'
longopts='able,verbose'
printf '  optstring: %s\n' "${options}"
printf '  longopts: %s\n\n' "${longopts}"

set -- '-ab'
resp="$(getopt -a -l "${longopts}" -- "${options}" "${@}" 2>&1)" || :
printf '  input: %s\n' "${*}"
printf '  output: %s\n' "${resp## }"
end
