package coverage

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/beck/go-coverage-analyzer/pkg/models"
)

// AnalysisEngine handles comprehensive coverage analysis
type AnalysisEngine struct {
	parser  *ProfileParser
	verbose bool
	fset    *token.FileSet
}

// NewAnalysisEngine creates a new coverage analysis engine
func NewAnalysisEngine(verbose bool) *AnalysisEngine {
	return &AnalysisEngine{
		parser:  NewProfileParser(verbose),
		verbose: verbose,
		fset:    token.NewFileSet(),
	}
}

// AnalyzeProject performs comprehensive coverage analysis on a Go project
func (e *AnalysisEngine) AnalyzeProject(opts *AnalysisOptions) (*models.AnalysisResult, error) {
	startTime := time.Now()

	if e.verbose {
		fmt.Printf("🔍 Starting comprehensive coverage analysis of: %s\n", opts.ProjectPath)
	}

	// Step 1: Get project information
	projectInfo, err := e.parser.GetProjectInfo(opts.ProjectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze project: %w", err)
	}

	// Step 2: Generate coverage profile if requested
	var profilePath string
	if opts.GenerateProfile {
		profilePath = opts.ProfileOutput
		if profilePath == "" {
			profilePath = filepath.Join(opts.ProjectPath, "coverage.out")
		}

		if err := e.parser.GenerateProfile(opts.ProjectPath, profilePath, opts.PackagePattern); err != nil {
			return nil, fmt.Errorf("failed to generate coverage profile: %w", err)
		}

		// Ensure the profile path is absolute for parsing
		if !filepath.IsAbs(profilePath) {
			profilePath = filepath.Join(opts.ProjectPath, profilePath)
		}
	} else if opts.ProfilePath != "" {
		profilePath = opts.ProfilePath
		if !filepath.IsAbs(profilePath) {
			profilePath = filepath.Join(opts.ProjectPath, profilePath)
		}
	} else {
		// Step 2.5: Check for existing coverage profiles automatically
		defaultProfilePath := filepath.Join(opts.ProjectPath, "coverage.out")
		if _, err := os.Stat(defaultProfilePath); err == nil {
			profilePath = defaultProfilePath
			if e.verbose {
				fmt.Printf("📋 Found existing coverage profile: %s\n", profilePath)
			}
		}
	}

	// Step 3: Parse coverage profile
	var profile *models.CoverageProfile
	if profilePath != "" {
		profile, err = e.parser.ParseProfile(profilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse coverage profile: %w", err)
		}

		if err := e.parser.ValidateProfile(profile); err != nil {
			return nil, fmt.Errorf("invalid coverage profile: %w", err)
		}
	}

	// Step 4: Parse source files and build AST
	packages, err := e.parseSourceFiles(opts.ProjectPath, opts.ExcludeDirs, opts.IncludeTests)
	if err != nil {
		return nil, fmt.Errorf("failed to parse source files: %w", err)
	}

	// Step 5: Analyze coverage and identify gaps
	result, err := e.buildAnalysisResult(packages, profile, projectInfo, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to build analysis result: %w", err)
	}

	// Step 6: Calculate summary statistics
	e.calculateSummaryStatistics(result)

	result.Metadata.AnalysisTime = time.Since(startTime)
	result.Metadata.ProfilePath = profilePath

	if e.verbose {
		fmt.Printf("✅ Analysis completed in %v\n", result.Metadata.AnalysisTime)
		fmt.Printf("📊 Found %d packages, %d files, %d functions\n",
			result.Summary.TotalPackages,
			result.Summary.TotalFiles,
			result.Summary.TotalFunctions)
	}

	return result, nil
}

// AnalysisOptions contains options for coverage analysis
type AnalysisOptions struct {
	ProjectPath         string
	ProfilePath         string
	ProfileOutput       string
	PackagePattern      string
	ExcludeDirs         []string
	IncludeTests        bool
	GenerateProfile     bool
	CalculateComplexity bool
	MinComplexity       int
}

