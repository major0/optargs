package goarg

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// TestScenarioGenerator generates test scenarios from various sources
type TestScenarioGenerator struct {
	scenarios []TestScenarioDefinition
	verbose   bool
}

// NewTestScenarioGenerator creates a new test scenario generator
func NewTestScenarioGenerator() *TestScenarioGenerator {
	return &TestScenarioGenerator{
		scenarios: make([]TestScenarioDefinition, 0),
		verbose:   false,
	}
}

// SetVerbose enables or disables verbose output
func (tsg *TestScenarioGenerator) SetVerbose(verbose bool) {
	tsg.verbose = verbose
}

// GenerateFromUpstreamTests extracts test scenarios from upstream go-arg tests
func (tsg *TestScenarioGenerator) GenerateFromUpstreamTests(upstreamPath string) error {
	if tsg.verbose {
		fmt.Printf("Generating test scenarios from upstream tests at: %s\n", upstreamPath)
	}

	// Find all test files
	testFiles, err := tsg.findTestFiles(upstreamPath)
	if err != nil {
		return fmt.Errorf("failed to find test files: %w", err)
	}

	for _, testFile := range testFiles {
		if err := tsg.parseTestFile(testFile); err != nil {
			if tsg.verbose {
				fmt.Printf("Warning: Failed to parse test file %s: %v\n", testFile, err)
			}
			continue
		}
	}

	if tsg.verbose {
		fmt.Printf("Generated %d test scenarios from upstream tests\n", len(tsg.scenarios))
	}

	return nil
}

// GenerateFromExamples extracts test scenarios from example code
func (tsg *TestScenarioGenerator) GenerateFromExamples(examplesPath string) error {
	if tsg.verbose {
		fmt.Printf("Generating test scenarios from examples at: %s\n", examplesPath)
	}

	// Find all example files
	exampleFiles, err := tsg.findGoFiles(examplesPath)
	if err != nil {
		return fmt.Errorf("failed to find example files: %w", err)
	}

	for _, exampleFile := range exampleFiles {
		if err := tsg.parseExampleFile(exampleFile); err != nil {
			if tsg.verbose {
				fmt.Printf("Warning: Failed to parse example file %s: %v\n", exampleFile, err)
			}
			continue
		}
	}

	if tsg.verbose {
		fmt.Printf("Generated %d additional scenarios from examples\n", len(tsg.scenarios))
	}

	return nil
}

// GenerateFromDocumentation extracts test scenarios from documentation
func (tsg *TestScenarioGenerator) GenerateFromDocumentation(docPath string) error {
	if tsg.verbose {
		fmt.Printf("Generating test scenarios from documentation at: %s\n", docPath)
	}

	// Find README and documentation files
	docFiles, err := tsg.findDocFiles(docPath)
	if err != nil {
		return fmt.Errorf("failed to find documentation files: %w", err)
	}

	for _, docFile := range docFiles {
		if err := tsg.parseDocumentationFile(docFile); err != nil {
			if tsg.verbose {
				fmt.Printf("Warning: Failed to parse documentation file %s: %v\n", docFile, err)
			}
			continue
		}
	}

	if tsg.verbose {
		fmt.Printf("Generated %d additional scenarios from documentation\n", len(tsg.scenarios))
	}

	return nil
}

