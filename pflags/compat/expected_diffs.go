// Package compat documents expected behavioral differences between
// github.com/major0/optargs/pflags and upstream github.com/spf13/pflag.
//
// Each entry describes: scenario, upstream behavior, our behavior, rationale.
package compat

// ExpectedDiff documents a single intentional behavioral divergence.
type ExpectedDiff struct {
	Scenario  string // what the user does
	Upstream  string // what spf13/pflag does
	Ours      string // what optargs/pflags does
	Rationale string // why we diverge
}

// ExpectedDiffs enumerates all known intentional divergences.
var ExpectedDiffs = []ExpectedDiff{
	{
		Scenario:  "POSIX short-option compaction (-abc)",
		Upstream:  "Not supported; -abc is parsed as -a with argument 'bc' for non-boolean flags, or three separate flags only if all are boolean NoArgument",
		Ours:      "Fully supported via OptArgs Core; -abc expands to -a -b -c when all are NoArgument, last flag in group may take an argument",
		Rationale: "POSIX/GNU compliance: getopt(3) specifies compaction behavior",
	},
	{
		Scenario:  "Short-only flags (no long name)",
		Upstream:  "Not supported; every flag must have a long name, shorthand is optional",
		Ours:      "Supported via ShortVar() API; a flag can exist only as a short option with no long equivalent",
		Rationale: "Many POSIX utilities have short-only options (e.g., tar -x, ls -l with no --long equivalent in some implementations)",
	},
	{
		Scenario:  "Many-to-one flag mappings (multiple flags → one destination)",
		Upstream:  "Not supported; each flag has its own destination variable",
		Ours:      "Supported via shared Flag.Handle callbacks writing to the same Value; enables ls --format=across / -x pattern",
		Rationale: "Common POSIX pattern where short options are aliases for long option values (GNU coreutils ls, sort, etc.)",
	},
	{
		Scenario:  "--no-<flag>=true and --no-<flag>=false (explicit values on negation)",
		Upstream:  "Not supported; --no-verbose is not recognized at all",
		Ours:      "--no-verbose sets false; --no-verbose=true sets false; --no-verbose=false sets true (double negation)",
		Rationale: "GNU convention for boolean negation; explicit values allow scripted flag composition without conditional logic",
	},
	{
		Scenario:  "-- prefix for short options (--x resolving single-char flag)",
		Upstream:  "Allowed; --x resolves to the flag named 'x' if it exists as a long option",
		Ours:      "Rejected unless long-only parsing is enabled; --x is a long option lookup, not a short option",
		Rationale: "Prevents ambiguity between short and long option namespaces; POSIX specifies - for short, -- for long",
	},
	{
		Scenario:  "Error message format for unknown flags",
		Upstream:  `"unknown flag: --name" (includes -- prefix in message)`,
		Ours:      `"unknown flag: --name" for long, "unknown shorthand flag: 'x'" for short`,
		Rationale: "Behavioral parity for long flags; short flag errors include quotes around the character for clarity",
	},
}
