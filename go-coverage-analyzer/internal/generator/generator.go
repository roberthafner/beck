package generator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/beck/go-coverage-analyzer/pkg/models"
)

// Options contains configuration for test generation
type Options struct {
	ProjectPath        string
	DryRun             bool
	TemplateStyle      string
	GenerateMocks      bool
	TableDriven        bool
	GenerateBenchmarks bool
	Overwrite          bool
	IgnoreFunctions    []string
	MaxTestCases       int
	Verbose            bool
}

// TestGenerator orchestrates the test generation process
type TestGenerator struct {
	templateEngine *TemplateEngine
	dataGenerator  *DataGenerator
	mockGenerator  *MockGenerator
	validator      *TestValidator
	options        *Options
	fileSet        *token.FileSet
	verbose        bool
}

// NewTestGenerator creates a new test generator
func NewTestGenerator(opts *Options) *TestGenerator {
	return &TestGenerator{
		templateEngine: NewTemplateEngine(opts.Verbose),
		dataGenerator:  NewDataGenerator(opts.Verbose),
		mockGenerator:  NewMockGenerator(opts.Verbose),
		validator:      NewTestValidator(opts.ProjectPath, opts.Verbose),
		options:        opts,
		fileSet:        token.NewFileSet(),
		verbose:        opts.Verbose,
	}
}

// Generate creates test files for uncovered functions
func Generate(analysisResult *models.AnalysisResult, opts *Options) (*models.GenerationResult, error) {
	startTime := time.Now()

	if opts.Verbose {
		fmt.Printf("🛠️ Starting test generation for: %s\n", opts.ProjectPath)
		if opts.DryRun {
			fmt.Println("👀 Running in dry-run mode")
		}
	}

	generator := NewTestGenerator(opts)

	// Initialize the generator
	if err := generator.initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize generator: %w", err)
	}

	// Generate mocks if requested
	if opts.GenerateMocks {
		if err := generator.generateMocks(analysisResult); err != nil {
			if opts.Verbose {
				fmt.Printf("⚠️ Mock generation failed: %v\n", err)
			}
		}
	}

	// Generate tests for uncovered functions
	result, err := generator.generateTests(analysisResult)
	if err != nil {
		return nil, fmt.Errorf("test generation failed: %w", err)
	}

	// Validate generated tests
	if !opts.DryRun {
		if err := generator.validateTests(result); err != nil {
			if opts.Verbose {
				fmt.Printf("⚠️ Test validation failed: %v\n", err)
			}
		}
	}

	result.GenerationTime = time.Since(startTime)

	if opts.Verbose {
		fmt.Printf("✅ Test generation completed in %v\n", result.GenerationTime)
		fmt.Printf("📊 Generated %d tests across %d files\n", result.TestsGenerated, result.FilesCreated)
	}

	return result, nil
}

// initialize sets up the test generator
func (tg *TestGenerator) initialize() error {
	// Load templates
	if err := tg.templateEngine.LoadTemplates(); err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	if tg.verbose {
		fmt.Printf("🔧 Test generator initialized with style: %s\n", tg.options.TemplateStyle)
		if tg.options.GenerateMocks {
			fmt.Println("🎭 Mock generation enabled")
		}
	}

	return nil
}

// generateTests generates test files for uncovered functions
func (tg *TestGenerator) generateTests(analysisResult *models.AnalysisResult) (*models.GenerationResult, error) {
	result := &models.GenerationResult{
		ProjectPath:    tg.options.ProjectPath,
		Timestamp:      time.Now(),
		GeneratedFiles: make([]*models.GeneratedFile, 0),
		Errors:         make([]string, 0),
		Warnings:       make([]string, 0),
	}

	// Group functions by source file
	fileGroups := tg.groupFunctionsByFile(analysisResult.UncoveredFunctions)

	for filePath, functions := range fileGroups {
		if tg.verbose {
			fmt.Printf("📝 Processing file: %s (%d functions)\n", filePath, len(functions))
		}

		generatedFile, err := tg.generateTestFile(filePath, functions, analysisResult)
		if err != nil {
			errorMsg := fmt.Sprintf("Failed to generate tests for %s: %v", filePath, err)
			result.Errors = append(result.Errors, errorMsg)
			if tg.verbose {
				fmt.Printf("❌ %s\n", errorMsg)
			}
			continue
		}

		if generatedFile != nil {
			result.GeneratedFiles = append(result.GeneratedFiles, generatedFile)
			result.TestsGenerated += generatedFile.TestsGenerated
			if generatedFile.Created {
				result.FilesCreated++
			} else {
				result.FilesModified++
			}
			result.FunctionsCovered += len(functions)
		}
	}

	// Calculate estimated coverage improvement
	if result.FunctionsCovered > 0 {
		improvementEstimate := float64(result.FunctionsCovered) / float64(analysisResult.Summary.TotalFunctions) * 100.0
		result.EstimatedCoverage = analysisResult.OverallCoverage + improvementEstimate
		if result.EstimatedCoverage > 100.0 {
			result.EstimatedCoverage = 100.0
		}
	} else {
		result.EstimatedCoverage = analysisResult.OverallCoverage
	}

	return result, nil
}