// GenerateBuiltinScenarios generates built-in test scenarios covering common patterns
func (tsg *TestScenarioGenerator) GenerateBuiltinScenarios() {
	if tsg.verbose {
		fmt.Printf("Generating built-in test scenarios\n")
	}

	builtinScenarios := []TestScenarioDefinition{
		{
			Name:        "basic_bool_flag",
			Description: "Basic boolean flag",
			StructDefinition: `type Args struct {
				Verbose bool ` + "`arg:\"-v,--verbose\"`" + `
			}`,
			Arguments:       []string{"-v"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "string_flag_with_value",
			Description: "String flag with value",
			StructDefinition: `type Args struct {
				Output string ` + "`arg:\"-o,--output\"`" + `
			}`,
			Arguments:       []string{"--output", "file.txt"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "integer_flag",
			Description: "Integer flag",
			StructDefinition: `type Args struct {
				Count int ` + "`arg:\"-c,--count\"`" + `
			}`,
			Arguments:       []string{"-c", "42"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "float_flag",
			Description: "Float flag",
			StructDefinition: `type Args struct {
				Rate float64 ` + "`arg:\"-r,--rate\"`" + `
			}`,
			Arguments:       []string{"--rate", "3.14"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "slice_flag",
			Description: "Slice flag with multiple values",
			StructDefinition: `type Args struct {
				Tags []string ` + "`arg:\"-t,--tag\"`" + `
			}`,
			Arguments:       []string{"-t", "tag1", "-t", "tag2", "--tag", "tag3"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "positional_args",
			Description: "Positional arguments",
			StructDefinition: `type Args struct {
				Files []string ` + "`arg:\"positional\"`" + `
			}`,
			Arguments:       []string{"file1.txt", "file2.txt"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "required_flag_missing",
			Description: "Required flag missing (should error)",
			StructDefinition: `type Args struct {
				Input string ` + "`arg:\"--input,required\"`" + `
			}`,
			Arguments:       []string{},
			ExpectedSuccess: false,
			TestType:        "error",
		},
		{
			Name:        "default_values",
			Description: "Default values",
			StructDefinition: `type Args struct {
				Port int    ` + "`arg:\"--port\" default:\"8080\"`" + `
				Host string ` + "`arg:\"--host\" default:\"localhost\"`" + `
			}`,
			Arguments:       []string{},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "help_flag",
			Description: "Help flag",
			StructDefinition: `type Args struct {
				Verbose bool ` + "`arg:\"-v,--verbose\" help:\"enable verbose output\"`" + `
			}`,
			Arguments:       []string{"--help"},
			ExpectedSuccess: false, // Help exits with non-zero
			TestType:        "help",
		},
		{
			Name:        "unknown_flag_error",
			Description: "Unknown flag error",
			StructDefinition: `type Args struct {
				Verbose bool ` + "`arg:\"-v,--verbose\"`" + `
			}`,
			Arguments:       []string{"--unknown"},
			ExpectedSuccess: false,
			TestType:        "error",
		},
		{
			Name:        "simple_subcommand",
			Description: "Simple subcommand",
			StructDefinition: `type Args struct {
				Server *ServerCmd ` + "`arg:\"subcommand:server\"`" + `
			}

			type ServerCmd struct {
				Port int ` + "`arg:\"-p,--port\" default:\"8080\"`" + `
			}`,
			Arguments:       []string{"server", "--port", "9000"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "nested_subcommands",
			Description: "Nested subcommands",
			StructDefinition: `type Args struct {
				Git *GitCmd ` + "`arg:\"subcommand:git\"`" + `
			}

			type GitCmd struct {
				Remote *RemoteCmd ` + "`arg:\"subcommand:remote\"`" + `
			}

			type RemoteCmd struct {
				Add *RemoteAddCmd ` + "`arg:\"subcommand:add\"`" + `
			}

			type RemoteAddCmd struct {
				Name string ` + "`arg:\"positional,required\"`" + `
				URL  string ` + "`arg:\"positional,required\"`" + `
			}`,
			Arguments:       []string{"git", "remote", "add", "origin", "https://github.com/user/repo.git"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
	}

	tsg.scenarios = append(tsg.scenarios, builtinScenarios...)

	if tsg.verbose {
		fmt.Printf("Generated %d built-in scenarios\n", len(builtinScenarios))
	}
}

// GetScenarios returns all generated scenarios
func (tsg *TestScenarioGenerator) GetScenarios() []TestScenarioDefinition {
	return tsg.scenarios
}

// SaveScenarios saves scenarios to a JSON file
func (tsg *TestScenarioGenerator) SaveScenarios(filename string) error {
	data, err := json.MarshalIndent(tsg.scenarios, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal scenarios: %w", err)
	}

	return ioutil.WriteFile(filename, data, 0644)
}

// LoadScenarios loads scenarios from a JSON file
func (tsg *TestScenarioGenerator) LoadScenarios(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read scenarios file: %w", err)
	}

	return json.Unmarshal(data, &tsg.scenarios)
}

// findTestFiles finds all test files in a directory
func (tsg *TestScenarioGenerator) findTestFiles(dir string) ([]string, error) {
	var testFiles []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, "_test.go") {
			testFiles = append(testFiles, path)
		}

		return nil
	})

	return testFiles, err
}

// findGoFiles finds all Go files in a directory
func (tsg *TestScenarioGenerator) findGoFiles(dir string) ([]string, error) {
	var goFiles []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			goFiles = append(goFiles, path)
		}

		return nil
	})

	return goFiles, err
}

// findDocFiles finds documentation files
func (tsg *TestScenarioGenerator) findDocFiles(dir string) ([]string, error) {
	var docFiles []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		name := strings.ToLower(info.Name())
		if strings.HasSuffix(name, ".md") || strings.HasSuffix(name, ".txt") || name == "readme" {
			docFiles = append(docFiles, path)
		}

		return nil
	})

	return docFiles, err
}

