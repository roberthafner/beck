package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/beck/go-coverage-analyzer/internal/analyzer"
	"github.com/beck/go-coverage-analyzer/internal/config"
	"github.com/beck/go-coverage-analyzer/internal/generator"
	"github.com/beck/go-coverage-analyzer/internal/reporter"
	"github.com/beck/go-coverage-analyzer/pkg/models"
	"github.com/spf13/cobra"
)

var (
	version = "0.1.0"
	cfg     *config.Config
)

func main() {
	var err error
	cfg, err = config.Load()
	if err != nil {
		// Use default configuration if loading fails
		cfg = &config.Config{}
		if err := setDefaultConfig(cfg); err != nil {
			log.Fatalf("Failed to initialize default configuration: %v", err)
		}
		// Only show warning in verbose mode
		if os.Getenv("GCOV_VERBOSE") == "true" {
			fmt.Fprintf(os.Stderr, "Warning: Using default configuration (config load failed: %v)\n", err)
		}
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// setDefaultConfig initializes a configuration with default values
func setDefaultConfig(cfg *config.Config) error {
	cfg.ExcludeDirs = []string{"vendor", "testdata", ".git", "node_modules"}
	cfg.IncludeTests = false
	cfg.CoverageThreshold = 80.0
	cfg.CalculateComplexity = true
	cfg.MinComplexity = 1
	cfg.OutputFormat = "console"
	cfg.OutputDir = "."
	cfg.Verbose = false
	cfg.ProfileOutput = "coverage.out"
	cfg.TemplateStyle = "standard"
	cfg.GenerateMocks = true
	cfg.TableDriven = true
	cfg.GenerateBenchmarks = false
	cfg.OverwriteTests = false
	cfg.MaxTestCases = 10
	cfg.MaxConcurrency = 4
	cfg.EnableCaching = true
	cfg.IgnoreFunctions = []string{}
	cfg.CustomPatterns = []string{}
	cfg.GoVersions = []string{}
	cfg.BuildTags = []string{}
	return nil
}

var rootCmd = &cobra.Command{
	Use:   "gcov",
	Short: "Go Test Coverage Analyzer & Generator",
	Long: `A comprehensive tool that analyzes Go project test coverage,
identifies uncovered code paths, and automatically generates intelligent
unit tests to improve coverage.

Features:
- Detailed coverage analysis with gap identification
- Intelligent test generation with mocks and table-driven tests
- Multi-format reporting (console, HTML, JSON)
- CI/CD integration and IDE support
- Template customization and extensibility`,
	Version: version,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Load configuration from flags
		if configPath, _ := cmd.Flags().GetString("config"); configPath != "" {
			if newCfg, err := config.LoadFromFile(configPath); err == nil {
				cfg = newCfg
			}
		}
	},
}

var analyzeCmd = &cobra.Command{
	Use:   "analyze [project-path]",
	Short: "Analyze test coverage for a Go project",
	Long: `Analyze the specified Go project (or current directory) to identify
uncovered code paths and generate a comprehensive coverage report.

The analyzer integrates with Go's built-in coverage tools to provide detailed
insights into test coverage gaps at function, package, and project levels.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runAnalysis,
}

var generateCmd = &cobra.Command{
	Use:   "generate [project-path]",
	Short: "Generate unit tests for uncovered code",
	Long: `Generate comprehensive unit tests for uncovered functions and methods
in the specified Go project. The generator creates intelligent tests with
appropriate mocks, table-driven test patterns, and edge case coverage.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runGeneration,
}

var validateCmd = &cobra.Command{
	Use:   "validate [project-path]",
	Short: "Validate existing or generated tests",
	Long: `Validate test files for syntax, compilation, and quality issues.
This command checks test files for common problems, ensures they compile
and run correctly, and provides recommendations for improvement.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runValidation,
}

var reportCmd = &cobra.Command{
	Use:   "report [project-path]",
	Short: "Generate coverage reports without analysis",
	Long: `Generate coverage reports from existing coverage data without
running a new analysis. Useful for creating different report formats
from previously collected coverage information.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runReporting,
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringP("config", "c", "", "Configuration file path")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringP("output", "o", "console", "Output format (console, json, html, xml)")
	rootCmd.PersistentFlags().StringSliceP("exclude", "e", []string{"vendor", "testdata", ".git"}, "Directories to exclude")
	rootCmd.PersistentFlags().Float64P("threshold", "t", 80.0, "Coverage threshold percentage")

	// Analyze command flags
	analyzeCmd.Flags().BoolP("include-tests", "i", false, "Include test files in analysis")
	analyzeCmd.Flags().StringP("package", "p", "", "Specific package pattern to analyze")
	analyzeCmd.Flags().BoolP("profile", "", false, "Generate coverage profile")
	analyzeCmd.Flags().StringP("profile-output", "", "coverage.out", "Coverage profile output file")
	analyzeCmd.Flags().BoolP("complexity", "", true, "Calculate cyclomatic complexity")
	analyzeCmd.Flags().IntP("min-complexity", "", 1, "Minimum complexity threshold for reporting")

	// Generate command flags
	generateCmd.Flags().BoolP("dry-run", "d", false, "Preview generated tests without writing files")
	generateCmd.Flags().StringP("template-style", "", "standard", "Test template style (standard, testify, table)")
	generateCmd.Flags().BoolP("generate-mocks", "m", true, "Generate mocks for interfaces")
	generateCmd.Flags().BoolP("table-driven", "", true, "Generate table-driven tests when applicable")
	generateCmd.Flags().BoolP("benchmarks", "b", false, "Generate benchmark tests")
	generateCmd.Flags().BoolP("overwrite", "w", false, "Overwrite existing test files")
	generateCmd.Flags().StringSliceP("ignore-functions", "", []string{}, "Function patterns to ignore")
	generateCmd.Flags().IntP("max-cases", "", 10, "Maximum test cases per function")

	// Validate command flags
	validateCmd.Flags().StringP("test-file", "", "", "Specific test file to validate")
	validateCmd.Flags().BoolP("compile-check", "", true, "Check if tests compile")
	validateCmd.Flags().BoolP("run-tests", "", true, "Run tests to check execution")
	validateCmd.Flags().BoolP("quality-check", "", true, "Run quality checks on test structure")

	// Report command flags
	reportCmd.Flags().StringP("input", "", "coverage.out", "Input coverage profile file")
	reportCmd.Flags().StringP("output-file", "", "", "Output file path (default: stdout)")
	reportCmd.Flags().BoolP("open", "", false, "Open HTML report in browser")

	// Add subcommands
	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(reportCmd)
}

func runAnalysis(cmd *cobra.Command, args []string) error {
	projectPath := "."
	if len(args) > 0 {
		projectPath = args[0]
	}

	// Get command-line flags
	verbose, _ := cmd.Flags().GetBool("verbose")
	outputFormat, _ := cmd.Flags().GetString("output")
	excludeDirs, _ := cmd.Flags().GetStringSlice("exclude")
	threshold, _ := cmd.Flags().GetFloat64("threshold")
	includeTests, _ := cmd.Flags().GetBool("include-tests")
	packagePattern, _ := cmd.Flags().GetString("package")
	generateProfile, _ := cmd.Flags().GetBool("profile")
	profileOutput, _ := cmd.Flags().GetString("profile-output")
	calculateComplexity, _ := cmd.Flags().GetBool("complexity")
	minComplexity, _ := cmd.Flags().GetInt("min-complexity")

	// Configure analysis options
	opts := &analyzer.Options{
		ProjectPath:         projectPath,
		ExcludeDirs:         excludeDirs,
		IncludeTests:        includeTests,
		PackagePattern:      packagePattern,
		GenerateProfile:     generateProfile,
		ProfileOutput:       profileOutput,
		CalculateComplexity: calculateComplexity,
		MinComplexity:       minComplexity,
		Verbose:             verbose,
	}

	if verbose {
		fmt.Printf("🔍 Analyzing Go project at: %s\n", projectPath)
	}

	// Run coverage analysis
	result, err := analyzer.Analyze(opts)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	// Generate report
	reportOpts := &reporter.Options{
		Format:      outputFormat,
		Threshold:   threshold,
		Verbose:     verbose,
		ShowDetails: verbose,
	}

	if err := reporter.Generate(result, reportOpts); err != nil {
		return fmt.Errorf("report generation failed: %w", err)
	}

	// Exit with error code if coverage is below threshold
	if result.OverallCoverage < threshold {
		if verbose {
			fmt.Printf("\n❌ Coverage %.1f%% is below threshold %.1f%%\n", result.OverallCoverage, threshold)
		}
		os.Exit(1)
	}

	if verbose {
		fmt.Printf("\n✅ Coverage %.1f%% meets threshold %.1f%%\n", result.OverallCoverage, threshold)
	}

	return nil
}

func runGeneration(cmd *cobra.Command, args []string) error {
	projectPath := "."
	if len(args) > 0 {
		projectPath = args[0]
	}

	// Get command-line flags
	verbose, _ := cmd.Flags().GetBool("verbose")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	templateStyle, _ := cmd.Flags().GetString("template-style")
	generateMocks, _ := cmd.Flags().GetBool("generate-mocks")
	tableDriven, _ := cmd.Flags().GetBool("table-driven")
	benchmarks, _ := cmd.Flags().GetBool("benchmarks")
	overwrite, _ := cmd.Flags().GetBool("overwrite")
	ignoreFunctions, _ := cmd.Flags().GetStringSlice("ignore-functions")
	maxCases, _ := cmd.Flags().GetInt("max-cases")
	excludeDirs, _ := cmd.Flags().GetStringSlice("exclude")

	if verbose {
		fmt.Printf("🛠️  Generating tests for project: %s\n", projectPath)
		if dryRun {
			fmt.Println("👀 Running in dry-run mode (no files will be written)")
		}
	}

	// First run analysis to identify uncovered code
	analyzeOpts := &analyzer.Options{
		ProjectPath:         projectPath,
		ExcludeDirs:         excludeDirs,
		IncludeTests:        false, // Don't include tests in generation analysis
		CalculateComplexity: true,
		Verbose:             verbose,
	}

	result, err := analyzer.Analyze(analyzeOpts)
	if err != nil {
		return fmt.Errorf("analysis for generation failed: %w", err)
	}

	// Configure generation options
	genOpts := &generator.Options{
		ProjectPath:        projectPath,
		DryRun:             dryRun,
		TemplateStyle:      templateStyle,
		GenerateMocks:      generateMocks,
		TableDriven:        tableDriven,
		GenerateBenchmarks: benchmarks,
		Overwrite:          overwrite,
		IgnoreFunctions:    ignoreFunctions,
		MaxTestCases:       maxCases,
		Verbose:            verbose,
	}

	// Generate tests
	genResult, err := generator.Generate(result, genOpts)
	if err != nil {
		return fmt.Errorf("test generation failed: %w", err)
	}

	if verbose {
		fmt.Printf("\n📊 Generation Summary:\n")
		fmt.Printf("   Tests Generated: %d\n", genResult.TestsGenerated)
		fmt.Printf("   Files Created:   %d\n", genResult.FilesCreated)
		fmt.Printf("   Functions Covered: %d\n", genResult.FunctionsCovered)

		if !dryRun {
			fmt.Printf("\n✅ Test generation completed successfully\n")
		} else {
			fmt.Printf("\n👀 Dry-run completed - no files were written\n")
		}
	}

	return nil
}

func runValidation(cmd *cobra.Command, args []string) error {
	projectPath := "."
	if len(args) > 0 {
		projectPath = args[0]
	}

	// Get command-line flags
	verbose, _ := cmd.Flags().GetBool("verbose")
	testFile, _ := cmd.Flags().GetString("test-file")
	_, _ = cmd.Flags().GetBool("compile-check")
	_, _ = cmd.Flags().GetBool("run-tests")
	_, _ = cmd.Flags().GetBool("quality-check")

	if verbose {
		fmt.Printf("🔍 Validating tests in project: %s\n", projectPath)
		if testFile != "" {
			fmt.Printf("🔍 Focusing on test file: %s\n", testFile)
		}
	}

	// Create validator
	validator := generator.NewTestValidator(projectPath, verbose)

	if testFile != "" {
		// Validate specific test file
		result, err := validator.ValidateIndividualTest(testFile)
		if err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}

		if verbose {
			fmt.Println(validator.GetValidationSummary(result))
		}

		if !result.Valid {
			return fmt.Errorf("test validation failed")
		}

		if verbose {
			fmt.Println("✅ Test validation passed")
		}
	} else {
		// Create a dummy generation result to validate all test files
		result := &models.GenerationResult{
			GeneratedFiles: make([]*models.GeneratedFile, 0),
		}

		// Find all test files in the project
		err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if strings.HasSuffix(info.Name(), "_test.go") {
				relativePath, err := filepath.Rel(projectPath, path)
				if err != nil {
					relativePath = path
				}

				result.GeneratedFiles = append(result.GeneratedFiles, &models.GeneratedFile{
					Path: relativePath,
				})
			}
			return nil
		})

		if err != nil {
			return fmt.Errorf("failed to find test files: %w", err)
		}

		if len(result.GeneratedFiles) == 0 {
			if verbose {
				fmt.Println("⚠️ No test files found in project")
			}
			return nil
		}

		// Run validation
		validationResult, err := validator.ValidateTests(result)
		if err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}

		if verbose {
			fmt.Println(validator.GetValidationSummary(validationResult))
		}

		if !validationResult.Valid {
			return fmt.Errorf("test validation failed")
		}

		if verbose {
			fmt.Printf("✅ All %d test files validated successfully\n", len(result.GeneratedFiles))
		}
	}

	return nil
}

func runReporting(cmd *cobra.Command, args []string) error {
	projectPath := "."
	if len(args) > 0 {
		projectPath = args[0]
	}

	// Get command-line flags
	verbose, _ := cmd.Flags().GetBool("verbose")
	outputFormat, _ := cmd.Flags().GetString("output")
	inputFile, _ := cmd.Flags().GetString("input")
	outputFile, _ := cmd.Flags().GetString("output-file")
	openReport, _ := cmd.Flags().GetBool("open")
	threshold, _ := cmd.Flags().GetFloat64("threshold")

	if verbose {
		fmt.Printf("📋 Generating report from: %s\n", inputFile)
	}

	// Configure reporting options
	reportOpts := &reporter.Options{
		Format:      outputFormat,
		InputFile:   inputFile,
		OutputFile:  outputFile,
		OpenReport:  openReport,
		Threshold:   threshold,
		Verbose:     verbose,
		ShowDetails: verbose,
	}

	// Generate report from existing coverage data
	if err := reporter.GenerateFromProfile(projectPath, reportOpts); err != nil {
		return fmt.Errorf("report generation failed: %w", err)
	}

	if verbose {
		fmt.Printf("✅ Report generated successfully\n")
	}

	return nil
}
