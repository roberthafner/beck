package generator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/beck/go-coverage-analyzer/pkg/models"
)

// ValidationResult represents the result of test validation
type ValidationResult struct {
	Valid            bool          `json:"valid"`
	CompilationTime  time.Duration `json:"compilation_time"`
	ExecutionTime    time.Duration `json:"execution_time"`
	TestsRun         int           `json:"tests_run"`
	TestsPassed      int           `json:"tests_passed"`
	TestsFailed      int           `json:"tests_failed"`
	CoverageImproved float64       `json:"coverage_improved"`
	SyntaxErrors     []string      `json:"syntax_errors,omitempty"`
	CompileErrors    []string      `json:"compile_errors,omitempty"`
	RuntimeErrors    []string      `json:"runtime_errors,omitempty"`
	Warnings         []string      `json:"warnings,omitempty"`
}

// TestValidator validates generated tests for correctness and quality
type TestValidator struct {
	fileSet     *token.FileSet
	projectPath string
	verbose     bool
}

// NewTestValidator creates a new test validator
func NewTestValidator(projectPath string, verbose bool) *TestValidator {
	return &TestValidator{
		fileSet:     token.NewFileSet(),
		projectPath: projectPath,
		verbose:     verbose,
	}
}

// ValidateTests performs comprehensive validation of generated tests
func (tv *TestValidator) ValidateTests(result *models.GenerationResult) (*ValidationResult, error) {
	if tv.verbose {
		fmt.Println("🔍 Starting test validation...")
	}

	startTime := time.Now()
	validationResult := &ValidationResult{
		Valid:         true,
		SyntaxErrors:  make([]string, 0),
		CompileErrors: make([]string, 0),
		RuntimeErrors: make([]string, 0),
		Warnings:      make([]string, 0),
	}

	// Step 1: Syntax validation
	if tv.verbose {
		fmt.Println("🔍 Validating syntax...")
	}

	syntaxValid := tv.validateSyntax(result, validationResult)
	if !syntaxValid {
		validationResult.Valid = false
		if tv.verbose {
			fmt.Printf("❌ Syntax validation failed with %d errors\n", len(validationResult.SyntaxErrors))
		}
		return validationResult, nil
	}

	// Step 2: Compilation validation
	if tv.verbose {
		fmt.Println("🔍 Validating compilation...")
	}

	compileStart := time.Now()
	compileValid := tv.validateCompilation(result, validationResult)
	validationResult.CompilationTime = time.Since(compileStart)

	if !compileValid {
		validationResult.Valid = false
		if tv.verbose {
			fmt.Printf("❌ Compilation validation failed with %d errors\n", len(validationResult.CompileErrors))
		}
		return validationResult, nil
	}

	// Step 3: Execution validation (if compilation passed)
	if tv.verbose {
		fmt.Println("🔍 Validating test execution...")
	}

	execStart := time.Now()
	execValid := tv.validateExecution(result, validationResult)
	validationResult.ExecutionTime = time.Since(execStart)

	if !execValid {
		validationResult.Valid = false
		if tv.verbose {
			fmt.Printf("❌ Execution validation failed with %d errors\n", len(validationResult.RuntimeErrors))
		}
	}

	// Step 4: Quality checks
	if tv.verbose {
		fmt.Println("🔍 Running quality checks...")
	}

	tv.runQualityChecks(result, validationResult)

	if tv.verbose {
		if validationResult.Valid {
			fmt.Printf("✅ All validations passed in %v\n", time.Since(startTime))
		} else {
			fmt.Printf("❌ Validation completed with issues in %v\n", time.Since(startTime))
		}
	}

	return validationResult, nil
}

// validateSyntax checks if generated test files have valid Go syntax
func (tv *TestValidator) validateSyntax(result *models.GenerationResult, validationResult *ValidationResult) bool {
	allValid := true

	for _, generatedFile := range result.GeneratedFiles {
		fullPath := filepath.Join(tv.projectPath, generatedFile.Path)

		if tv.verbose {
			fmt.Printf("🔍 Checking syntax: %s\n", generatedFile.Path)
		}

		// Read the generated file
		src, err := os.ReadFile(fullPath)
		if err != nil {
			validationResult.SyntaxErrors = append(validationResult.SyntaxErrors,
				fmt.Sprintf("Failed to read %s: %v", generatedFile.Path, err))
			allValid = false
			continue
		}

		// Parse the file for syntax errors
		_, err = parser.ParseFile(tv.fileSet, fullPath, src, parser.ParseComments)
		if err != nil {
			validationResult.SyntaxErrors = append(validationResult.SyntaxErrors,
				fmt.Sprintf("Syntax error in %s: %v", generatedFile.Path, err))
			allValid = false
		}
	}

	return allValid
}

