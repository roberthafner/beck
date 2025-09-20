package generator

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/beck/go-coverage-analyzer/pkg/models"
)

// TemplateEngine handles test template processing
type TemplateEngine struct {
	templates map[string]*template.Template
	verbose   bool
}

// NewTemplateEngine creates a new template engine
func NewTemplateEngine(verbose bool) *TemplateEngine {
	return &TemplateEngine{
		templates: make(map[string]*template.Template),
		verbose:   verbose,
	}
}

// TemplateData contains data for template rendering
type TemplateData struct {
	PackageName    string
	Function       *models.Function
	TestName       string
	Imports        []string
	TestCases      []TestCaseData
	HasMocks       bool
	MockStructs    []MockData
	SetupCode      string
	TeardownCode   string
	TableDriven    bool
	BenchmarkTest  bool
	AssertionStyle string // "testing", "testify", "assert"
	ProjectPath    string
	FileName       string
	Comment        string
}

// TestCaseData represents a single test case
type TestCaseData struct {
	Name           string
	Description    string
	Inputs         []InputData
	ExpectedOutput []OutputData
	ExpectError    bool
	ErrorMessage   string
	MockSetup      string
	Comment        string
}

// InputData represents function input parameters
type InputData struct {
	Name  string
	Value string
	Type  string
}

// OutputData represents expected function outputs
type OutputData struct {
	Value string
	Type  string
}

// MockData represents mock generation data
type MockData struct {
	InterfaceName string
	Methods       []MockMethodData
	PackageName   string
}

// MockMethodData represents a mock method
type MockMethodData struct {
	Name       string
	Signature  string
	ReturnType string
}

// LoadTemplates loads templates from external files first, then falls back to built-in templates
func (te *TemplateEngine) LoadTemplates() error {
	// Try to load external templates first
	externalTemplates := te.loadExternalTemplates()

	// Built-in templates as fallback
	builtinTemplates := map[string]string{
		"function_test":  functionTestTemplate,
		"table_test":     tableTestTemplate,
		"benchmark_test": benchmarkTestTemplate,
		"testify_test":   testifyTestTemplate,
		"method_test":    methodTestTemplate,
		"error_test":     errorTestTemplate,
		"mock_interface": mockInterfaceTemplate,
		"file_header":    fileHeaderTemplate,
	}

	// Merge external and built-in templates (external takes precedence)
	templates := make(map[string]string)
	for name, content := range builtinTemplates {
		templates[name] = content
	}
	for name, content := range externalTemplates {
		templates[name] = content
	}

	funcMap := template.FuncMap{
		"title":                  strings.Title,
		"lower":                  strings.ToLower,
		"upper":                  strings.ToUpper,
		"camelCase":              toCamelCase,
		"snakeCase":              toSnakeCase,
		"join":                   strings.Join,
		"hasPrefix":              strings.HasPrefix,
		"hasSuffix":              strings.HasSuffix,
		"replace":                strings.ReplaceAll,
		"trimPrefix":             strings.TrimPrefix,
		"quote":                  func(s string) string { return fmt.Sprintf("%q", s) },
		"generateValue":          generateTestValue,
		"generateBenchmarkValue": generateBenchmarkValue,
		"isBasicType":            isBasicType,
		"isSliceType":            isSliceType,
		"isMapType":              isMapType,
		"baseType":               getBaseType,
		"zeroValue":              getZeroValue,
	}

	for name, tmplContent := range templates {
		tmpl, err := template.New(name).Funcs(funcMap).Parse(tmplContent)
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", name, err)
		}
		te.templates[name] = tmpl
	}

	if te.verbose {
		fmt.Printf("📋 Loaded %d test templates\n", len(templates))
	}

	return nil
}

