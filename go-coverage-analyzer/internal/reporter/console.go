package reporter

import (
	"fmt"
	"strings"

	"github.com/beck/go-coverage-analyzer/pkg/models"
)

// GenerateComparisonReport generates a comparison report between two analysis results
func GenerateComparisonReport(current, previous *models.AnalysisResult, opts *Options) error {
	fmt.Printf("%s%sCOVERAGE COMPARISON REPORT%s\n", ColorBold, ColorCyan, ColorReset)
	fmt.Println("=" + strings.Repeat("=", 60) + "=")

	// Overall comparison
	fmt.Printf("Current Coverage:  %s%.1f%%%s\n",
		getCoverageColor(current.OverallCoverage, opts.Threshold), current.OverallCoverage, ColorReset)
	fmt.Printf("Previous Coverage: %s%.1f%%%s\n",
		getCoverageColor(previous.OverallCoverage, opts.Threshold), previous.OverallCoverage, ColorReset)

	change := current.OverallCoverage - previous.OverallCoverage
	changeColor := ColorGreen
	changeSymbol := "📈"

	if change < 0 {
		changeColor = ColorRed
		changeSymbol = "📉"
	} else if change == 0 {
		changeColor = ColorYellow
		changeSymbol = "➡️"
	}

	fmt.Printf("Change:            %s%s %.2f%% (%s%.1f%%)%s\n",
		changeColor, changeSymbol, change, changeColor, change, ColorReset)

	fmt.Println()

	// Function comparison
	currentUntested := current.Summary.UntestedFunctions
	previousUntested := previous.Summary.UntestedFunctions
	functionChange := previousUntested - currentUntested

	fmt.Printf("Current Untested Functions:  %s%d%s\n", ColorRed, currentUntested, ColorReset)
	fmt.Printf("Previous Untested Functions: %s%d%s\n", ColorRed, previousUntested, ColorReset)

	if functionChange > 0 {
		fmt.Printf("Improvement: %s✅ %d functions now tested%s\n", ColorGreen, functionChange, ColorReset)
	} else if functionChange < 0 {
		fmt.Printf("Regression:  %s❌ %d additional untested functions%s\n", ColorRed, -functionChange, ColorReset)
	} else {
		fmt.Printf("No change in function coverage\n")
	}

	fmt.Println()

	// Package-level changes
	fmt.Printf("%s%sPACKAGE CHANGES%s\n", ColorBold, ColorWhite, ColorReset)
	fmt.Println(strings.Repeat("-", 70))
	fmt.Printf("%-20s %-12s %-12s %-12s\n", "Package", "Current", "Previous", "Change")
	fmt.Println(strings.Repeat("-", 70))

	for pkgName, currentPkg := range current.PackageCoverage {
		previousPkg, exists := previous.PackageCoverage[pkgName]
		if !exists {
			fmt.Printf("%-20s %s%7.1f%%%s %12s %s+NEW%s\n",
				truncate(pkgName, 20),
				getCoverageColor(currentPkg.Coverage, opts.Threshold), currentPkg.Coverage, ColorReset,
				"N/A",
				ColorGreen, ColorReset)
			continue
		}

		pkgChange := currentPkg.Coverage - previousPkg.Coverage
		changeSymbol := "="
		changeColor := ColorYellow

		if pkgChange > 0.1 {
			changeSymbol = "+"
			changeColor = ColorGreen
		} else if pkgChange < -0.1 {
			changeSymbol = "-"
			changeColor = ColorRed
		}

		fmt.Printf("%-20s %s%7.1f%%%s %s%7.1f%%%s %s%s%6.1f%%%s\n",
			truncate(pkgName, 20),
			getCoverageColor(currentPkg.Coverage, opts.Threshold), currentPkg.Coverage, ColorReset,
			getCoverageColor(previousPkg.Coverage, opts.Threshold), previousPkg.Coverage, ColorReset,
			changeColor, changeSymbol, pkgChange, ColorReset)
	}

	// Identify removed packages
	for pkgName := range previous.PackageCoverage {
		if _, exists := current.PackageCoverage[pkgName]; !exists {
			fmt.Printf("%-20s %12s %12s %sREMOVED%s\n",
				truncate(pkgName, 20), "N/A", "N/A", ColorRed, ColorReset)
		}
	}

	fmt.Println()

	// Summary recommendations
	fmt.Printf("%s%sRECOMMENDATIONS%s\n", ColorBold, ColorWhite, ColorReset)
	fmt.Println(strings.Repeat("-", 30))

	if change > 0 {
		fmt.Printf("✅ Great improvement! Coverage increased by %.1f%%\n", change)
	} else if change < 0 {
		fmt.Printf("⚠️  Coverage decreased by %.1f%%. Consider reviewing recent changes\n", -change)
	}

	if functionChange > 0 {
		fmt.Printf("🎯 Excellent! %d more functions are now tested\n", functionChange)
	} else if functionChange < 0 {
		fmt.Printf("📝 %d functions lost test coverage. Review recent changes\n", -functionChange)
	}

	if current.OverallCoverage < opts.Threshold {
		fmt.Printf("🎯 Focus on reaching %.1f%% coverage threshold\n", opts.Threshold)
	}

	return nil
}

