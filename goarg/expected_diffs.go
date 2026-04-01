package goarg

// ExpectedDiff documents an intentional divergence from upstream alexflint/go-arg.
type ExpectedDiff struct {
	Scenario         string // e.g. "unknown_option.error"
	UpstreamBehavior string // what upstream produces
	OurBehavior      string // what we produce
	Rationale        string // why we diverge
}

// expectedDiffs lists all known intentional divergences.
// Each entry must be justified — POSIX/GNU compliance, upstream bugs, or
// extension-gated features.
var expectedDiffs = []ExpectedDiff{
	{
		Scenario:         "slice_option.values",
		UpstreamBehavior: "&{Files:[b.txt]}",
		OurBehavior:      "&{Files:[a.txt b.txt]}",
		Rationale: "Upstream treats repeated slice flags as greedy consumers " +
			"that reset the slice on each occurrence (--file a --file b → [b]). " +
			"Our POSIX-based core treats each --file as RequiredArgument taking " +
			"one value, appending to the slice (--file a --file b → [a b]). " +
			"Greedy multi-value consumption (--file a b c) is incompatible " +
			"with POSIX getopt semantics.",
	},
	{
		Scenario:         "required_missing.error",
		UpstreamBehavior: "INPUT is required",
		OurBehavior:      "required argument missing: input",
		Rationale: "Upstream uses 'FIELD is required' format with uppercase field name. " +
			"Our error translator uses 'required argument missing: field' with lowercase. " +
			"Both convey the same information; ours is consistent with other error formats.",
	},
	{
		Scenario:         "unknown_option.error",
		UpstreamBehavior: "unknown argument --unknown",
		OurBehavior:      "unrecognized argument: --unknown",
		Rationale: "Upstream uses 'unknown argument' without colon separator. " +
			"Our error translator uses 'unrecognized argument:' with colon. " +
			"Both convey the same information; ours follows GNU error conventions.",
	},
	{
		Scenario:         "map_type.values",
		UpstreamBehavior: "&{Headers:map[Accept:text/html]}",
		OurBehavior:      "&{Headers:map[Accept:text/html Content-Type:application/json]}",
		Rationale: "Upstream treats repeated map flags as greedy consumers " +
			"that reset the map on each occurrence (--header a=b --header c=d → {c:d}). " +
			"Our POSIX-based core treats each --header as RequiredArgument taking " +
			"one value, merging into the map (--header a=b --header c=d → {a:b, c:d}). " +
			"Same root cause as the slice_option divergence.",
	},
}

// HelpUsageDiffRationale explains the systematic help/usage formatting
// differences. These affect all scenarios uniformly and are not listed
// individually in expectedDiffs.
const HelpUsageDiffRationale = "Help and usage formatting differs systematically from upstream: " +
	"(1) usage line uses [OPTIONS] instead of listing each option, " +
	"(2) short options listed before long (POSIX convention) vs upstream long-first, " +
	"(3) column widths and alignment differ, " +
	"(4) default values use (default: X) vs upstream [default: X], " +
	"(5) subcommand help shows root-level view vs upstream shows active subcommand. " +
	"These are deliberate formatting choices; parsed values and error semantics " +
	"are the compatibility surface."

// loadExpectedDiffs returns a map keyed by scenario for O(1) lookup.
func loadExpectedDiffs() map[string]ExpectedDiff {
	m := make(map[string]ExpectedDiff, len(expectedDiffs))
	for _, d := range expectedDiffs {
		m[d.Scenario] = d
	}
	return m
}