// GenerateTest generates a test for a specific function
func (te *TemplateEngine) GenerateTest(function *models.Function, style string, tableStyle bool) (string, error) {
	templateName := te.selectTemplate(function, style, tableStyle)
	tmpl, exists := te.templates[templateName]
	if !exists {
		return "", fmt.Errorf("template not found: %s", templateName)
	}

	data := te.buildTemplateData(function, style, tableStyle)

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// selectTemplate chooses the appropriate template based on function and style
func (te *TemplateEngine) selectTemplate(function *models.Function, style string, tableStyle bool) string {
	if tableStyle && len(function.Parameters) > 1 {
		return "table_test"
	}

	switch style {
	case "testify":
		return "testify_test"
	case "benchmark":
		return "benchmark_test"
	default:
		if function.IsMethod {
			return "method_test"
		}
		if function.HasErrorReturn {
			return "error_test"
		}
		return "function_test"
	}
}

// buildTemplateData constructs template data for a function
func (te *TemplateEngine) buildTemplateData(function *models.Function, style string, tableStyle bool) *TemplateData {
	data := &TemplateData{
		PackageName:    function.Package,
		Function:       function,
		TestName:       fmt.Sprintf("Test%s", function.Name),
		Imports:        te.generateImports(function, style),
		TableDriven:    tableStyle,
		AssertionStyle: style,
		FileName:       filepath.Base(function.File),
		Comment:        fmt.Sprintf("// %s tests the %s function\n", fmt.Sprintf("Test%s", function.Name), function.Name),
	}

	// Generate test cases
	data.TestCases = te.generateTestCases(function, style)

	// Add mocks if needed
	if te.needsMocks(function) {
		data.HasMocks = true
		data.MockStructs = te.generateMockData(function)
	}

	return data
}

// generateImports creates necessary import statements
func (te *TemplateEngine) generateImports(function *models.Function, style string) []string {
	imports := []string{"testing"}

	// Add testify imports if using testify style
	if style == "testify" {
		imports = append(imports, "github.com/stretchr/testify/assert")
		if te.needsMocks(function) {
			imports = append(imports, "github.com/stretchr/testify/mock")
		}
	}

	// Add context if function uses it
	if te.usesContext(function) {
		imports = append(imports, "context")
	}

	// Add other dependencies based on function signature
	imports = append(imports, te.extractImports(function)...)

	return removeDuplicates(imports)
}

// generateTestCases creates test case data for the function
func (te *TemplateEngine) generateTestCases(function *models.Function, style string) []TestCaseData {
	var testCases []TestCaseData

	// Generate basic positive test case
	testCases = append(testCases, te.generatePositiveTestCase(function))

	// Generate edge cases
	testCases = append(testCases, te.generateEdgeCases(function)...)

	// Generate error cases if function returns error
	if function.HasErrorReturn {
		testCases = append(testCases, te.generateErrorCases(function)...)
	}

	return testCases
}

// generatePositiveTestCase creates a basic positive test case
func (te *TemplateEngine) generatePositiveTestCase(function *models.Function) TestCaseData {
	testCase := TestCaseData{
		Name:           "positive_case",
		Description:    "Test basic functionality",
		Inputs:         make([]InputData, 0),
		ExpectedOutput: make([]OutputData, 0),
	}

	// Generate inputs
	for _, param := range function.Parameters {
		input := InputData{
			Name:  param.Name,
			Type:  param.Type,
			Value: generateTestValue(param.Type, "positive"),
		}
		testCase.Inputs = append(testCase.Inputs, input)
	}

	// Generate expected outputs
	for _, returnType := range function.ReturnTypes {
		if returnType == "error" {
			output := OutputData{
				Value: "nil",
				Type:  "error",
			}
			testCase.ExpectedOutput = append(testCase.ExpectedOutput, output)
		} else {
			output := OutputData{
				Value: generateExpectedValue(returnType, function, testCase.Inputs),
				Type:  returnType,
			}
			testCase.ExpectedOutput = append(testCase.ExpectedOutput, output)
		}
	}

	return testCase
}

// generateEdgeCases creates edge case test scenarios
func (te *TemplateEngine) generateEdgeCases(function *models.Function) []TestCaseData {
	var testCases []TestCaseData

	// Generate edge cases based on parameter types
	for _, param := range function.Parameters {
		if edgeCase := te.generateEdgeCaseForType(param, function); edgeCase != nil {
			testCases = append(testCases, *edgeCase)
		}
	}

	return testCases
}

// generateEdgeCaseForType creates edge cases for specific types
func (te *TemplateEngine) generateEdgeCaseForType(param *models.Param, function *models.Function) *TestCaseData {
	switch param.Type {
	case "string":
		return &TestCaseData{
			Name:        "empty_string",
			Description: fmt.Sprintf("Test with empty %s", param.Name),
			Inputs: []InputData{{
				Name:  param.Name,
				Type:  param.Type,
				Value: `""`,
			}},
		}
	case "int", "int32", "int64":
		return &TestCaseData{
			Name:        "zero_value",
			Description: fmt.Sprintf("Test with zero %s", param.Name),
			Inputs: []InputData{{
				Name:  param.Name,
				Type:  param.Type,
				Value: "0",
			}},
		}
	case "[]string", "[]int":
		return &TestCaseData{
			Name:        "empty_slice",
			Description: fmt.Sprintf("Test with empty %s", param.Name),
			Inputs: []InputData{{
				Name:  param.Name,
				Type:  param.Type,
				Value: "nil",
			}},
		}
	}
	return nil
}

// generateErrorCases creates error test scenarios
func (te *TemplateEngine) generateErrorCases(function *models.Function) []TestCaseData {
	var testCases []TestCaseData

	// Generate error cases based on function signature
	if strings.Contains(strings.ToLower(function.Name), "divide") {
		testCases = append(testCases, TestCaseData{
			Name:         "division_by_zero",
			Description:  "Test division by zero error",
			ExpectError:  true,
			ErrorMessage: "division by zero",
		})
	}

	if strings.Contains(strings.ToLower(function.Name), "parse") ||
		strings.Contains(strings.ToLower(function.Name), "convert") {
		testCases = append(testCases, TestCaseData{
			Name:         "invalid_input",
			Description:  "Test with invalid input",
			ExpectError:  true,
			ErrorMessage: "invalid",
		})
	}

	return testCases
}

// Helper functions for templates

func generateTestValue(paramType, scenario string) string {
	switch paramType {
	case "string":
		if scenario == "positive" {
			return `"test"`
		}
		return `""`
	case "int", "int32", "int64":
		if scenario == "positive" {
			return "42"
		}
		return "0"
	case "float32", "float64":
		if scenario == "positive" {
			return "3.14"
		}
		return "0.0"
	case "bool":
		return "true"
	case "[]string":
		if scenario == "positive" {
			return `[]string{"item1", "item2"}`
		}
		return "nil"
	case "[]int":
		if scenario == "positive" {
			return `[]int{1, 2, 3}`
		}
		return "nil"
	default:
		if strings.HasPrefix(paramType, "*") {
			return "nil"
		}
		return fmt.Sprintf("%s{}", paramType)
	}
}

func generateExpectedValue(returnType string, function *models.Function, inputs []InputData) string {
	// Simple heuristics for expected values based on function name and inputs
	switch returnType {
	case "string":
		if strings.Contains(strings.ToLower(function.Name), "name") {
			return `"expected_name"`
		}
		return `"result"`
	case "int", "int32", "int64":
		if strings.Contains(strings.ToLower(function.Name), "add") && len(inputs) == 2 {
			return "expected_sum" // Template will handle this
		}
		return "42"
	case "bool":
		return "true"
	case "float32", "float64":
		return "3.14"
	default:
		return getZeroValue(returnType)
	}
}

func isBasicType(t string) bool {
	basicTypes := []string{"string", "int", "int32", "int64", "float32", "float64", "bool", "byte", "rune"}
	for _, bt := range basicTypes {
		if t == bt {
			return true
		}
	}
	return false
}

func isSliceType(t string) bool {
	return strings.HasPrefix(t, "[]")
}

func isMapType(t string) bool {
	return strings.HasPrefix(t, "map[")
}

func getBaseType(t string) string {
	if strings.HasPrefix(t, "[]") {
		return t[2:]
	}
	if strings.HasPrefix(t, "*") {
		return t[1:]
	}
	return t
}

func getZeroValue(t string) string {
	switch t {
	case "string":
		return `""`
	case "int", "int32", "int64", "int8", "int16":
		return "0"
	case "uint", "uint32", "uint64", "uint8", "uint16":
		return "0"
	case "float32", "float64":
		return "0.0"
	case "bool":
		return "false"
	case "byte":
		return "0"
	case "rune":
		return "0"
	default:
		if strings.HasPrefix(t, "[]") || strings.HasPrefix(t, "map[") || strings.HasPrefix(t, "*") {
			return "nil"
		}
		return fmt.Sprintf("%s{}", t)
	}
}

func toCamelCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func toSnakeCase(s string) string {
	return strings.ToLower(s)
}

func (te *TemplateEngine) needsMocks(function *models.Function) bool {
	// Check if function has interface parameters or calls external dependencies
	return function.CallsExternal || len(function.Dependencies) > 0
}

func (te *TemplateEngine) usesContext(function *models.Function) bool {
	for _, param := range function.Parameters {
		if strings.Contains(param.Type, "context.Context") {
			return true
		}
	}
	return false
}

func (te *TemplateEngine) extractImports(function *models.Function) []string {
	var imports []string
	// Extract imports from parameter types and return types
	// This is a simplified version - in practice, you'd need more sophisticated parsing
	return imports
}

func (te *TemplateEngine) generateMockData(function *models.Function) []MockData {
	// Generate mock data for interfaces used by the function
	var mocks []MockData
	// Implementation would analyze function dependencies and generate appropriate mocks
	return mocks
}

func removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	var result []string

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}

