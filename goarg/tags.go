package goarg

import (
	"reflect"

	"github.com/major0/optargs"
)

// StructMetadata represents parsed struct information
type StructMetadata struct {
	Fields      []FieldMetadata
	Subcommands map[string]*StructMetadata
	Program     string
	Description string
	Version     string
}

// FieldMetadata represents a single struct field's CLI mapping
type FieldMetadata struct {
	Name        string
	Type        reflect.Type
	Tag         string
	Short       string
	Long        string
	Help        string
	Required    bool
	Positional  bool
	Env         string
	Default     interface{}

	// Direct OptArgs Core mapping
	CoreFlag *optargs.Flag
	ArgType  optargs.ArgType
}

// TagParser processes struct tags - identical behavior to alexflint/go-arg
type TagParser struct{}

// ParseStruct parses a struct and returns its metadata
func (tp *TagParser) ParseStruct(dest interface{}) (*StructMetadata, error) {
	// TODO: Implement struct parsing
	return &StructMetadata{
		Fields:      []FieldMetadata{},
		Subcommands: make(map[string]*StructMetadata),
	}, nil
}

// ParseField parses a single struct field and returns its metadata
func (tp *TagParser) ParseField(field reflect.StructField) (*FieldMetadata, error) {
	// TODO: Implement field parsing
	return &FieldMetadata{
		Name: field.Name,
		Type: field.Type,
		Tag:  string(field.Tag),
	}, nil
}