// parseSourceFiles parses all Go source files in the project
func (e *AnalysisEngine) parseSourceFiles(projectPath string, excludeDirs []string, includeTests bool) (map[string]*models.Package, error) {
	packages := make(map[string]*models.Package)

	excludeMap := make(map[string]bool)
	for _, dir := range excludeDirs {
		excludeMap[dir] = true
	}

	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Skip excluded directories
			if excludeMap[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip test files if not included
		if strings.HasSuffix(path, "_test.go") && !includeTests {
			return nil
		}

		return e.parseGoFile(path, projectPath, packages)
	})

	return packages, err
}

// parseGoFile parses a single Go file and extracts functions
func (e *AnalysisEngine) parseGoFile(filePath, projectPath string, packages map[string]*models.Package) error {
	src, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Parse the file
	file, err := parser.ParseFile(e.fset, filePath, src, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse file %s: %w", filePath, err)
	}

	packageName := file.Name.Name
	if packages[packageName] == nil {
		relPath, _ := filepath.Rel(projectPath, filepath.Dir(filePath))
		packages[packageName] = &models.Package{
			Name:  packageName,
			Path:  relPath,
			Files: make(map[string]*models.File),
		}
	}

	relFilePath, _ := filepath.Rel(projectPath, filePath)
	fileModel := &models.File{
		Name:      filepath.Base(filePath),
		Path:      relFilePath,
		Package:   packageName,
		Functions: make([]*models.Function, 0),
		HasTests:  strings.HasSuffix(filePath, "_test.go"),
	}

	// Extract functions from the AST
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			function := e.extractFunction(node, fileModel, string(src))
			if function != nil {
				fileModel.Functions = append(fileModel.Functions, function)
			}
		}
		return true
	})

	// Count lines
	lines := strings.Split(string(src), "\n")
	fileModel.TotalLines = len(lines)

	packages[packageName].Files[relFilePath] = fileModel
	return nil
}

// extractFunction extracts function information from an AST function declaration
func (e *AnalysisEngine) extractFunction(funcDecl *ast.FuncDecl, file *models.File, source string) *models.Function {
	if funcDecl.Name == nil {
		return nil
	}

	position := e.fset.Position(funcDecl.Pos())
	endPosition := e.fset.Position(funcDecl.End())

	function := &models.Function{
		Name:         funcDecl.Name.Name,
		File:         file.Path,
		Package:      file.Package,
		StartLine:    position.Line,
		EndLine:      endPosition.Line,
		IsMethod:     funcDecl.Recv != nil,
		IsExported:   ast.IsExported(funcDecl.Name.Name),
		Parameters:   make([]*models.Param, 0),
		ReturnTypes:  make([]string, 0),
		Dependencies: make([]string, 0),
		IsTestable:   true,
	}

	// Extract receiver type for methods
	if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
		if recv := funcDecl.Recv.List[0]; recv.Type != nil {
			function.ReceiverType = e.extractTypeName(recv.Type)
		}
	}

	// Extract parameters
	if funcDecl.Type.Params != nil {
		for _, param := range funcDecl.Type.Params.List {
			paramType := e.extractTypeName(param.Type)
			for _, name := range param.Names {
				function.Parameters = append(function.Parameters, &models.Param{
					Name: name.Name,
					Type: paramType,
				})
			}
		}
	}

	// Extract return types
	if funcDecl.Type.Results != nil {
		for _, result := range funcDecl.Type.Results.List {
			resultType := e.extractTypeName(result.Type)
			function.ReturnTypes = append(function.ReturnTypes, resultType)

			// Check for error return type
			if resultType == "error" {
				function.HasErrorReturn = true
			}
		}
	}

	// Build function signature
	function.Signature = e.buildFunctionSignature(function)

	// Calculate cyclomatic complexity (simplified)
	function.Complexity = e.calculateComplexity(funcDecl)

	// Determine if function is testable
	function.IsTestable = e.isFunctionTestable(function)

	// Check if it's a test function itself
	if strings.HasPrefix(function.Name, "Test") && file.HasTests {
		function.IsTestable = false
	}

	return function
}

// extractTypeName extracts type name from AST expression
func (e *AnalysisEngine) extractTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		if pkg, ok := t.X.(*ast.Ident); ok {
			return pkg.Name + "." + t.Sel.Name
		}
		return t.Sel.Name
	case *ast.StarExpr:
		return "*" + e.extractTypeName(t.X)
	case *ast.ArrayType:
		return "[]" + e.extractTypeName(t.Elt)
	case *ast.MapType:
		return "map[" + e.extractTypeName(t.Key) + "]" + e.extractTypeName(t.Value)
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.StructType:
		return "struct{}"
	case *ast.FuncType:
		return "func"
	default:
		return "unknown"
	}
}