// PrintCoverageTrend prints a simple coverage trend
func PrintCoverageTrend(trends []*models.CoverageTrend) {
	if len(trends) == 0 {
		return
	}

	fmt.Printf("%s%sCOVERAGE TREND%s\n", ColorBold, ColorCyan, ColorReset)
	fmt.Println(strings.Repeat("-", 40))

	for i, trend := range trends {
		symbol := "➡️"
		color := ColorYellow

		switch trend.Direction {
		case "up":
			symbol = "📈"
			color = ColorGreen
		case "down":
			symbol = "📉"
			color = ColorRed
		}

		fmt.Printf("%d. %s %s%.1f%%%s (%s%+.1f%%%s)\n",
			i+1, symbol,
			getCoverageColor(trend.CurrentCoverage, 80.0), trend.CurrentCoverage, ColorReset,
			color, trend.Change, ColorReset)
	}

	fmt.Println()
}

// PrintDetailedPackageReport prints detailed information about a specific package
func PrintDetailedPackageReport(pkg *models.Package, opts *Options) {
	fmt.Printf("%s%sPACKAGE DETAILS: %s%s\n", ColorBold, ColorCyan, pkg.Name, ColorReset)
	fmt.Println("=" + strings.Repeat("=", 50) + "=")

	// Package summary
	fmt.Printf("Path: %s\n", pkg.Path)
	fmt.Printf("Coverage: %s%.1f%%%s\n", getCoverageColor(pkg.Coverage, opts.Threshold), pkg.Coverage, ColorReset)
	fmt.Printf("Functions: %d total, %d covered, %d uncovered\n",
		pkg.TotalFunctions, pkg.CoveredFunctions, pkg.TotalFunctions-pkg.CoveredFunctions)
	fmt.Printf("Lines: %d total, %d covered, %d uncovered\n",
		pkg.TotalLines, pkg.CoveredLines, pkg.UncoveredLines)
	fmt.Printf("Complexity: %d\n", pkg.Complexity)

	fmt.Println()

	// File details
	if len(pkg.Files) > 0 {
		fmt.Printf("%s%sFILES%s\n", ColorBold, ColorWhite, ColorReset)
		fmt.Println(strings.Repeat("-", 80))
		fmt.Printf("%-30s %-10s %-10s %-10s %-8s\n", "File", "Coverage", "Functions", "Lines", "Complex")
		fmt.Println(strings.Repeat("-", 80))

		for _, file := range pkg.Files {
			fmt.Printf("%-30s %s%7.1f%%%s %7d/%-3d %7d/%-5d %8d\n",
				truncate(file.Name, 30),
				getCoverageColor(file.Coverage, opts.Threshold), file.Coverage, ColorReset,
				len(file.Functions), len(file.Functions), // Simplified for now
				file.CoveredLines, file.TotalLines,
				file.Complexity)
		}

		fmt.Println()
	}

	// Uncovered functions in this package
	uncoveredFunctions := make([]*models.Function, 0)
	for _, file := range pkg.Files {
		for _, function := range file.Functions {
			if function.IsTestable && !function.IsCovered {
				uncoveredFunctions = append(uncoveredFunctions, function)
			}
		}
	}

	if len(uncoveredFunctions) > 0 {
		fmt.Printf("%s%sUNCOVERED FUNCTIONS%s\n", ColorBold, ColorWhite, ColorReset)
		fmt.Println(strings.Repeat("-", 70))
		fmt.Printf("%-25s %-20s %-10s %-10s\n", "Function", "File", "Lines", "Complexity")
		fmt.Println(strings.Repeat("-", 70))

		for _, function := range uncoveredFunctions[:min(10, len(uncoveredFunctions))] {
			complexityColor := ColorGreen
			if function.Complexity > 5 {
				complexityColor = ColorYellow
			}
			if function.Complexity > 10 {
				complexityColor = ColorRed
			}

			fmt.Printf("%-25s %-20s %4d-%-4d %s%8d%s\n",
				truncate(function.Name, 25),
				truncate(function.File, 20),
				function.StartLine, function.EndLine,
				complexityColor, function.Complexity, ColorReset)
		}

		if len(uncoveredFunctions) > 10 {
			fmt.Printf("\n... and %d more uncovered functions\n", len(uncoveredFunctions)-10)
		}
	}

	fmt.Println()
}

