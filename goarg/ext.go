package goarg

// extEnabled is set to true by ext_enabled.go when built with -tags goarg_ext.
var extEnabled bool

// ExtensionsEnabled reports whether goarg extensions are compiled in.
func ExtensionsEnabled() bool {
	return extEnabled
}
