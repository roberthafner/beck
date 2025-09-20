package analyzer

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/beck/go-coverage-analyzer/pkg/coverage"
	"github.com/beck/go-coverage-analyzer/pkg/models"
)

// Options contains configuration for the analysis
type Options struct {
	ProjectPath         string
	ExcludeDirs         []string
	IncludeTests        bool
	PackagePattern      string
	GenerateProfile     bool
	ProfileOutput       string
	ProfilePath         string
	CalculateComplexity bool
	MinComplexity       int
	Verbose             bool
}

// Analyze performs coverage analysis on the specified Go project
func Analyze(opts *Options) (*models.AnalysisResult, error) {
	if opts.Verbose {
		fmt.Printf("🔍 Starting analysis of: %s\n", opts.ProjectPath)
	}

	// Create coverage analysis engine
	engine := coverage.NewAnalysisEngine(opts.Verbose)

	// Set default profile output if generating profile
	profileOutput := opts.ProfileOutput
	if opts.GenerateProfile && profileOutput == "" {
		profileOutput = filepath.Join(opts.ProjectPath, "coverage.out")
	}

	// Configure analysis options
	analysisOpts := &coverage.AnalysisOptions{
		ProjectPath:         opts.ProjectPath,
		ProfilePath:         opts.ProfilePath,
		ProfileOutput:       profileOutput,
		PackagePattern:      opts.PackagePattern,
		ExcludeDirs:         opts.ExcludeDirs,
		IncludeTests:        opts.IncludeTests,
		GenerateProfile:     opts.GenerateProfile,
		CalculateComplexity: opts.CalculateComplexity,
		MinComplexity:       opts.MinComplexity,
	}

	// Perform comprehensive analysis
	result, err := engine.AnalyzeProject(analysisOpts)
	if err != nil {
		return nil, fmt.Errorf("coverage analysis failed: %w", err)
	}

	// Apply minimum complexity filter if specified
	if opts.MinComplexity > 1 {
		result.UncoveredFunctions = filterByComplexity(result.UncoveredFunctions, opts.MinComplexity)
	}

	if opts.Verbose {
		fmt.Printf("✅ Analysis completed successfully\n")
		fmt.Printf("📊 Overall Coverage: %.1f%%\n", result.OverallCoverage)
		fmt.Printf("🔍 Uncovered Functions: %d\n", len(result.UncoveredFunctions))
	}

	return result, nil
}

// filterByComplexity filters functions by minimum complexity threshold
func filterByComplexity(functions []*models.Function, minComplexity int) []*models.Function {
	filtered := make([]*models.Function, 0)

	for _, function := range functions {
		if function.Complexity >= minComplexity {
			filtered = append(filtered, function)
		}
	}

	return filtered
}

// AnalyzeFromProfile performs analysis from an existing coverage profile
func AnalyzeFromProfile(profilePath, projectPath string, opts *Options) (*models.AnalysisResult, error) {
	if opts.Verbose {
		fmt.Printf("📋 Analyzing from existing profile: %s\n", profilePath)
	}

	// Update options to use existing profile
	analysisOpts := &coverage.AnalysisOptions{
		ProjectPath:         projectPath,
		ProfilePath:         profilePath,
		PackagePattern:      opts.PackagePattern,
		ExcludeDirs:         opts.ExcludeDirs,
		IncludeTests:        opts.IncludeTests,
		GenerateProfile:     false, // Don't generate, use existing
		CalculateComplexity: opts.CalculateComplexity,
		MinComplexity:       opts.MinComplexity,
	}

	// Create coverage analysis engine
	engine := coverage.NewAnalysisEngine(opts.Verbose)

	// Perform analysis
	result, err := engine.AnalyzeProject(analysisOpts)
	if err != nil {
		return nil, fmt.Errorf("profile analysis failed: %w", err)
	}

	return result, nil
}

// GetProjectStatistics returns basic project statistics without full analysis
func GetProjectStatistics(projectPath string, excludeDirs []string, verbose bool) (*models.ProjectInfo, error) {
	parser := coverage.NewProfileParser(verbose)
	return parser.GetProjectInfo(projectPath)
}

// ValidateCoverageProfile validates a coverage profile file
func ValidateCoverageProfile(profilePath string, verbose bool) error {
	parser := coverage.NewProfileParser(verbose)

	profile, err := parser.ParseProfile(profilePath)
	if err != nil {
		return fmt.Errorf("failed to parse profile: %w", err)
	}

	if err := parser.ValidateProfile(profile); err != nil {
		return fmt.Errorf("profile validation failed: %w", err)
	}

	if verbose {
		fmt.Printf("✅ Profile is valid: %d blocks across %d files\n",
			len(profile.Blocks), len(profile.Files))
	}

	return nil
}