// PrintFunctionDetails prints detailed information about a specific function
func PrintFunctionDetails(function *models.Function) {
	fmt.Printf("%s%sFUNCTION DETAILS: %s%s\n", ColorBold, ColorCyan, function.Name, ColorReset)
	fmt.Println("=" + strings.Repeat("=", 40) + "=")

	fmt.Printf("Package: %s\n", function.Package)
	fmt.Printf("File: %s\n", function.File)
	fmt.Printf("Lines: %d-%d\n", function.StartLine, function.EndLine)
	fmt.Printf("Signature: %s\n", function.Signature)
	fmt.Printf("Coverage: %s%.1f%%%s\n", getCoverageColor(function.Coverage, 80.0), function.Coverage, ColorReset)

	complexityColor := ColorGreen
	if function.Complexity > 5 {
		complexityColor = ColorYellow
	}
	if function.Complexity > 10 {
		complexityColor = ColorRed
	}
	fmt.Printf("Complexity: %s%d%s\n", complexityColor, function.Complexity, ColorReset)

	// Properties
	fmt.Printf("Exported: %v\n", function.IsExported)
	fmt.Printf("Method: %v\n", function.IsMethod)
	fmt.Printf("Testable: %v\n", function.IsTestable)
	fmt.Printf("Has Tests: %v\n", function.HasTests)
	fmt.Printf("Has Error Return: %v\n", function.HasErrorReturn)

	if function.ReceiverType != "" {
		fmt.Printf("Receiver: %s\n", function.ReceiverType)
	}

	if len(function.Parameters) > 0 {
		fmt.Printf("Parameters: ")
		for i, param := range function.Parameters {
			if i > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%s %s", param.Name, param.Type)
		}
		fmt.Println()
	}

	if len(function.ReturnTypes) > 0 {
		fmt.Printf("Returns: %s\n", strings.Join(function.ReturnTypes, ", "))
	}

	if len(function.Dependencies) > 0 {
		fmt.Printf("Dependencies: %s\n", strings.Join(function.Dependencies, ", "))
	}

	fmt.Println()
}

// PrintProgressBar prints a simple progress bar for coverage
func PrintProgressBar(current, total int, width int) {
	if total == 0 {
		return
	}

	percentage := float64(current) / float64(total)
	filled := int(percentage * float64(width))

	fmt.Printf("[%s%s%s] %.1f%% (%d/%d)",
		ColorGreen, strings.Repeat("=", filled), ColorReset,
		percentage*100, current, total)
}