// parseTestFile parses a Go test file and extracts test scenarios
func (tsg *TestScenarioGenerator) parseTestFile(filename string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	// Extract test functions and analyze them
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			if strings.HasPrefix(x.Name.Name, "Test") {
				tsg.extractTestScenario(x, filename)
			}
		}
		return true
	})

	return nil
}

// parseExampleFile parses an example Go file
func (tsg *TestScenarioGenerator) parseExampleFile(filename string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	// Extract struct definitions and usage patterns
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.TypeSpec:
			if structType, ok := x.Type.(*ast.StructType); ok {
				tsg.extractStructScenario(x.Name.Name, structType, filename)
			}
		}
		return true
	})

	return nil
}

// parseDocumentationFile parses documentation files for code examples
func (tsg *TestScenarioGenerator) parseDocumentationFile(filename string) error {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Extract code blocks from markdown
	codeBlockRegex := regexp.MustCompile("```go\\n([\\s\\S]*?)\\n```")
	matches := codeBlockRegex.FindAllStringSubmatch(string(content), -1)

	for i, match := range matches {
		if len(match) > 1 {
			scenarioName := fmt.Sprintf("doc_example_%d", i+1)
			tsg.extractCodeBlockScenario(scenarioName, match[1], filename)
		}
	}

	return nil
}

// extractTestScenario extracts a test scenario from a test function
func (tsg *TestScenarioGenerator) extractTestScenario(funcDecl *ast.FuncDecl, filename string) {
	// This is a simplified extraction - in a real implementation,
	// you would analyze the function body to extract struct definitions,
	// argument patterns, and expected results

	scenarioName := fmt.Sprintf("upstream_%s", strings.ToLower(funcDecl.Name.Name))

	// For now, create a placeholder scenario
	scenario := TestScenarioDefinition{
		Name:        scenarioName,
		Description: fmt.Sprintf("Extracted from %s in %s", funcDecl.Name.Name, filepath.Base(filename)),
		StructDefinition: `type Args struct {
			// TODO: Extract from test function
		}`,
		Arguments:       []string{}, // TODO: Extract from test function
		ExpectedSuccess: true,
		TestType:        "parsing",
		Metadata: map[string]interface{}{
			"source_file":     filename,
			"source_function": funcDecl.Name.Name,
		},
	}

	tsg.scenarios = append(tsg.scenarios, scenario)
}

// extractStructScenario extracts a scenario from a struct definition
func (tsg *TestScenarioGenerator) extractStructScenario(name string, structType *ast.StructType, filename string) {
	// Analyze struct fields and generate appropriate test arguments
	var structDef strings.Builder
	structDef.WriteString(fmt.Sprintf("type %s struct {\n", name))

	var testArgs []string

	for _, field := range structType.Fields.List {
		if len(field.Names) > 0 {
			fieldName := field.Names[0].Name

			// Extract field type and tags
			fieldType := tsg.extractTypeString(field.Type)
			var tagString string
			if field.Tag != nil {
				tagString = field.Tag.Value
			}

			structDef.WriteString(fmt.Sprintf("\t%s %s", fieldName, fieldType))
			if tagString != "" {
				structDef.WriteString(fmt.Sprintf(" %s", tagString))
			}
			structDef.WriteString("\n")

			// Generate test arguments based on field type and tags
			if args := tsg.generateArgsForField(fieldName, fieldType, tagString); len(args) > 0 {
				testArgs = append(testArgs, args...)
			}
		}
	}

	structDef.WriteString("}")

	scenario := TestScenarioDefinition{
		Name:             fmt.Sprintf("struct_%s", strings.ToLower(name)),
		Description:      fmt.Sprintf("Generated from struct %s in %s", name, filepath.Base(filename)),
		StructDefinition: structDef.String(),
		Arguments:        testArgs,
		ExpectedSuccess:  true,
		TestType:         "parsing",
		Metadata: map[string]interface{}{
			"source_file":   filename,
			"source_struct": name,
		},
	}

	tsg.scenarios = append(tsg.scenarios, scenario)
}