// loadExternalTemplates loads templates from external template files
func (te *TemplateEngine) loadExternalTemplates() map[string]string {
	templates := make(map[string]string)

	templateDirs := []string{"templates/standard", "templates/testify", "templates/table"}

	for _, dir := range templateDirs {
		files, err := filepath.Glob(filepath.Join(dir, "*.tmpl"))
		if err != nil {
			continue
		}

		for _, file := range files {
			content, err := os.ReadFile(file)
			if err != nil {
				continue
			}

			// Extract template name from file path
			baseName := filepath.Base(file)
			templateName := strings.TrimSuffix(baseName, ".tmpl")
			dirName := filepath.Base(filepath.Dir(file))
			fullName := dirName + "_" + templateName

			templates[fullName] = string(content)

			if te.verbose {
				fmt.Printf("📋 Loaded external template: %s\n", fullName)
			}
		}
	}

	return templates
}

// generateBenchmarkValue generates appropriate values for benchmark tests
func generateBenchmarkValue(typeName string) string {
	switch typeName {
	case "string":
		return `"benchmark_string"`
	case "int", "int8", "int16", "int32", "int64":
		return "100"
	case "uint", "uint8", "uint16", "uint32", "uint64":
		return "100"
	case "float32", "float64":
		return "3.14"
	case "bool":
		return "true"
	case "[]string":
		return `[]string{"benchmark", "test"}`
	case "[]int":
		return "[]int{1, 2, 3, 4, 5}"
	default:
		if strings.HasPrefix(typeName, "[]") {
			return "nil"
		}
		if strings.HasPrefix(typeName, "*") {
			return "nil"
		}
		return "nil"
	}
}

