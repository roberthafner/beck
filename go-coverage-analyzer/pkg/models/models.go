package models

import (
	"time"
)

// AnalysisResult represents the complete result of coverage analysis
type AnalysisResult struct {
	ProjectPath        string              `json:"project_path"`
	Timestamp          time.Time           `json:"timestamp"`
	OverallCoverage    float64             `json:"overall_coverage"`
	FunctionCoverage   float64             `json:"function_coverage"`
	BranchCoverage     float64             `json:"branch_coverage"`
	LineCoverage       float64             `json:"line_coverage"`
	PackageCoverage    map[string]*Package `json:"packages"`
	UncoveredFunctions []*Function         `json:"uncovered_functions"`
	Summary            *Summary            `json:"summary"`
	Metadata           *Metadata           `json:"metadata"`
}

// Package represents coverage information for a Go package
type Package struct {
	Name             string           `json:"name"`
	Path             string           `json:"path"`
	Coverage         float64          `json:"coverage"`
	FunctionCoverage float64          `json:"function_coverage"`
	BranchCoverage   float64          `json:"branch_coverage"`
	LineCoverage     float64          `json:"line_coverage"`
	Files            map[string]*File `json:"files"`
	TotalLines       int              `json:"total_lines"`
	CoveredLines     int              `json:"covered_lines"`
	UncoveredLines   int              `json:"uncovered_lines"`
	TotalFunctions   int              `json:"total_functions"`
	CoveredFunctions int              `json:"covered_functions"`
	Complexity       int              `json:"complexity"`
}

// File represents coverage information for a Go source file
type File struct {
	Name             string      `json:"name"`
	Path             string      `json:"path"`
	Package          string      `json:"package"`
	Coverage         float64     `json:"coverage"`
	FunctionCoverage float64     `json:"function_coverage"`
	BranchCoverage   float64     `json:"branch_coverage"`
	LineCoverage     float64     `json:"line_coverage"`
	Functions        []*Function `json:"functions"`
	TotalLines       int         `json:"total_lines"`
	CoveredLines     int         `json:"covered_lines"`
	UncoveredLines   int         `json:"uncovered_lines"`
	CoverageBlocks   []*Block    `json:"coverage_blocks"`
	Complexity       int         `json:"complexity"`
	HasTests         bool        `json:"has_tests"`
	TestFiles        []string    `json:"test_files,omitempty"`
}

// Function represents a function or method that can be tested
type Function struct {
	Name           string   `json:"name"`
	Signature      string   `json:"signature"`
	File           string   `json:"file"`
	Package        string   `json:"package"`
	StartLine      int      `json:"start_line"`
	EndLine        int      `json:"end_line"`
	Coverage       float64  `json:"coverage"`
	IsCovered      bool     `json:"is_covered"`
	IsTestable     bool     `json:"is_testable"`
	IsMethod       bool     `json:"is_method"`
	IsExported     bool     `json:"is_exported"`
	ReceiverType   string   `json:"receiver_type,omitempty"`
	Parameters     []*Param `json:"parameters"`
	ReturnTypes    []string `json:"return_types"`
	Complexity     int      `json:"complexity"`
	HasTests       bool     `json:"has_tests"`
	TestFiles      []string `json:"test_files,omitempty"`
	Dependencies   []string `json:"dependencies,omitempty"`
	CallsExternal  bool     `json:"calls_external"`
	HasErrorReturn bool     `json:"has_error_return"`
	CanPanic       bool     `json:"can_panic"`
}

// Param represents a function parameter
type Param struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// Block represents a coverage block (statement or branch)
type Block struct {
	StartLine int   `json:"start_line"`
	StartCol  int   `json:"start_col"`
	EndLine   int   `json:"end_line"`
	EndCol    int   `json:"end_col"`
	NumStmts  int   `json:"num_statements"`
	Count     int64 `json:"count"`
	IsCovered bool  `json:"is_covered"`
}

