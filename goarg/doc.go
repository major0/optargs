// Package goarg provides 100% API compatibility with alexflint/go-arg
// while leveraging OptArgs Core's superior POSIX/GNU compliance.
//
// This package implements a complete compatibility layer that allows
// existing alexflint/go-arg applications to work without modification
// while benefiting from enhanced argument parsing capabilities.
//
// The architecture is intentionally simple: goarg interfaces directly
// with OptArgs Core without intermediate layers. Extensions are handled
// architecturally through separate -ext.go files.
package goarg