const functionTestTemplate = `{{.Comment}}func {{.TestName}}(t *testing.T) {
	{{if .HasMocks}}// Setup mocks
	{{range .MockStructs}}{{.Name}} := &Mock{{.InterfaceName}}{}
	{{end}}{{end}}

	{{if .TableDriven}}tests := []struct {
		name string
		{{range .Function.Parameters}}{{.Name}} {{.Type}}
		{{end}}{{if .Function.ReturnTypes}}want {{index .Function.ReturnTypes 0}}{{end}}
		{{if .Function.HasErrorReturn}}wantErr bool{{end}}
	}{
		{{range .TestCases}}{
			name: "{{.Name}}",
			{{range .Inputs}}{{.Name}}: {{.Value}},
			{{end}}{{range .ExpectedOutput}}want: {{.Value}},
			{{end}}{{if .ExpectError}}wantErr: true,{{end}}
		},
		{{end}}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			{{if .Function.IsMethod}}receiver := &{{.Function.ReceiverType}}{}
			{{if .Function.ReturnTypes}}got{{if .Function.HasErrorReturn}}, err{{end}} := receiver.{{.Function.Name}}({{range $i, $param := .Function.Parameters}}{{if $i}}, {{end}}tt.{{$param.Name}}{{end}}){{else}}receiver.{{.Function.Name}}({{range $i, $param := .Function.Parameters}}{{if $i}}, {{end}}tt.{{$param.Name}}{{end}}){{end}}{{else}}{{if .Function.ReturnTypes}}got{{if .Function.HasErrorReturn}}, err{{end}} := {{.Function.Name}}({{range $i, $param := .Function.Parameters}}{{if $i}}, {{end}}tt.{{$param.Name}}{{end}}){{else}}{{.Function.Name}}({{range $i, $param := .Function.Parameters}}{{if $i}}, {{end}}tt.{{$param.Name}}{{end}}){{end}}{{end}}

			{{if .Function.HasErrorReturn}}if (err != nil) != tt.wantErr {
				t.Errorf("{{.Function.Name}}() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("{{.Function.Name}}() = %v, want %v", got, tt.want)
			}{{else if .Function.ReturnTypes}}if got != tt.want {
				t.Errorf("{{.Function.Name}}() = %v, want %v", got, tt.want)
			}{{end}}
		})
	}{{else}}// Test case
	{{range $i, $case := .TestCases}}{{if $i}}
	t.Run("{{$case.Name}}", func(t *testing.T) {
		{{end}}{{if $.Function.IsMethod}}receiver := &{{$.Function.ReceiverType}}{}
		{{if $.Function.ReturnTypes}}got{{if $.Function.HasErrorReturn}}, err{{end}} := receiver.{{$.Function.Name}}({{range $j, $input := $case.Inputs}}{{if $j}}, {{end}}{{$input.Value}}{{end}}){{else}}receiver.{{$.Function.Name}}({{range $j, $input := $case.Inputs}}{{if $j}}, {{end}}{{$input.Value}}{{end}}){{end}}{{else}}{{if $.Function.ReturnTypes}}got{{if $.Function.HasErrorReturn}}, err{{end}} := {{$.Function.Name}}({{range $j, $input := $case.Inputs}}{{if $j}}, {{end}}{{$input.Value}}{{end}}){{else}}{{$.Function.Name}}({{range $j, $input := $case.Inputs}}{{if $j}}, {{end}}{{$input.Value}}{{end}}){{end}}{{end}}

		{{if $.Function.HasErrorReturn}}{{if $case.ExpectError}}if err == nil {
			t.Errorf("{{$.Function.Name}}() expected error but got none")
		}{{else}}if err != nil {
			t.Errorf("{{$.Function.Name}}() unexpected error: %v", err)
		}{{end}}{{end}}
		{{if and $.Function.ReturnTypes (not $case.ExpectError)}}{{range $k, $output := $case.ExpectedOutput}}if got != {{$output.Value}} {
			t.Errorf("{{$.Function.Name}}() = %v, want %v", got, {{$output.Value}})
		}{{end}}{{end}}{{if $i}}
	}){{end}}{{end}}{{end}}
}`