// buildFunctionSignature creates a function signature string
func (e *AnalysisEngine) buildFunctionSignature(function *models.Function) string {
	var sig strings.Builder

	sig.WriteString("func ")

	if function.ReceiverType != "" {
		sig.WriteString("(")
		sig.WriteString(function.ReceiverType)
		sig.WriteString(") ")
	}

	sig.WriteString(function.Name)
	sig.WriteString("(")

	for i, param := range function.Parameters {
		if i > 0 {
			sig.WriteString(", ")
		}
		sig.WriteString(param.Name)
		sig.WriteString(" ")
		sig.WriteString(param.Type)
	}

	sig.WriteString(")")

	if len(function.ReturnTypes) > 0 {
		if len(function.ReturnTypes) == 1 {
			sig.WriteString(" ")
			sig.WriteString(function.ReturnTypes[0])
		} else {
			sig.WriteString(" (")
			sig.WriteString(strings.Join(function.ReturnTypes, ", "))
			sig.WriteString(")")
		}
	}

	return sig.String()
}

// calculateComplexity calculates cyclomatic complexity (simplified implementation)
func (e *AnalysisEngine) calculateComplexity(funcDecl *ast.FuncDecl) int {
	complexity := 1 // Base complexity

	ast.Inspect(funcDecl, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.SwitchStmt,
			*ast.TypeSwitchStmt, *ast.SelectStmt:
			complexity++
		case *ast.CaseClause:
			complexity++
		}
		return true
	})

	return complexity
}

// isFunctionTestable determines if a function should have tests generated
func (e *AnalysisEngine) isFunctionTestable(function *models.Function) bool {
	// Skip test functions themselves
	if strings.HasPrefix(function.Name, "Test") ||
		strings.HasPrefix(function.Name, "Benchmark") ||
		strings.HasPrefix(function.Name, "Example") {
		return false
	}

	// Skip init functions
	if function.Name == "init" {
		return false
	}

	// Skip main function
	if function.Name == "main" {
		return false
	}

	// Only test exported functions and methods
	return function.IsExported
}

// buildAnalysisResult combines AST analysis with coverage data
func (e *AnalysisEngine) buildAnalysisResult(packages map[string]*models.Package, profile *models.CoverageProfile, projectInfo *models.ProjectInfo, opts *AnalysisOptions) (*models.AnalysisResult, error) {
	result := &models.AnalysisResult{
		ProjectPath:        opts.ProjectPath,
		Timestamp:          time.Now(),
		PackageCoverage:    packages,
		UncoveredFunctions: make([]*models.Function, 0),
		Summary:            &models.Summary{},
		Metadata: &models.Metadata{
			Version:          "0.1.0",
			GoVersion:        projectInfo.GoVersion,
			ModulePath:       projectInfo.ModulePath,
			ExcludedDirs:     opts.ExcludeDirs,
			IncludedPackages: make([]string, 0),
		},
	}

	// Apply coverage data if available
	if profile != nil {
		e.applyCoverageData(packages, profile)
	}

	// Identify uncovered functions
	for _, pkg := range packages {
		for _, file := range pkg.Files {
			for _, function := range file.Functions {
				if function.IsTestable && !function.IsCovered {
					result.UncoveredFunctions = append(result.UncoveredFunctions, function)
				}
			}
		}
	}

	// Sort uncovered functions by complexity (descending)
	sort.Slice(result.UncoveredFunctions, func(i, j int) bool {
		return result.UncoveredFunctions[i].Complexity > result.UncoveredFunctions[j].Complexity
	})

	return result, nil
}