// validateCompilation checks if generated tests compile successfully
func (tv *TestValidator) validateCompilation(result *models.GenerationResult, validationResult *ValidationResult) bool {
	// Use go build to check compilation
	cmd := exec.Command("go", "build", "-o", "/dev/null", "./...")
	cmd.Dir = tv.projectPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		validationResult.CompileErrors = append(validationResult.CompileErrors,
			fmt.Sprintf("Compilation failed: %v\nOutput: %s", err, string(output)))
		return false
	}

	// Also specifically check test compilation
	cmd = exec.Command("go", "test", "-c", "-o", "/dev/null", "./...")
	cmd.Dir = tv.projectPath

	output, err = cmd.CombinedOutput()
	if err != nil {
		validationResult.CompileErrors = append(validationResult.CompileErrors,
			fmt.Sprintf("Test compilation failed: %v\nOutput: %s", err, string(output)))
		return false
	}

	return true
}

// validateExecution runs the generated tests to ensure they execute properly
func (tv *TestValidator) validateExecution(result *models.GenerationResult, validationResult *ValidationResult) bool {
	// Run tests with verbose output to get detailed results
	cmd := exec.Command("go", "test", "-v", "./...")
	cmd.Dir = tv.projectPath

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// Parse test results
	tv.parseTestResults(outputStr, validationResult)

	if err != nil {
		validationResult.RuntimeErrors = append(validationResult.RuntimeErrors,
			fmt.Sprintf("Test execution failed: %v\nOutput: %s", err, outputStr))
		return false
	}

	// Check if any tests failed
	if validationResult.TestsFailed > 0 {
		validationResult.RuntimeErrors = append(validationResult.RuntimeErrors,
			fmt.Sprintf("%d tests failed during execution", validationResult.TestsFailed))
		return false
	}

	return true
}

// parseTestResults parses go test output to extract test statistics
func (tv *TestValidator) parseTestResults(output string, validationResult *ValidationResult) {
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		// Count individual test results
		if strings.Contains(line, "--- PASS:") {
			validationResult.TestsPassed++
		} else if strings.Contains(line, "--- FAIL:") {
			validationResult.TestsFailed++
		}

		// Parse final summary line
		if strings.HasPrefix(line, "PASS") || strings.HasPrefix(line, "FAIL") {
			if strings.Contains(line, "coverage:") {
				// Try to extract coverage information
				tv.extractCoverageInfo(line, validationResult)
			}
		}
	}

	validationResult.TestsRun = validationResult.TestsPassed + validationResult.TestsFailed
}

// extractCoverageInfo extracts coverage information from test output
func (tv *TestValidator) extractCoverageInfo(line string, validationResult *ValidationResult) {
	// Look for coverage percentage in output like "coverage: 75.0% of statements"
	parts := strings.Split(line, "coverage:")
	if len(parts) > 1 {
		coveragePart := strings.TrimSpace(parts[1])
		if strings.Contains(coveragePart, "%") {
			coverageStr := strings.Split(coveragePart, "%")[0]
			if coverage, err := parseFloat(coverageStr); err == nil {
				validationResult.CoverageImproved = coverage
			}
		}
	}
}

// runQualityChecks performs additional quality checks on generated tests
func (tv *TestValidator) runQualityChecks(result *models.GenerationResult, validationResult *ValidationResult) {
	for _, generatedFile := range result.GeneratedFiles {
		tv.checkTestQuality(generatedFile, validationResult)
	}
}

// checkTestQuality checks the quality of individual test files
func (tv *TestValidator) checkTestQuality(generatedFile *models.GeneratedFile, validationResult *ValidationResult) {
	fullPath := filepath.Join(tv.projectPath, generatedFile.Path)

	src, err := os.ReadFile(fullPath)
	if err != nil {
		return
	}

	file, err := parser.ParseFile(tv.fileSet, fullPath, src, parser.ParseComments)
	if err != nil {
		return
	}

	// Check for common quality issues
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			if strings.HasPrefix(node.Name.Name, "Test") {
				tv.checkTestFunction(node, generatedFile.Path, validationResult)
			}
		}
		return true
	})
}