// Summary provides high-level coverage statistics
type Summary struct {
	TotalPackages     int     `json:"total_packages"`
	TotalFiles        int     `json:"total_files"`
	TotalFunctions    int     `json:"total_functions"`
	TestedFunctions   int     `json:"tested_functions"`
	UntestedFunctions int     `json:"untested_functions"`
	TotalLines        int     `json:"total_lines"`
	CoveredLines      int     `json:"covered_lines"`
	UncoveredLines    int     `json:"uncovered_lines"`
	OverallCoverage   float64 `json:"overall_coverage"`
	FunctionCoverage  float64 `json:"function_coverage"`
	BranchCoverage    float64 `json:"branch_coverage"`
	LineCoverage      float64 `json:"line_coverage"`

	// Coverage by category
	PublicFunctionCoverage  float64 `json:"public_function_coverage"`
	PrivateFunctionCoverage float64 `json:"private_function_coverage"`
	MethodCoverage          float64 `json:"method_coverage"`

	// Quality metrics
	AvgComplexity           float64 `json:"avg_complexity"`
	HighComplexityFunctions int     `json:"high_complexity_functions"`
	MaxComplexity           int     `json:"max_complexity"`
	TotalComplexity         int     `json:"total_complexity"`

	// Test statistics
	TotalTestFiles int     `json:"total_test_files"`
	TestCoverage   float64 `json:"test_coverage"`
}

// Metadata contains information about the analysis execution
type Metadata struct {
	Version          string        `json:"version"`
	AnalysisTime     time.Duration `json:"analysis_time"`
	GoVersion        string        `json:"go_version"`
	ModulePath       string        `json:"module_path"`
	BuildConstraints []string      `json:"build_constraints,omitempty"`
	ExcludedDirs     []string      `json:"excluded_dirs"`
	IncludedPackages []string      `json:"included_packages"`
	Configuration    interface{}   `json:"configuration,omitempty"`
	ProfilePath      string        `json:"profile_path,omitempty"`
}

// GenerationResult represents the result of test generation
type GenerationResult struct {
	ProjectPath       string           `json:"project_path"`
	Timestamp         time.Time        `json:"timestamp"`
	TestsGenerated    int              `json:"tests_generated"`
	FilesCreated      int              `json:"files_created"`
	FilesModified     int              `json:"files_modified"`
	FunctionsCovered  int              `json:"functions_covered"`
	GeneratedFiles    []*GeneratedFile `json:"generated_files"`
	EstimatedCoverage float64          `json:"estimated_coverage"`
	GenerationTime    time.Duration    `json:"generation_time"`
	Errors            []string         `json:"errors,omitempty"`
	Warnings          []string         `json:"warnings,omitempty"`
}

// GeneratedFile represents a test file that was generated
type GeneratedFile struct {
	Path           string      `json:"path"`
	Package        string      `json:"package"`
	TestsGenerated int         `json:"tests_generated"`
	TestCases      []*TestCase `json:"test_cases"`
	Size           int64       `json:"size"`
	Created        bool        `json:"created"`
	Modified       bool        `json:"modified"`
}

// TestCase represents a generated test case
type TestCase struct {
	FunctionName  string `json:"function_name"`
	TestName      string `json:"test_name"`
	TestType      string `json:"test_type"` // unit, table, benchmark, integration
	InputCount    int    `json:"input_count"`
	HasMocks      bool   `json:"has_mocks"`
	HasSetup      bool   `json:"has_setup"`
	HasTeardown   bool   `json:"has_teardown"`
	ExpectedLines int    `json:"expected_lines"`
	Complexity    int    `json:"complexity"`
}

// ProjectInfo represents information about a Go project
type ProjectInfo struct {
	ModulePath     string   `json:"module_path"`
	GoVersion      string   `json:"go_version"`
	RootDir        string   `json:"root_dir"`
	Packages       []string `json:"packages"`
	TotalFiles     int      `json:"total_files"`
	TotalLines     int      `json:"total_lines"`
	TotalFunctions int      `json:"total_functions"`
	HasTests       bool     `json:"has_tests"`
	TestFiles      []string `json:"test_files"`
	BuildTags      []string `json:"build_tags,omitempty"`
	Dependencies   []string `json:"dependencies,omitempty"`
}

// CoverageProfile represents parsed coverage profile data
type CoverageProfile struct {
	Mode   string                  `json:"mode"`
	Blocks []*ProfileBlock         `json:"blocks"`
	Files  map[string]*FileProfile `json:"files"`
}

// ProfileBlock represents a single coverage block from a profile
type ProfileBlock struct {
	FileName  string `json:"file_name"`
	StartLine int    `json:"start_line"`
	StartCol  int    `json:"start_col"`
	EndLine   int    `json:"end_line"`
	EndCol    int    `json:"end_col"`
	NumStmts  int    `json:"num_statements"`
	Count     int64  `json:"count"`
}