// applyCoverageData applies coverage information to functions and files
func (e *AnalysisEngine) applyCoverageData(packages map[string]*models.Package, profile *models.CoverageProfile) {
	// Create a mapping of file paths to coverage blocks
	fileBlocks := make(map[string][]*models.ProfileBlock)

	if e.verbose {
		fmt.Printf("📋 Processing %d coverage blocks\n", len(profile.Blocks))
	}

	// Map profile blocks with multiple path variants
	for _, block := range profile.Blocks {
		// Store blocks under multiple keys to improve matching
		keys := e.generatePathVariants(block.FileName)
		for _, key := range keys {
			fileBlocks[key] = append(fileBlocks[key], block)
		}

		if e.verbose && len(fileBlocks) < 5 { // Only show first few for debugging
			fmt.Printf("   Block: %s -> %v\n", block.FileName, keys)
		}
	}

	// Apply coverage data to each function
	for _, pkg := range packages {
		for _, file := range pkg.Files {
			// Try multiple path variants to find matching blocks
			var blocks []*models.ProfileBlock
			pathVariants := e.generatePathVariants(file.Path)

			for _, variant := range pathVariants {
				if foundBlocks, exists := fileBlocks[variant]; exists {
					blocks = foundBlocks
					if e.verbose {
						fmt.Printf("   Matched file %s with variant %s (%d blocks)\n", file.Path, variant, len(blocks))
					}
					break
				}
			}

			if len(blocks) == 0 {
				if e.verbose {
					fmt.Printf("   No coverage blocks found for file %s (tried: %v)\n", file.Path, pathVariants)
				}
				continue
			}

			// Convert profile blocks to internal blocks
			file.CoverageBlocks = e.parser.ConvertToBlocks(blocks)

			// Apply coverage to functions
			for _, function := range file.Functions {
				function.Coverage, function.IsCovered = e.calculateFunctionCoverage(function, blocks)
			}

			// Calculate file-level coverage
			e.calculateFileCoverage(file)
		}

		// Calculate package-level coverage
		e.calculatePackageCoverage(pkg)
	}
}

// generatePathVariants creates multiple path variants to improve matching
func (e *AnalysisEngine) generatePathVariants(path string) []string {
	variants := []string{path} // Always include original path

	// Add filename only variant
	filename := filepath.Base(path)
	if filename != path {
		variants = append(variants, filename)
	}

	// If path contains module prefix, create relative variant
	if strings.Contains(path, "/") {
		parts := strings.Split(path, "/")
		if len(parts) > 1 {
			// Try removing first N parts to get relative paths
			for i := 1; i < len(parts); i++ {
				relativePath := strings.Join(parts[i:], "/")
				variants = append(variants, relativePath)
			}
		}
	}

	return variants
}

// calculateFunctionCoverage determines coverage percentage for a function
func (e *AnalysisEngine) calculateFunctionCoverage(function *models.Function, blocks []*models.ProfileBlock) (float64, bool) {
	totalStmts := 0
	coveredStmts := 0

	for _, block := range blocks {
		// Check if block overlaps with function
		if block.StartLine >= function.StartLine && block.EndLine <= function.EndLine {
			totalStmts += block.NumStmts
			if block.Count > 0 {
				coveredStmts += block.NumStmts
			}
		}
	}

	if totalStmts == 0 {
		return 0.0, false
	}

	coverage := float64(coveredStmts) / float64(totalStmts) * 100.0
	return coverage, coveredStmts > 0
}

// calculateFileCoverage calculates coverage statistics for a file
func (e *AnalysisEngine) calculateFileCoverage(file *models.File) {
	totalStmts := 0
	coveredStmts := 0
	coveredFunctions := 0

	for _, block := range file.CoverageBlocks {
		totalStmts += block.NumStmts
		if block.IsCovered {
			coveredStmts += block.NumStmts
		}
	}

	for _, function := range file.Functions {
		if function.IsCovered {
			coveredFunctions++
		}
	}

	if totalStmts > 0 {
		file.Coverage = float64(coveredStmts) / float64(totalStmts) * 100.0
		file.LineCoverage = file.Coverage
	}

	if len(file.Functions) > 0 {
		file.FunctionCoverage = float64(coveredFunctions) / float64(len(file.Functions)) * 100.0
	}

	file.CoveredLines = coveredStmts
	file.UncoveredLines = totalStmts - coveredStmts
}

