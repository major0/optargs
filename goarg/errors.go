package goarg

import "errors"

// ErrHelp indicates that the builtin --help flag was provided.
var ErrHelp = errors.New("help requested by user")

// ErrVersion indicates that the builtin --version flag was provided.
var ErrVersion = errors.New("version requested by user")

// Versioned is implemented by destination structs that provide a version string.
// When implemented, --version is registered and the version appears in help output.
type Versioned interface {
	Version() string
}

// Described is implemented by destination structs that provide a description.
// When implemented, the description appears at the top of help output.
type Described interface {
	Description() string
}

// Epilogued is implemented by destination structs that provide epilogue text.
// When implemented, the epilogue appears at the bottom of help output.
type Epilogued interface {
	Epilogue() string
}
