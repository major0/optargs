// Extension system for goarg.
//
// Extensions add capabilities beyond base alexflint/go-arg compatibility.
// They are included at build time via the "goarg_ext" build tag:
//
//	go build -tags goarg_ext
//	go test -tags goarg_ext ./...
//
// Without the tag, goarg provides 100% alexflint/go-arg API compatibility.
// With the tag, additional features become available (enhanced tag syntax,
// additional parser options, etc.).
//
// Extension files follow the naming convention *_ext.go and use the
// //go:build goarg_ext directive.
package goarg

// extEnabled is set to true by ext_enabled.go when built with -tags goarg_ext.
var extEnabled bool

// ExtensionsEnabled reports whether goarg extensions are compiled in.
func ExtensionsEnabled() bool {
	return extEnabled
}
