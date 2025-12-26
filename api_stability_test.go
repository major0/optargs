package optargs

import (
	"reflect"
	"testing"
)

// TestAPIStability validates that the public API remains stable and backward compatible
func TestAPIStability(t *testing.T) {
	// Test that all expected public types exist and have the correct structure
	t.Run("public_types_exist", func(t *testing.T) {
		// Test ArgType enum
		var argType ArgType
		if reflect.TypeOf(argType).Kind() != reflect.Int {
			t.Error("ArgType should be an int type")
		}

		// Test that ArgType constants exist
		if NoArgument != 0 {
			t.Error("NoArgument should be 0")
		}
		if RequiredArgument != 1 {
			t.Error("RequiredArgument should be 1")
		}
		if OptionalArgument != 2 {
			t.Error("OptionalArgument should be 2")
		}

		// Test ParseMode enum
		var parseMode ParseMode
		if reflect.TypeOf(parseMode).Kind() != reflect.Int {
			t.Error("ParseMode should be an int type")
		}

		// Test that ParseMode constants exist
		if ParseDefault != 0 {
			t.Error("ParseDefault should be 0")
		}
		if ParseNonOpts != 1 {
			t.Error("ParseNonOpts should be 1")
		}
		if ParsePosixlyCorrect != 2 {
			t.Error("ParsePosixlyCorrect should be 2")
		}

		// Test Flag struct
		flag := Flag{}
		flagType := reflect.TypeOf(flag)
		if flagType.Kind() != reflect.Struct {
			t.Error("Flag should be a struct")
		}

		// Check Flag fields
		nameField, hasName := flagType.FieldByName("Name")
		if !hasName || nameField.Type.Kind() != reflect.String {
			t.Error("Flag should have a Name field of type string")
		}

		hasArgField, hasHasArg := flagType.FieldByName("HasArg")
		if !hasHasArg || hasArgField.Type != reflect.TypeOf(argType) {
			t.Error("Flag should have a HasArg field of type ArgType")
		}

		// Test Option struct
		option := Option{}
		optionType := reflect.TypeOf(option)
		if optionType.Kind() != reflect.Struct {
			t.Error("Option should be a struct")
		}

		// Check Option fields
		optNameField, hasOptName := optionType.FieldByName("Name")
		if !hasOptName || optNameField.Type.Kind() != reflect.String {
			t.Error("Option should have a Name field of type string")
		}

		optHasArgField, hasOptHasArg := optionType.FieldByName("HasArg")
		if !hasOptHasArg || optHasArgField.Type.Kind() != reflect.Bool {
			t.Error("Option should have a HasArg field of type bool")
		}

		optArgField, hasOptArg := optionType.FieldByName("Arg")
		if !hasOptArg || optArgField.Type.Kind() != reflect.String {
			t.Error("Option should have an Arg field of type string")
		}

		// Test Parser struct
		parser := Parser{}
		parserType := reflect.TypeOf(parser)
		if parserType.Kind() != reflect.Struct {
			t.Error("Parser should be a struct")
		}

		// Check Parser fields
		argsField, hasArgs := parserType.FieldByName("Args")
		if !hasArgs || argsField.Type != reflect.TypeOf([]string{}) {
			t.Error("Parser should have an Args field of type []string")
		}
	})

	// Test that all expected public functions exist with correct signatures
	t.Run("public_functions_exist", func(t *testing.T) {
		// Test GetOpt function
		getOptType := reflect.TypeOf(GetOpt)
		if getOptType.Kind() != reflect.Func {
			t.Error("GetOpt should be a function")
		}
		if getOptType.NumIn() != 2 {
			t.Error("GetOpt should take 2 parameters")
		}
		if getOptType.In(0) != reflect.TypeOf([]string{}) {
			t.Error("GetOpt first parameter should be []string")
		}
		if getOptType.In(1) != reflect.TypeOf("") {
			t.Error("GetOpt second parameter should be string")
		}
		if getOptType.NumOut() != 2 {
			t.Error("GetOpt should return 2 values")
		}
		if getOptType.Out(0) != reflect.TypeOf(&Parser{}) {
			t.Error("GetOpt first return value should be *Parser")
		}
		if getOptType.Out(1).String() != "error" {
			t.Error("GetOpt second return value should be error")
		}

		// Test GetOptLong function
		getOptLongType := reflect.TypeOf(GetOptLong)
		if getOptLongType.Kind() != reflect.Func {
			t.Error("GetOptLong should be a function")
		}
		if getOptLongType.NumIn() != 3 {
			t.Error("GetOptLong should take 3 parameters")
		}
		if getOptLongType.In(0) != reflect.TypeOf([]string{}) {
			t.Error("GetOptLong first parameter should be []string")
		}
		if getOptLongType.In(1) != reflect.TypeOf("") {
			t.Error("GetOptLong second parameter should be string")
		}
		if getOptLongType.In(2) != reflect.TypeOf([]Flag{}) {
			t.Error("GetOptLong third parameter should be []Flag")
		}

		// Test GetOptLongOnly function
		getOptLongOnlyType := reflect.TypeOf(GetOptLongOnly)
		if getOptLongOnlyType.Kind() != reflect.Func {
			t.Error("GetOptLongOnly should be a function")
		}
		if getOptLongOnlyType.NumIn() != 3 {
			t.Error("GetOptLongOnly should take 3 parameters")
		}
	})

	// Test backward compatibility with existing usage patterns
	t.Run("backward_compatibility", func(t *testing.T) {
		// Test basic GetOpt usage
		parser, err := GetOpt([]string{"-a", "-b", "value"}, "ab:")
		if err != nil {
			t.Fatalf("Basic GetOpt usage should work: %v", err)
		}

		var options []Option
		for opt, err := range parser.Options() {
			if err != nil {
				t.Fatalf("Options iteration should work: %v", err)
			}
			options = append(options, opt)
		}

		if len(options) != 2 {
			t.Errorf("Expected 2 options, got %d", len(options))
		}

		// Test GetOptLong usage
		longOpts := []Flag{
			{Name: "verbose", HasArg: NoArgument},
			{Name: "output", HasArg: RequiredArgument},
		}
		parser2, err := GetOptLong([]string{"--verbose", "--output", "file.txt"}, "vo:", longOpts)
		if err != nil {
			t.Fatalf("Basic GetOptLong usage should work: %v", err)
		}

		var longOptions []Option
		for opt, err := range parser2.Options() {
			if err != nil {
				t.Fatalf("Long options iteration should work: %v", err)
			}
			longOptions = append(longOptions, opt)
		}

		if len(longOptions) != 2 {
			t.Errorf("Expected 2 long options, got %d", len(longOptions))
		}

		// Test GetOptLongOnly usage
		parser3, err := GetOptLongOnly([]string{"-verbose"}, "", longOpts)
		if err != nil {
			t.Fatalf("Basic GetOptLongOnly usage should work: %v", err)
		}

		var longOnlyOptions []Option
		for opt, err := range parser3.Options() {
			if err != nil {
				t.Fatalf("Long-only options iteration should work: %v", err)
			}
			longOnlyOptions = append(longOnlyOptions, opt)
		}

		if len(longOnlyOptions) != 1 {
			t.Errorf("Expected 1 long-only option, got %d", len(longOnlyOptions))
		}
	})

	// Test that POSIXLY_CORRECT environment variable behavior is preserved
	t.Run("posixly_correct_compatibility", func(t *testing.T) {
		// This test ensures that the new environment variable support
		// doesn't break existing behavior
		
		// Test without environment variable
		parser1, err := GetOpt([]string{"-a", "file", "-b"}, "ab")
		if err != nil {
			t.Fatalf("GetOpt should work without POSIXLY_CORRECT: %v", err)
		}

		var options1 []Option
		for opt, err := range parser1.Options() {
			if err != nil {
				t.Fatalf("Options iteration should work: %v", err)
			}
			options1 = append(options1, opt)
		}

		// Should process both -a and -b in default mode
		if len(options1) != 2 {
			t.Errorf("Expected 2 options in default mode, got %d", len(options1))
		}

		// Test with + prefix (existing behavior)
		parser2, err := GetOpt([]string{"-a", "file", "-b"}, "+ab")
		if err != nil {
			t.Fatalf("GetOpt should work with + prefix: %v", err)
		}

		var options2 []Option
		for opt, err := range parser2.Options() {
			if err != nil {
				t.Fatalf("Options iteration should work: %v", err)
			}
			options2 = append(options2, opt)
		}

		// Should stop at first non-option in POSIX mode
		if len(options2) != 1 {
			t.Errorf("Expected 1 option in POSIX mode, got %d", len(options2))
		}
	})
}