// FileProfile represents coverage information for a file from a profile
type FileProfile struct {
	FileName     string          `json:"file_name"`
	Blocks       []*ProfileBlock `json:"blocks"`
	Coverage     float64         `json:"coverage"`
	TotalStmts   int             `json:"total_statements"`
	CoveredStmts int             `json:"covered_statements"`
}

// Template represents a test template
type Template struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Path        string `json:"path"`
	Content     string `json:"content"`
	Description string `json:"description"`
}

// MockInfo represents information about a mock that needs to be generated
type MockInfo struct {
	InterfaceName string   `json:"interface_name"`
	PackageName   string   `json:"package_name"`
	Methods       []string `json:"methods"`
	FilePath      string   `json:"file_path"`
}

// Helper methods for AnalysisResult

// GetUncoveredFunctions returns all functions that lack test coverage
func (ar *AnalysisResult) GetUncoveredFunctions() []*Function {
	var uncovered []*Function

	for _, pkg := range ar.PackageCoverage {
		for _, file := range pkg.Files {
			for _, function := range file.Functions {
				if !function.IsCovered && function.IsTestable {
					uncovered = append(uncovered, function)
				}
			}
		}
	}

	return uncovered
}

// GetPackageByName returns a package by name
func (ar *AnalysisResult) GetPackageByName(name string) *Package {
	return ar.PackageCoverage[name]
}

// GetLowCoveragePackages returns packages below the specified threshold
func (ar *AnalysisResult) GetLowCoveragePackages(threshold float64) []*Package {
	var lowCoverage []*Package

	for _, pkg := range ar.PackageCoverage {
		if pkg.Coverage < threshold {
			lowCoverage = append(lowCoverage, pkg)
		}
	}

	return lowCoverage
}

// GetHighComplexityFunctions returns functions with complexity above threshold
func (ar *AnalysisResult) GetHighComplexityFunctions(threshold int) []*Function {
	var highComplexity []*Function

	for _, pkg := range ar.PackageCoverage {
		for _, file := range pkg.Files {
			for _, function := range file.Functions {
				if function.Complexity > threshold {
					highComplexity = append(highComplexity, function)
				}
			}
		}
	}

	return highComplexity
}

// GetTestableFunction returns all testable functions
func (ar *AnalysisResult) GetTestableFunctions() []*Function {
	var testable []*Function

	for _, pkg := range ar.PackageCoverage {
		for _, file := range pkg.Files {
			for _, function := range file.Functions {
				if function.IsTestable {
					testable = append(testable, function)
				}
			}
		}
	}

	return testable
}

// GetFunctionsWithoutTests returns functions that don't have corresponding test functions
func (ar *AnalysisResult) GetFunctionsWithoutTests() []*Function {
	var withoutTests []*Function

	for _, pkg := range ar.PackageCoverage {
		for _, file := range pkg.Files {
			for _, function := range file.Functions {
				if function.IsTestable && !function.HasTests {
					withoutTests = append(withoutTests, function)
				}
			}
		}
	}

	return withoutTests
}

// GetExternalDependencies returns functions that call external dependencies
func (ar *AnalysisResult) GetExternalDependencies() []*Function {
	var external []*Function

	for _, pkg := range ar.PackageCoverage {
		for _, file := range pkg.Files {
			for _, function := range file.Functions {
				if function.CallsExternal {
					external = append(external, function)
				}
			}
		}
	}

	return external
}

// CoverageTrend represents coverage trends over time
type CoverageTrend struct {
	Timestamp        time.Time `json:"timestamp"`
	CurrentCoverage  float64   `json:"current_coverage"`
	PreviousCoverage float64   `json:"previous_coverage"`
	Change           float64   `json:"change"`
	Direction        string    `json:"direction"` // up, down, stable
}

// PackageMetrics represents detailed metrics for a package
type PackageMetrics struct {
	PackageName       string  `json:"package_name"`
	TotalFunctions    int     `json:"total_functions"`
	TestedFunctions   int     `json:"tested_functions"`
	UntestedFunctions int     `json:"untested_functions"`
	FunctionCoverage  float64 `json:"function_coverage"`
	LineCoverage      float64 `json:"line_coverage"`
	ComplexityScore   int     `json:"complexity_score"`
	TestableExported  int     `json:"testable_exported"`
	UntestedExported  int     `json:"untested_exported"`
	TestablePrivate   int     `json:"testable_private"`
	UntestedPrivate   int     `json:"untested_private"`
}