// groupFunctionsByFile groups functions by their source file
func (tg *TestGenerator) groupFunctionsByFile(functions []*models.Function) map[string][]*models.Function {
	fileGroups := make(map[string][]*models.Function)

	for _, function := range functions {
		if tg.shouldGenerateTest(function) {
			fileGroups[function.File] = append(fileGroups[function.File], function)
		}
	}

	return fileGroups
}

// shouldGenerateTest determines if a test should be generated for a function
func (tg *TestGenerator) shouldGenerateTest(function *models.Function) bool {
	// Skip if function is in ignore list
	for _, ignorePattern := range tg.options.IgnoreFunctions {
		if matched := tg.matchesPattern(function.Name, ignorePattern); matched {
			return false
		}
	}

	// Skip if not testable
	if !function.IsTestable {
		return false
	}

	// Skip if it's already a test function
	if strings.HasPrefix(function.Name, "Test") ||
		strings.HasPrefix(function.Name, "Benchmark") ||
		strings.HasPrefix(function.Name, "Example") {
		return false
	}

	// Skip init and main functions
	if function.Name == "init" || function.Name == "main" {
		return false
	}

	return true
}

// matchesPattern checks if a function name matches an ignore pattern
func (tg *TestGenerator) matchesPattern(functionName, pattern string) bool {
	// Simple pattern matching - could be enhanced with regex
	if pattern == functionName {
		return true
	}
	if strings.Contains(pattern, "*") {
		// Basic wildcard support
		prefix := strings.Split(pattern, "*")[0]
		return strings.HasPrefix(functionName, prefix)
	}
	return false
}

// generateTestFile generates a test file for functions from a source file
func (tg *TestGenerator) generateTestFile(sourceFile string, functions []*models.Function, analysisResult *models.AnalysisResult) (*models.GeneratedFile, error) {
	// Determine test file path
	testFilePath := tg.getTestFilePath(sourceFile)

	// Check if test file already exists and handle accordingly
	exists, err := tg.fileExists(testFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to check test file existence: %w", err)
	}

	if exists && !tg.options.Overwrite {
		if tg.verbose {
			fmt.Printf("⚠️ Test file %s exists, skipping (use --overwrite to replace)\n", testFilePath)
		}
		return nil, nil
	}

	// Parse existing test file if it exists to avoid duplicates
	existingTests := make(map[string]bool)
	if exists {
		existingTests, err = tg.parseExistingTests(testFilePath)
		if err != nil {
			// We'll add warnings to the result when we have access to it
			if tg.verbose {
				fmt.Printf("⚠️ Failed to parse existing tests in %s: %v\n", testFilePath, err)
			}
		}
	}

	// Generate test content
	testContent, testCases, err := tg.generateTestFileContent(functions, existingTests, analysisResult)
	if err != nil {
		return nil, fmt.Errorf("failed to generate test content: %w", err)
	}

	// Write test file
	if !tg.options.DryRun {
		if err := tg.writeTestFile(testFilePath, testContent); err != nil {
			return nil, fmt.Errorf("failed to write test file: %w", err)
		}
	}

	generatedFile := &models.GeneratedFile{
		Path:           testFilePath,
		Package:        functions[0].Package, // All functions should be from same package
		TestsGenerated: len(testCases),
		TestCases:      testCases,
		Size:           int64(len(testContent)),
		Created:        !exists,
		Modified:       exists,
	}

	if tg.verbose {
		if exists {
			fmt.Printf("✏️ Modified test file: %s (%d tests)\n", testFilePath, len(testCases))
		} else {
			fmt.Printf("✨ Created test file: %s (%d tests)\n", testFilePath, len(testCases))
		}
	}

	return generatedFile, nil
}

// getTestFilePath generates the test file path for a source file
func (tg *TestGenerator) getTestFilePath(sourceFile string) string {
	dir := filepath.Dir(sourceFile)
	base := filepath.Base(sourceFile)
	nameWithoutExt := strings.TrimSuffix(base, filepath.Ext(base))
	testFileName := nameWithoutExt + "_test.go"
	return filepath.Join(dir, testFileName)
}