// checkTestFunction checks the quality of a specific test function
func (tv *TestValidator) checkTestFunction(funcDecl *ast.FuncDecl, filePath string, validationResult *ValidationResult) {
	// Check if test function has proper structure
	if funcDecl.Body == nil || len(funcDecl.Body.List) == 0 {
		validationResult.Warnings = append(validationResult.Warnings,
			fmt.Sprintf("Empty test function %s in %s", funcDecl.Name.Name, filePath))
		return
	}

	// Check for proper test setup (t.Run, assertions, etc.)
	hasAssertions := false
	hasTestCases := false

	ast.Inspect(funcDecl, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.CallExpr:
			if ident, ok := node.Fun.(*ast.Ident); ok {
				// Look for testing patterns
				if strings.Contains(ident.Name, "Error") ||
					strings.Contains(ident.Name, "Fail") ||
					strings.Contains(ident.Name, "Fatal") {
					hasAssertions = true
				}
			}
			if selExpr, ok := node.Fun.(*ast.SelectorExpr); ok {
				if selExpr.Sel.Name == "Run" {
					hasTestCases = true
				}
			}
		}
		return true
	})

	// Report quality issues
	if !hasAssertions {
		validationResult.Warnings = append(validationResult.Warnings,
			fmt.Sprintf("Test function %s in %s lacks proper assertions", funcDecl.Name.Name, filePath))
	}

	if !hasTestCases && strings.Contains(funcDecl.Name.Name, "Table") {
		validationResult.Warnings = append(validationResult.Warnings,
			fmt.Sprintf("Table test function %s in %s lacks t.Run calls", funcDecl.Name.Name, filePath))
	}
}

// ValidateIndividualTest validates a single test file
func (tv *TestValidator) ValidateIndividualTest(testFilePath string) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:         true,
		SyntaxErrors:  make([]string, 0),
		CompileErrors: make([]string, 0),
		RuntimeErrors: make([]string, 0),
		Warnings:      make([]string, 0),
	}

	// Create a minimal generation result for validation
	generatedFile := &models.GeneratedFile{
		Path: testFilePath,
	}

	genResult := &models.GenerationResult{
		GeneratedFiles: []*models.GeneratedFile{generatedFile},
	}

	// Run syntax validation
	if !tv.validateSyntax(genResult, result) {
		result.Valid = false
	}

	// Run quality checks
	tv.runQualityChecks(genResult, result)

	return result, nil
}

// parseFloat parses a string to float64, returns 0 on error
func parseFloat(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty string")
	}

	// Simple float parsing - could use strconv.ParseFloat for production
	var result float64
	_, err := fmt.Sscanf(s, "%f", &result)
	return result, err
}

// GetValidationSummary provides a human-readable summary of validation results
func (tv *TestValidator) GetValidationSummary(result *ValidationResult) string {
	var summary strings.Builder

	summary.WriteString("🔍 Test Validation Summary\n")
	summary.WriteString("========================\n")

	if result.Valid {
		summary.WriteString("✅ Overall Status: PASSED\n")
	} else {
		summary.WriteString("❌ Overall Status: FAILED\n")
	}

	summary.WriteString(fmt.Sprintf("⏱️  Compilation Time: %v\n", result.CompilationTime))
	summary.WriteString(fmt.Sprintf("⏱️  Execution Time: %v\n", result.ExecutionTime))

	if result.TestsRun > 0 {
		summary.WriteString(fmt.Sprintf("🧪 Tests Run: %d\n", result.TestsRun))
		summary.WriteString(fmt.Sprintf("✅ Tests Passed: %d\n", result.TestsPassed))

		if result.TestsFailed > 0 {
			summary.WriteString(fmt.Sprintf("❌ Tests Failed: %d\n", result.TestsFailed))
		}

		if result.CoverageImproved > 0 {
			summary.WriteString(fmt.Sprintf("📈 Coverage: %.1f%%\n", result.CoverageImproved))
		}
	}

	if len(result.SyntaxErrors) > 0 {
		summary.WriteString(fmt.Sprintf("❌ Syntax Errors: %d\n", len(result.SyntaxErrors)))
		for _, err := range result.SyntaxErrors {
			summary.WriteString(fmt.Sprintf("   • %s\n", err))
		}
	}

	if len(result.CompileErrors) > 0 {
		summary.WriteString(fmt.Sprintf("❌ Compile Errors: %d\n", len(result.CompileErrors)))
		for _, err := range result.CompileErrors {
			summary.WriteString(fmt.Sprintf("   • %s\n", err))
		}
	}

	if len(result.RuntimeErrors) > 0 {
		summary.WriteString(fmt.Sprintf("❌ Runtime Errors: %d\n", len(result.RuntimeErrors)))
		for _, err := range result.RuntimeErrors {
			summary.WriteString(fmt.Sprintf("   • %s\n", err))
		}
	}

	if len(result.Warnings) > 0 {
		summary.WriteString(fmt.Sprintf("⚠️  Warnings: %d\n", len(result.Warnings)))
		for _, warning := range result.Warnings {
			summary.WriteString(fmt.Sprintf("   • %s\n", warning))
		}
	}

	return summary.String()
}