// extractCodeBlockScenario extracts a scenario from a code block
func (tsg *TestScenarioGenerator) extractCodeBlockScenario(name, code, filename string) {
	// Parse the code block to extract struct definitions
	fset := token.NewFileSet()

	// Wrap the code in a package to make it parseable
	wrappedCode := fmt.Sprintf("package main\n\n%s", code)

	node, err := parser.ParseFile(fset, "", wrappedCode, parser.ParseComments)
	if err != nil {
		if tsg.verbose {
			fmt.Printf("Warning: Failed to parse code block: %v\n", err)
		}
		return
	}

	// Extract struct definitions
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.TypeSpec:
			if structType, ok := x.Type.(*ast.StructType); ok {
				tsg.extractStructScenario(x.Name.Name, structType, filename)
			}
		}
		return true
	})
}

// extractTypeString extracts a string representation of a type
func (tsg *TestScenarioGenerator) extractTypeString(expr ast.Expr) string {
	switch x := expr.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.ArrayType:
		return "[]" + tsg.extractTypeString(x.Elt)
	case *ast.StarExpr:
		return "*" + tsg.extractTypeString(x.X)
	default:
		return "interface{}"
	}
}

// generateArgsForField generates test arguments for a struct field
func (tsg *TestScenarioGenerator) generateArgsForField(fieldName, fieldType, tagString string) []string {
	// Parse the tag to extract arg information
	argTag := tsg.extractArgTag(tagString)
	if argTag == "" {
		return nil
	}

	// Generate appropriate arguments based on field type
	switch fieldType {
	case "bool":
		if strings.Contains(argTag, "-") {
			return []string{strings.Split(argTag, ",")[0]}
		}
	case "string":
		if strings.Contains(argTag, "-") {
			flag := strings.Split(argTag, ",")[0]
			return []string{flag, "test_value"}
		}
	case "int":
		if strings.Contains(argTag, "-") {
			flag := strings.Split(argTag, ",")[0]
			return []string{flag, "42"}
		}
	case "[]string":
		if strings.Contains(argTag, "-") {
			flag := strings.Split(argTag, ",")[0]
			return []string{flag, "value1", flag, "value2"}
		}
	}

	return nil
}

// extractArgTag extracts the arg tag value from a tag string
func (tsg *TestScenarioGenerator) extractArgTag(tagString string) string {
	if tagString == "" {
		return ""
	}

	// Remove backticks
	tagString = strings.Trim(tagString, "`")

	// Extract arg tag
	argRegex := regexp.MustCompile(`arg:"([^"]*)"`)
	matches := argRegex.FindStringSubmatch(tagString)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

// DownloadUpstreamTests downloads test files from upstream repository
func (tsg *TestScenarioGenerator) DownloadUpstreamTests(repoURL, targetDir string) error {
	if tsg.verbose {
		fmt.Printf("Downloading upstream tests from %s to %s\n", repoURL, targetDir)
	}

	// Create target directory
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Download common test files
	testFiles := []string{
		"arg_test.go",
		"example_test.go",
		"subcommand_test.go",
	}

	for _, testFile := range testFiles {
		url := fmt.Sprintf("%s/raw/master/%s", repoURL, testFile)
		targetPath := filepath.Join(targetDir, testFile)

		if err := tsg.downloadFile(url, targetPath); err != nil {
			if tsg.verbose {
				fmt.Printf("Warning: Failed to download %s: %v\n", testFile, err)
			}
			continue
		}

		if tsg.verbose {
			fmt.Printf("Downloaded %s\n", testFile)
		}
	}

	return nil
}

// downloadFile downloads a file from a URL
func (tsg *TestScenarioGenerator) downloadFile(url, targetPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	return ioutil.WriteFile(targetPath, content, 0644)
}