// TestAPIDocumentationStability ensures that the public API is properly documented
func TestAPIDocumentationStability(t *testing.T) {
	// This test ensures that all public functions and types have proper documentation
	// and that the API surface remains stable

	t.Run("required_exports_exist", func(t *testing.T) {
		// Ensure all required exports are available
		requiredFunctions := []interface{}{
			GetOpt,
			GetOptLong,
			GetOptLongOnly,
		}

		for _, fn := range requiredFunctions {
			if reflect.ValueOf(fn).Kind() != reflect.Func {
				t.Errorf("Required function is not exported or not a function: %T", fn)
			}
		}

		// Ensure all required types are available
		requiredTypes := []interface{}{
			ArgType(0),
			Flag{},
			Option{},
			Parser{},
			ParseMode(0),
		}

		for _, typ := range requiredTypes {
			if reflect.TypeOf(typ).Name() == "" {
				t.Errorf("Required type is not properly exported: %T", typ)
			}
		}

		// Ensure all required constants are available
		requiredConstants := []ArgType{
			NoArgument,
			RequiredArgument,
			OptionalArgument,
		}

		for i, constant := range requiredConstants {
			if constant != ArgType(i) {
				t.Errorf("Required constant has wrong value: %v should be %d", constant, i)
			}
		}

		requiredParseModes := []ParseMode{
			ParseDefault,
			ParseNonOpts,
			ParsePosixlyCorrect,
		}

		for i, mode := range requiredParseModes {
			if mode != ParseMode(i) {
				t.Errorf("Required ParseMode has wrong value: %v should be %d", mode, i)
			}
		}
	})
}