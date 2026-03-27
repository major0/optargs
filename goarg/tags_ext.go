//go:build goarg_ext

package goarg

// Extended struct tag features beyond base alexflint/go-arg compatibility.
// Available only when built with -tags goarg_ext.

// ExtTagParser wraps TagParser with extended tag syntax support.
type ExtTagParser struct {
	TagParser
}

// ParseStructExt parses a struct with extended tag support.
// Currently identical to base ParseStruct — extended tag syntax
// (e.g. validation rules, custom type hints) will be added here.
func (etp *ExtTagParser) ParseStructExt(dest interface{}) (*StructMetadata, error) {
	return etp.ParseStruct(dest)
}