// GenerateCoverageProfile generates a coverage profile for the project
func GenerateCoverageProfile(projectPath, outputPath, packagePattern string, verbose bool) error {
	parser := coverage.NewProfileParser(verbose)

	return parser.GenerateProfile(projectPath, outputPath, packagePattern)
}

// GetUncoveredFunctionsByComplexity returns uncovered functions sorted by complexity
func GetUncoveredFunctionsByComplexity(result *models.AnalysisResult, minComplexity int) []*models.Function {
	functions := make([]*models.Function, 0)

	for _, function := range result.UncoveredFunctions {
		if function.Complexity >= minComplexity {
			functions = append(functions, function)
		}
	}

	return functions
}

// GetCoverageGaps identifies specific coverage gaps in the codebase
func GetCoverageGaps(result *models.AnalysisResult, threshold float64) map[string][]*models.Function {
	gaps := make(map[string][]*models.Function)

	for _, pkg := range result.PackageCoverage {
		if pkg.Coverage < threshold {
			packageFunctions := make([]*models.Function, 0)

			for _, file := range pkg.Files {
				for _, function := range file.Functions {
					if function.IsTestable && !function.IsCovered {
						packageFunctions = append(packageFunctions, function)
					}
				}
			}

			if len(packageFunctions) > 0 {
				gaps[pkg.Name] = packageFunctions
			}
		}
	}

	return gaps
}

// GetHighComplexityUncoveredFunctions returns uncovered functions with high complexity
func GetHighComplexityUncoveredFunctions(result *models.AnalysisResult, complexityThreshold int) []*models.Function {
	functions := make([]*models.Function, 0)

	for _, function := range result.UncoveredFunctions {
		if function.Complexity > complexityThreshold {
			functions = append(functions, function)
		}
	}

	return functions
}

// CalculateTrendData calculates coverage trends (placeholder for future implementation)
func CalculateTrendData(current, previous *models.AnalysisResult) *models.CoverageTrend {
	// Placeholder for trend analysis
	trend := &models.CoverageTrend{
		Timestamp:        time.Now(),
		CurrentCoverage:  current.OverallCoverage,
		PreviousCoverage: 0.0,
		Change:           current.OverallCoverage,
		Direction:        "unknown",
	}

	if previous != nil {
		trend.PreviousCoverage = previous.OverallCoverage
		trend.Change = current.OverallCoverage - previous.OverallCoverage

		if trend.Change > 0 {
			trend.Direction = "up"
		} else if trend.Change < 0 {
			trend.Direction = "down"
		} else {
			trend.Direction = "stable"
		}
	}

	return trend
}

// AnalyzePackage performs analysis on a specific package
func AnalyzePackage(opts *Options, packagePath string) (*models.Package, error) {
	// Modify options to target specific package
	modifiedOpts := *opts
	modifiedOpts.PackagePattern = packagePath

	result, err := Analyze(&modifiedOpts)
	if err != nil {
		return nil, err
	}

	// Find the target package in results
	for _, pkg := range result.PackageCoverage {
		if strings.Contains(pkg.Path, packagePath) {
			return pkg, nil
		}
	}

	return nil, fmt.Errorf("package not found: %s", packagePath)
}

// GetFunctionsByFile returns functions grouped by file
func GetFunctionsByFile(result *models.AnalysisResult) map[string][]*models.Function {
	fileMap := make(map[string][]*models.Function)

	for _, pkg := range result.PackageCoverage {
		for _, file := range pkg.Files {
			fileMap[file.Path] = file.Functions
		}
	}

	return fileMap
}

// CalculatePackageMetrics calculates additional metrics for packages
func CalculatePackageMetrics(pkg *models.Package) *models.PackageMetrics {
	metrics := &models.PackageMetrics{
		PackageName:       pkg.Name,
		TotalFunctions:    pkg.TotalFunctions,
		TestedFunctions:   pkg.CoveredFunctions,
		UntestedFunctions: pkg.TotalFunctions - pkg.CoveredFunctions,
		FunctionCoverage:  pkg.FunctionCoverage,
		LineCoverage:      pkg.LineCoverage,
		ComplexityScore:   pkg.Complexity,
		TestableExported:  0,
		UntestedExported:  0,
		TestablePrivate:   0,
		UntestedPrivate:   0,
	}

	// Calculate exported vs private function metrics
	for _, file := range pkg.Files {
		for _, function := range file.Functions {
			if !function.IsTestable {
				continue
			}

			if function.IsExported {
				metrics.TestableExported++
				if !function.IsCovered {
					metrics.UntestedExported++
				}
			} else {
				metrics.TestablePrivate++
				if !function.IsCovered {
					metrics.UntestedPrivate++
				}
			}
		}
	}

	return metrics
}