// calculatePackageCoverage calculates coverage statistics for a package
func (e *AnalysisEngine) calculatePackageCoverage(pkg *models.Package) {
	totalStmts := 0
	coveredStmts := 0
	totalFunctions := 0
	coveredFunctions := 0
	totalComplexity := 0

	for _, file := range pkg.Files {
		totalStmts += file.CoveredLines + file.UncoveredLines
		coveredStmts += file.CoveredLines
		totalFunctions += len(file.Functions)

		for _, function := range file.Functions {
			if function.IsCovered {
				coveredFunctions++
			}
			totalComplexity += function.Complexity
		}
	}

	if totalStmts > 0 {
		pkg.Coverage = float64(coveredStmts) / float64(totalStmts) * 100.0
		pkg.LineCoverage = pkg.Coverage
	}

	if totalFunctions > 0 {
		pkg.FunctionCoverage = float64(coveredFunctions) / float64(totalFunctions) * 100.0
	}

	pkg.TotalLines = totalStmts
	pkg.CoveredLines = coveredStmts
	pkg.UncoveredLines = totalStmts - coveredStmts
	pkg.TotalFunctions = totalFunctions
	pkg.CoveredFunctions = coveredFunctions
	pkg.Complexity = totalComplexity
}

// calculateSummaryStatistics calculates overall project statistics
func (e *AnalysisEngine) calculateSummaryStatistics(result *models.AnalysisResult) {
	summary := result.Summary

	totalStmts := 0
	coveredStmts := 0
	totalFunctions := 0
	coveredFunctions := 0
	totalComplexity := 0
	maxComplexity := 0
	highComplexityCount := 0
	publicFunctions := 0
	privateFunctions := 0
	coveredPublic := 0
	coveredPrivate := 0
	methods := 0
	coveredMethods := 0

	for _, pkg := range result.PackageCoverage {
		summary.TotalPackages++

		for _, file := range pkg.Files {
			summary.TotalFiles++
			totalStmts += file.CoveredLines + file.UncoveredLines
			coveredStmts += file.CoveredLines

			for _, function := range file.Functions {
				if !function.IsTestable {
					continue
				}

				totalFunctions++
				totalComplexity += function.Complexity

				if function.Complexity > maxComplexity {
					maxComplexity = function.Complexity
				}

				if function.Complexity > 10 { // High complexity threshold
					highComplexityCount++
				}

				if function.IsExported {
					publicFunctions++
					if function.IsCovered {
						coveredPublic++
					}
				} else {
					privateFunctions++
					if function.IsCovered {
						coveredPrivate++
					}
				}

				if function.IsMethod {
					methods++
					if function.IsCovered {
						coveredMethods++
					}
				}

				if function.IsCovered {
					coveredFunctions++
				}
			}
		}
	}

	// Calculate coverage percentages
	if totalStmts > 0 {
		result.OverallCoverage = float64(coveredStmts) / float64(totalStmts) * 100.0
		result.LineCoverage = result.OverallCoverage
		summary.LineCoverage = result.LineCoverage
		summary.OverallCoverage = result.OverallCoverage
	}

	if totalFunctions > 0 {
		result.FunctionCoverage = float64(coveredFunctions) / float64(totalFunctions) * 100.0
		summary.FunctionCoverage = result.FunctionCoverage
	}

	if publicFunctions > 0 {
		summary.PublicFunctionCoverage = float64(coveredPublic) / float64(publicFunctions) * 100.0
	}

	if privateFunctions > 0 {
		summary.PrivateFunctionCoverage = float64(coveredPrivate) / float64(privateFunctions) * 100.0
	}

	if methods > 0 {
		summary.MethodCoverage = float64(coveredMethods) / float64(methods) * 100.0
	}

	// Set summary values
	summary.TotalLines = totalStmts
	summary.CoveredLines = coveredStmts
	summary.UncoveredLines = totalStmts - coveredStmts
	summary.TotalFunctions = totalFunctions
	summary.TestedFunctions = coveredFunctions
	summary.UntestedFunctions = totalFunctions - coveredFunctions
	summary.TotalComplexity = totalComplexity
	summary.MaxComplexity = maxComplexity
	summary.HighComplexityFunctions = highComplexityCount

	if totalFunctions > 0 {
		summary.AvgComplexity = float64(totalComplexity) / float64(totalFunctions)
	}

	// Set branch coverage (simplified - same as line coverage for now)
	result.BranchCoverage = result.LineCoverage
	summary.BranchCoverage = result.BranchCoverage
}