const testifyTestTemplate = `{{.Comment}}func {{.TestName}}(t *testing.T) {
	{{range .TestCases}}t.Run("{{.Name}}", func(t *testing.T) {
		// Arrange
		{{range .Inputs}}{{.Name}} := {{.Value}}
		{{end}}

		// Act
		{{if .Function.IsMethod}}receiver := &{{.Function.ReceiverType}}{}
		{{if .Function.ReturnTypes}}result{{if .Function.HasErrorReturn}}, err{{end}} := receiver.{{.Function.Name}}({{range $i, $input := .Inputs}}{{if $i}}, {{end}}{{$input.Name}}{{end}}){{else}}receiver.{{.Function.Name}}({{range $i, $input := .Inputs}}{{if $i}}, {{end}}{{$input.Name}}{{end}}){{end}}{{else}}{{if .Function.ReturnTypes}}result{{if .Function.HasErrorReturn}}, err{{end}} := {{.Function.Name}}({{range $i, $input := .Inputs}}{{if $i}}, {{end}}{{$input.Name}}{{end}}){{else}}{{.Function.Name}}({{range $i, $input := .Inputs}}{{if $i}}, {{end}}{{$input.Name}}{{end}}){{end}}{{end}}

		// Assert
		{{if .ExpectError}}assert.Error(t, err){{else}}{{if .Function.HasErrorReturn}}assert.NoError(t, err){{end}}{{end}}
		{{if and .Function.ReturnTypes (not .ExpectError)}}{{range .ExpectedOutput}}assert.Equal(t, {{.Value}}, result){{end}}{{end}}
	})
	{{end}}
}`

const tableTestTemplate = functionTestTemplate

const benchmarkTestTemplate = `func Benchmark{{.Function.Name}}(b *testing.B) {
	{{range .Function.Parameters}}{{.Name}} := {{generateValue .Type "positive"}}
	{{end}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		{{if .Function.IsMethod}}receiver := &{{.Function.ReceiverType}}{}
		{{if .Function.ReturnTypes}}_ {{if .Function.HasErrorReturn}}, _{{end}} = receiver.{{.Function.Name}}({{range $i, $param := .Function.Parameters}}{{if $i}}, {{end}}{{$param.Name}}{{end}}){{else}}receiver.{{.Function.Name}}({{range $i, $param := .Function.Parameters}}{{if $i}}, {{end}}{{$param.Name}}{{end}}){{end}}{{else}}{{if .Function.ReturnTypes}}_ {{if .Function.HasErrorReturn}}, _{{end}} = {{.Function.Name}}({{range $i, $param := .Function.Parameters}}{{if $i}}, {{end}}{{$param.Name}}{{end}}){{else}}{{.Function.Name}}({{range $i, $param := .Function.Parameters}}{{if $i}}, {{end}}{{$param.Name}}{{end}}){{end}}{{end}}
	}
}`

const methodTestTemplate = functionTestTemplate

const errorTestTemplate = functionTestTemplate

const mockInterfaceTemplate = `// Mock{{.InterfaceName}} is a mock implementation of {{.InterfaceName}}
type Mock{{.InterfaceName}} struct {
	mock.Mock
}

{{range .Methods}}func (m *Mock{{$.InterfaceName}}) {{.Name}}{{.Signature}} {{.ReturnType}} {
	args := m.Called({{.Args}})
	return args.{{.ReturnCall}}
}

{{end}}`

const fileHeaderTemplate = `package {{.PackageName}}

import (
{{range .Imports}}	"{{.}}"
{{end}}
)
`
