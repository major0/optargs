package pflag

import (
	"flag"
)

// goFlagValue wraps a Go stdlib flag.Value to satisfy pflag.Value.
type goFlagValue struct {
	inner    flag.Value
	typeName string
}

func (v *goFlagValue) String() string     { return v.inner.String() }
func (v *goFlagValue) Set(s string) error { return v.inner.Set(s) }
func (v *goFlagValue) Type() string       { return v.typeName }

// PFlagFromGoFlag converts a Go stdlib flag.Flag to a pflag Flag.
func PFlagFromGoFlag(goflag *flag.Flag) *Flag { //nolint:revive // name matches spf13/pflag API
	return &Flag{
		Name:     goflag.Name,
		Usage:    goflag.Usage,
		Value:    &goFlagValue{inner: goflag.Value, typeName: typeNameString},
		DefValue: goflag.DefValue,
	}
}

// AddGoFlag adds a single Go stdlib flag to the FlagSet.
func (f *FlagSet) AddGoFlag(goflag *flag.Flag) {
	f.AddFlag(PFlagFromGoFlag(goflag))
}

// AddGoFlagSet adds all flags from a Go stdlib FlagSet.
func (f *FlagSet) AddGoFlagSet(goflags *flag.FlagSet) {
	if goflags == nil {
		return
	}
	goflags.VisitAll(func(goflag *flag.Flag) {
		f.AddGoFlag(goflag)
	})
}

// CopyToGoFlagSet copies all pflag flags to a Go stdlib FlagSet.
// Each flag is registered as a string flag using its current string value.
func CopyToGoFlagSet(pfs *FlagSet, gofs *flag.FlagSet) {
	pfs.VisitAll(func(pf *Flag) {
		gofs.String(pf.Name, pf.DefValue, pf.Usage)
	})
}