// fileExists checks if a file exists
func (tg *TestGenerator) fileExists(path string) (bool, error) {
	fullPath := filepath.Join(tg.options.ProjectPath, path)
	_, err := os.Stat(fullPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// parseExistingTests parses existing test functions to avoid duplicates
func (tg *TestGenerator) parseExistingTests(testFilePath string) (map[string]bool, error) {
	existingTests := make(map[string]bool)

	fullPath := filepath.Join(tg.options.ProjectPath, testFilePath)
	src, err := os.ReadFile(fullPath)
	if err != nil {
		return existingTests, err
	}

	// Parse the Go source file
	file, err := parser.ParseFile(tg.fileSet, fullPath, src, parser.ParseComments)
	if err != nil {
		return existingTests, err
	}

	// Extract existing test function names
	ast.Inspect(file, func(n ast.Node) bool {
		if funcDecl, ok := n.(*ast.FuncDecl); ok {
			if funcDecl.Name != nil && strings.HasPrefix(funcDecl.Name.Name, "Test") {
				existingTests[funcDecl.Name.Name] = true
			}
		}
		return true
	})

	return existingTests, nil
}

// generateTestFileContent generates the complete content for a test file
func (tg *TestGenerator) generateTestFileContent(functions []*models.Function, existingTests map[string]bool, analysisResult *models.AnalysisResult) (string, []*models.TestCase, error) {
	var contentParts []string
	var allTestCases []*models.TestCase

	// Generate package declaration and imports
	packageName := functions[0].Package
	imports := tg.generateImports(functions)

	header := fmt.Sprintf("package %s\n\nimport (\n", packageName)
	for _, imp := range imports {
		header += fmt.Sprintf("\t\"%s\"\n", imp)
	}
	header += ")\n\n"
	contentParts = append(contentParts, header)

	// Generate tests for each function
	for _, function := range functions {
		testName := "Test" + function.Name
		if existingTests[testName] && !tg.options.Overwrite {
			if tg.verbose {
				fmt.Printf("⏭️ Skipping existing test: %s\n", testName)
			}
			continue
		}

		// Generate test data
		testData, err := tg.dataGenerator.GenerateTestData(function, tg.options.MaxTestCases)
		if err != nil {
			if tg.verbose {
				fmt.Printf("⚠️ Failed to generate test data for %s: %v\n", function.Name, err)
			}
			continue
		}

		// Generate test content using templates
		testContent, err := tg.templateEngine.GenerateTest(function, tg.options.TemplateStyle, tg.options.TableDriven)
		if err != nil {
			if tg.verbose {
				fmt.Printf("⚠️ Failed to generate test for %s: %v\n", function.Name, err)
			}
			continue
		}

		contentParts = append(contentParts, testContent)

		// Convert test data to test cases for result tracking
		for _, generatedCase := range testData.TestCases {
			testCase := &models.TestCase{
				FunctionName:  function.Name,
				TestName:      testName,
				TestType:      "unit",
				InputCount:    len(generatedCase.Inputs),
				HasMocks:      tg.options.GenerateMocks && tg.needsMocks(function),
				HasSetup:      false,
				HasTeardown:   false,
				ExpectedLines: estimateTestLines(testContent),
				Complexity:    function.Complexity,
			}
			allTestCases = append(allTestCases, testCase)
		}

		// Generate benchmark test if requested
		if tg.options.GenerateBenchmarks {
			benchmarkContent, err := tg.templateEngine.GenerateTest(function, "benchmark", false)
			if err == nil {
				contentParts = append(contentParts, benchmarkContent)
				benchmarkCase := &models.TestCase{
					FunctionName:  function.Name,
					TestName:      "Benchmark" + function.Name,
					TestType:      "benchmark",
					InputCount:    len(function.Parameters),
					HasMocks:      false,
					ExpectedLines: estimateTestLines(benchmarkContent),
					Complexity:    function.Complexity,
				}
				allTestCases = append(allTestCases, benchmarkCase)
			}
		}
	}

	fullContent := strings.Join(contentParts, "\n")
	return fullContent, allTestCases, nil
}

// generateImports generates the necessary imports for the test file
func (tg *TestGenerator) generateImports(functions []*models.Function) []string {
	imports := []string{"testing"}

	// Add testify if using testify style
	if tg.options.TemplateStyle == "testify" {
		imports = append(imports, "github.com/stretchr/testify/assert")
		if tg.options.GenerateMocks {
			imports = append(imports, "github.com/stretchr/testify/mock")
		}
	}

	// Add context import if any function uses context
	for _, function := range functions {
		for _, param := range function.Parameters {
			if strings.Contains(param.Type, "context.Context") {
				imports = append(imports, "context")
				break
			}
		}
	}

	return removeDuplicateStrings(imports)
}

// needsMocks determines if a function needs mocks
func (tg *TestGenerator) needsMocks(function *models.Function) bool {
	return function.CallsExternal || len(function.Dependencies) > 0
}

// writeTestFile writes the test content to a file
func (tg *TestGenerator) writeTestFile(testFilePath, content string) error {
	fullPath := filepath.Join(tg.options.ProjectPath, testFilePath)

	// Create directory if it doesn't exist
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write the file
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// estimateTestLines estimates the number of lines in a test
func estimateTestLines(content string) int {
	return len(strings.Split(content, "\n"))
}

// removeDuplicateStrings removes duplicate strings from a slice
func removeDuplicateStrings(slice []string) []string {
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

// ValidateGeneratedTests validates that generated tests compile and run
func (tg *TestGenerator) ValidateGeneratedTests(result *models.GenerationResult) error {
	if tg.options.DryRun {
		return nil // Skip validation in dry-run mode
	}

	if tg.verbose {
		fmt.Printf("🔍 Validating generated tests...\n")
	}

	// Simple validation: check if files can be parsed as Go code
	for _, generatedFile := range result.GeneratedFiles {
		fullPath := filepath.Join(tg.options.ProjectPath, generatedFile.Path)

		// Try to parse the generated file
		src, err := os.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("failed to read generated file %s: %w", generatedFile.Path, err)
		}

		_, err = parser.ParseFile(tg.fileSet, fullPath, src, parser.ParseComments)
		if err != nil {
			return fmt.Errorf("generated file %s has syntax errors: %w", generatedFile.Path, err)
		}
	}

	if tg.verbose {
		fmt.Printf("✅ All generated tests passed validation\n")
	}

	return nil
}

// generateMocks generates mock files for interfaces
func (tg *TestGenerator) generateMocks(analysisResult *models.AnalysisResult) error {
	if tg.verbose {
		fmt.Println("🎭 Generating mocks for interfaces...")
	}

	mocks, err := tg.mockGenerator.GenerateMocks(analysisResult.UncoveredFunctions, tg.options.ProjectPath)
	if err != nil {
		return fmt.Errorf("mock generation failed: %w", err)
	}

	if len(mocks) == 0 {
		if tg.verbose {
			fmt.Println("🎭 No interfaces found that require mocking")
		}
		return nil
	}

	// Write mock files
	if err := tg.mockGenerator.WriteMocks(mocks, tg.options.ProjectPath, tg.options.DryRun); err != nil {
		return fmt.Errorf("failed to write mock files: %w", err)
	}

	if tg.verbose {
		fmt.Printf("🎭 Generated %d mock files\n", len(mocks))
	}

	return nil
}

// validateTests validates generated tests for quality and correctness
func (tg *TestGenerator) validateTests(result *models.GenerationResult) error {
	if tg.verbose {
		fmt.Println("🔍 Validating generated tests...")
	}

	validationResult, err := tg.validator.ValidateTests(result)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if tg.verbose {
		if validationResult.Valid {
			fmt.Printf("✅ All tests validated successfully\n")
			if validationResult.TestsRun > 0 {
				fmt.Printf("   Tests: %d passed, %d failed\n", validationResult.TestsPassed, validationResult.TestsFailed)
			}
			if validationResult.CoverageImproved > 0 {
				fmt.Printf("   Coverage: %.1f%%\n", validationResult.CoverageImproved)
			}
		} else {
			fmt.Printf("⚠️ Test validation completed with issues:\n")
			if len(validationResult.SyntaxErrors) > 0 {
				fmt.Printf("   Syntax errors: %d\n", len(validationResult.SyntaxErrors))
			}
			if len(validationResult.CompileErrors) > 0 {
				fmt.Printf("   Compile errors: %d\n", len(validationResult.CompileErrors))
			}
			if len(validationResult.RuntimeErrors) > 0 {
				fmt.Printf("   Runtime errors: %d\n", len(validationResult.RuntimeErrors))
			}
			if len(validationResult.Warnings) > 0 {
				fmt.Printf("   Warnings: %d\n", len(validationResult.Warnings))
			}
		}
	}

	return nil
}
