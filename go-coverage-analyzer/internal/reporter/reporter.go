package reporter

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/beck/go-coverage-analyzer/pkg/coverage"
	"github.com/beck/go-coverage-analyzer/pkg/models"
)

// Options contains configuration for report generation
type Options struct {
	Format      string
	InputFile   string
	OutputFile  string
	OpenReport  bool
	Threshold   float64
	Verbose     bool
	ShowDetails bool
	SortBy      string // name, coverage, complexity
	FilterBy    string // all, uncovered, low-coverage
}

// Colors for terminal output
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorBold   = "\033[1m"
)

// Generate creates and outputs a coverage report based on the specified format
func Generate(result *models.AnalysisResult, opts *Options) error {
	switch strings.ToLower(opts.Format) {
	case "json":
		return generateJSONReport(result, opts)
	case "html":
		return generateHTMLReport(result, opts)
	case "xml":
		return generateXMLReport(result, opts)
	case "console", "":
		return generateConsoleReport(result, opts)
	default:
		return fmt.Errorf("unsupported output format: %s", opts.Format)
	}
}

// GenerateFromProfile generates a report from an existing coverage profile
func GenerateFromProfile(projectPath string, opts *Options) error {
	if opts.Verbose {
		fmt.Printf("📋 Generating report from profile: %s\n", opts.InputFile)
	}

	parser := coverage.NewProfileParser(opts.Verbose)

	// Parse the existing profile
	_, err := parser.ParseProfile(opts.InputFile)
	if err != nil {
		return fmt.Errorf("failed to parse profile: %w", err)
	}

	// Create analysis engine and analyze with existing profile
	engine := coverage.NewAnalysisEngine(opts.Verbose)
	analysisOpts := &coverage.AnalysisOptions{
		ProjectPath:     projectPath,
		ProfilePath:     opts.InputFile,
		ExcludeDirs:     []string{"vendor", ".git", "node_modules"},
		IncludeTests:    false,
		GenerateProfile: false,
	}

	result, err := engine.AnalyzeProject(analysisOpts)
	if err != nil {
		return fmt.Errorf("failed to analyze project with profile: %w", err)
	}

	return Generate(result, opts)
}

// generateConsoleReport creates a rich human-readable console report
func generateConsoleReport(result *models.AnalysisResult, opts *Options) error {
	printHeader(result)
	printOverallSummary(result, opts.Threshold)

	if opts.ShowDetails {
		printPackageDetails(result, opts)
		printUncoveredFunctions(result, opts)
		printComplexityAnalysis(result)
	}

	printRecommendations(result, opts.Threshold)
	return nil
}

// printHeader prints the report header
func printHeader(result *models.AnalysisResult) {
	fmt.Printf("%s%s", ColorBold, ColorCyan)
	fmt.Println("=" + strings.Repeat("=", 70) + "=")
	fmt.Println("                GO COVERAGE ANALYSIS REPORT")
	fmt.Println("=" + strings.Repeat("=", 70) + "=")
	fmt.Printf("%s", ColorReset)
	fmt.Printf("Project: %s%s%s\n", ColorBold, result.ProjectPath, ColorReset)
	fmt.Printf("Analysis Time: %s%v%s\n", ColorBlue, result.Metadata.AnalysisTime, ColorReset)
	fmt.Printf("Timestamp: %s%s%s\n", ColorBlue, result.Timestamp.Format("2006-01-02 15:04:05"), ColorReset)

	if result.Metadata.ModulePath != "" {
		fmt.Printf("Module: %s%s%s\n", ColorBlue, result.Metadata.ModulePath, ColorReset)
	}

	if result.Metadata.GoVersion != "" {
		fmt.Printf("Go Version: %s%s%s\n", ColorBlue, result.Metadata.GoVersion, ColorReset)
	}

	fmt.Println()
}

// printOverallSummary prints the overall coverage summary
func printOverallSummary(result *models.AnalysisResult, threshold float64) {
	summary := result.Summary

	fmt.Printf("%s%sOVERALL SUMMARY%s\n", ColorBold, ColorWhite, ColorReset)
	fmt.Println(strings.Repeat("-", 50))

	// Overall coverage with color coding
	coverageColor := getCoverageColor(result.OverallCoverage, threshold)
	fmt.Printf("Overall Coverage:        %s%.1f%%%s\n", coverageColor, result.OverallCoverage, ColorReset)
	fmt.Printf("Function Coverage:       %s%.1f%%%s\n", getCoverageColor(summary.FunctionCoverage, threshold), summary.FunctionCoverage, ColorReset)
	fmt.Printf("Line Coverage:           %s%.1f%%%s\n", getCoverageColor(summary.LineCoverage, threshold), summary.LineCoverage, ColorReset)
	fmt.Printf("Branch Coverage:         %s%.1f%%%s\n", getCoverageColor(summary.BranchCoverage, threshold), summary.BranchCoverage, ColorReset)

	fmt.Println()

	// Project statistics
	fmt.Printf("Total Packages:          %s%d%s\n", ColorCyan, summary.TotalPackages, ColorReset)
	fmt.Printf("Total Files:             %s%d%s\n", ColorCyan, summary.TotalFiles, ColorReset)
	fmt.Printf("Total Functions:         %s%d%s\n", ColorCyan, summary.TotalFunctions, ColorReset)
	fmt.Printf("Tested Functions:        %s%s%d%s\n", ColorGreen, ColorBold, summary.TestedFunctions, ColorReset)
	fmt.Printf("Untested Functions:      %s%s%d%s\n", ColorRed, ColorBold, summary.UntestedFunctions, ColorReset)

	fmt.Println()

	// Line statistics
	fmt.Printf("Total Lines:             %s%d%s\n", ColorCyan, summary.TotalLines, ColorReset)
	fmt.Printf("Covered Lines:           %s%s%d%s\n", ColorGreen, ColorBold, summary.CoveredLines, ColorReset)
	fmt.Printf("Uncovered Lines:         %s%s%d%s\n", ColorRed, ColorBold, summary.UncoveredLines, ColorReset)

	fmt.Println()

	// Complexity metrics
	if summary.TotalComplexity > 0 {
		fmt.Printf("Average Complexity:      %s%.1f%s\n", ColorYellow, summary.AvgComplexity, ColorReset)
		fmt.Printf("Max Complexity:          %s%d%s\n", ColorYellow, summary.MaxComplexity, ColorReset)
		fmt.Printf("High Complexity (>10):   %s%d%s\n", ColorYellow, summary.HighComplexityFunctions, ColorReset)
	}

	fmt.Println()

	// Coverage threshold check
	if result.OverallCoverage < threshold {
		fmt.Printf("%s⚠️  WARNING: Coverage %.1f%% is below threshold %.1f%%%s\n",
			ColorRed, result.OverallCoverage, threshold, ColorReset)
	} else {
		fmt.Printf("%s✅ Coverage %.1f%% meets threshold %.1f%%%s\n",
			ColorGreen, result.OverallCoverage, threshold, ColorReset)
	}

	fmt.Println()
}

// printPackageDetails prints detailed package information
func printPackageDetails(result *models.AnalysisResult, opts *Options) {
	packages := make([]*models.Package, 0, len(result.PackageCoverage))
	for _, pkg := range result.PackageCoverage {
		packages = append(packages, pkg)
	}

	// Sort packages
	switch opts.SortBy {
	case "coverage":
		sort.Slice(packages, func(i, j int) bool {
			return packages[i].Coverage < packages[j].Coverage
		})
	case "complexity":
		sort.Slice(packages, func(i, j int) bool {
			return packages[i].Complexity > packages[j].Complexity
		})
	default: // name
		sort.Slice(packages, func(i, j int) bool {
			return packages[i].Name < packages[j].Name
		})
	}

	fmt.Printf("%s%sPACKAGE DETAILS%s\n", ColorBold, ColorWhite, ColorReset)
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("%-20s %-10s %-10s %-10s %-10s %-8s\n", "Package", "Coverage", "Functions", "Lines", "Complexity", "Status")
	fmt.Println(strings.Repeat("-", 80))

	for _, pkg := range packages {
		// Apply filter
		if opts.FilterBy == "uncovered" && pkg.Coverage > 0 {
			continue
		}
		if opts.FilterBy == "low-coverage" && pkg.Coverage >= opts.Threshold {
			continue
		}

		status := "✅"
		statusColor := ColorGreen
		if pkg.Coverage < opts.Threshold {
			status = "⚠️"
			statusColor = ColorYellow
			if pkg.Coverage == 0 {
				status = "❌"
				statusColor = ColorRed
			}
		}

		fmt.Printf("%-20s %s%7.1f%%%s %7d/%-3d %7d/%-5d %10d %s%s%s\n",
			truncate(pkg.Name, 20),
			getCoverageColor(pkg.Coverage, opts.Threshold), pkg.Coverage, ColorReset,
			pkg.CoveredFunctions, pkg.TotalFunctions,
			pkg.CoveredLines, pkg.TotalLines,
			pkg.Complexity,
			statusColor, status, ColorReset,
		)
	}

	fmt.Println()
}

// printUncoveredFunctions prints top uncovered functions
func printUncoveredFunctions(result *models.AnalysisResult, opts *Options) {
	if len(result.UncoveredFunctions) == 0 {
		return
	}

	fmt.Printf("%s%sUNCOVERED FUNCTIONS (Top 20)%s\n", ColorBold, ColorWhite, ColorReset)
	fmt.Println(strings.Repeat("-", 90))
	fmt.Printf("%-25s %-20s %-15s %-10s %-15s\n", "Function", "Package", "File", "Complexity", "Type")
	fmt.Println(strings.Repeat("-", 90))

	count := 0
	for _, function := range result.UncoveredFunctions {
		if count >= 20 {
			break
		}

		functionType := "Function"
		if function.IsMethod {
			functionType = "Method"
		}

		complexityColor := ColorGreen
		if function.Complexity > 5 {
			complexityColor = ColorYellow
		}
		if function.Complexity > 10 {
			complexityColor = ColorRed
		}

		fmt.Printf("%-25s %-20s %-15s %s%7d%s     %-15s\n",
			truncate(function.Name, 25),
			truncate(function.Package, 20),
			truncate(filepath.Base(function.File), 15),
			complexityColor, function.Complexity, ColorReset,
			functionType,
		)
		count++
	}

	if len(result.UncoveredFunctions) > 20 {
		fmt.Printf("\n%s... and %d more uncovered functions%s\n",
			ColorYellow, len(result.UncoveredFunctions)-20, ColorReset)
	}

	fmt.Println()
}

// printComplexityAnalysis prints complexity analysis
func printComplexityAnalysis(result *models.AnalysisResult) {
	highComplexity := result.GetHighComplexityFunctions(10)
	if len(highComplexity) == 0 {
		return
	}

	fmt.Printf("%s%sHIGH COMPLEXITY FUNCTIONS (>10)%s\n", ColorBold, ColorWhite, ColorReset)
	fmt.Println(strings.Repeat("-", 90))
	fmt.Printf("%-25s %-20s %-15s %-10s %-10s\n", "Function", "Package", "File", "Complexity", "Covered")
	fmt.Println(strings.Repeat("-", 90))

	for _, function := range highComplexity[:min(10, len(highComplexity))] {
		coveredStatus := "❌"
		coveredColor := ColorRed
		if function.IsCovered {
			coveredStatus = "✅"
			coveredColor = ColorGreen
		}

		fmt.Printf("%-25s %-20s %-15s %s%7d%s     %s%s%s\n",
			truncate(function.Name, 25),
			truncate(function.Package, 20),
			truncate(filepath.Base(function.File), 15),
			ColorRed, function.Complexity, ColorReset,
			coveredColor, coveredStatus, ColorReset,
		)
	}

	fmt.Println()
}

// printRecommendations prints actionable recommendations
func printRecommendations(result *models.AnalysisResult, threshold float64) {
	fmt.Printf("%s%sRECOMMENDATIONS%s\n", ColorBold, ColorWhite, ColorReset)
	fmt.Println(strings.Repeat("-", 50))

	if result.Summary.UntestedFunctions > 0 {
		fmt.Printf("🎯 Add tests for %s%d%s untested functions\n",
			ColorBold, result.Summary.UntestedFunctions, ColorReset)
	}

	if result.OverallCoverage < threshold {
		needed := threshold - result.OverallCoverage
		fmt.Printf("📈 Increase coverage by %s%.1f%%%s to meet threshold\n",
			ColorYellow, needed, ColorReset)
	}

	highComplexity := result.GetHighComplexityFunctions(10)
	uncoveredHighComplexity := 0
	for _, fn := range highComplexity {
		if !fn.IsCovered {
			uncoveredHighComplexity++
		}
	}

	if uncoveredHighComplexity > 0 {
		fmt.Printf("⚠️  Prioritize testing %s%d%s high-complexity functions\n",
			ColorRed, uncoveredHighComplexity, ColorReset)
	}

	fmt.Printf("🛠️  Run %s'gcov generate'%s to create test templates\n",
		ColorCyan, ColorReset)
	fmt.Printf("📊 Run %s'gcov report --format=html'%s for detailed HTML report\n",
		ColorCyan, ColorReset)

	fmt.Println()
}

// generateJSONReport creates a JSON report
func generateJSONReport(result *models.AnalysisResult, opts *Options) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return writeOutput(string(data), opts.OutputFile)
}

// generateHTMLReport creates a comprehensive HTML report
func generateHTMLReport(result *models.AnalysisResult, opts *Options) error {
	htmlContent := generateHTMLContent(result, opts)

	outputFile := opts.OutputFile
	if outputFile == "" {
		outputFile = "coverage-report.html"
	}

	if err := writeOutput(htmlContent, outputFile); err != nil {
		return err
	}

	if opts.OpenReport {
		return openInBrowser(outputFile)
	}

	if opts.Verbose {
		fmt.Printf("📄 HTML report generated: %s\n", outputFile)
	}

	return nil
}

// generateHTMLContent creates the HTML report content
func generateHTMLContent(result *models.AnalysisResult, opts *Options) string {
	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Go Coverage Report - {{.ProjectPath}}</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            margin: 0;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container { max-width: 1200px; margin: 0 auto; background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .header { text-align: center; margin-bottom: 40px; padding: 20px; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; border-radius: 8px; }
        .summary { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; margin-bottom: 30px; }
        .metric { background: #f8f9fa; padding: 20px; border-radius: 8px; text-align: center; border-left: 4px solid #007bff; }
        .metric-value { font-size: 2em; font-weight: bold; color: #333; }
        .metric-label { color: #666; margin-top: 5px; }
        .coverage-good { color: #28a745; }
        .coverage-warning { color: #ffc107; }
        .coverage-danger { color: #dc3545; }
        .packages-table { width: 100%; border-collapse: collapse; margin: 20px 0; }
        .packages-table th, .packages-table td { padding: 12px; text-align: left; border-bottom: 1px solid #ddd; }
        .packages-table th { background-color: #f8f9fa; font-weight: 600; }
        .packages-table tr:hover { background-color: #f5f5f5; }
        .progress-bar { width: 100%; height: 20px; background-color: #e9ecef; border-radius: 10px; overflow: hidden; }
        .progress-fill { height: 100%; transition: width 0.3s ease; }
        .uncovered-list { list-style: none; padding: 0; }
        .uncovered-list li { padding: 10px; margin: 5px 0; background: #fff3cd; border-left: 4px solid #ffc107; border-radius: 4px; }
        .section { margin: 30px 0; }
        .section-title { font-size: 1.5em; font-weight: 600; margin-bottom: 15px; color: #333; }
        .complexity-high { color: #dc3545; font-weight: bold; }
        .complexity-medium { color: #ffc107; font-weight: bold; }
        .complexity-low { color: #28a745; }
        .footer { margin-top: 40px; text-align: center; color: #666; font-size: 0.9em; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Go Coverage Report</h1>
            <p>{{.ProjectPath}}</p>
            <p>Generated: {{.Timestamp.Format "2006-01-02 15:04:05"}}</p>
            {{if .Metadata.ModulePath}}<p>Module: {{.Metadata.ModulePath}}</p>{{end}}
        </div>

        <div class="summary">
            <div class="metric">
                <div class="metric-value {{getCoverageClass .OverallCoverage}}">{{printf "%.1f%%" .OverallCoverage}}</div>
                <div class="metric-label">Overall Coverage</div>
            </div>
            <div class="metric">
                <div class="metric-value {{getCoverageClass .Summary.FunctionCoverage}}">{{printf "%.1f%%" .Summary.FunctionCoverage}}</div>
                <div class="metric-label">Function Coverage</div>
            </div>
            <div class="metric">
                <div class="metric-value">{{.Summary.TotalFunctions}}</div>
                <div class="metric-label">Total Functions</div>
            </div>
            <div class="metric">
                <div class="metric-value coverage-danger">{{.Summary.UntestedFunctions}}</div>
                <div class="metric-label">Untested Functions</div>
            </div>
        </div>

        <div class="section">
            <h2 class="section-title">Package Coverage</h2>
            <table class="packages-table">
                <thead>
                    <tr>
                        <th>Package</th>
                        <th>Coverage</th>
                        <th>Functions</th>
                        <th>Lines</th>
                        <th>Complexity</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .PackagesByName}}
                    <tr>
                        <td>{{.Name}}</td>
                        <td>
                            <div class="progress-bar">
                                <div class="progress-fill {{getCoverageClass .Coverage}}" style="width: {{.Coverage}}%; background-color: {{getCoverageColor .Coverage}};"></div>
                            </div>
                            <span class="{{getCoverageClass .Coverage}}">{{printf "%.1f%%" .Coverage}}</span>
                        </td>
                        <td>{{.CoveredFunctions}}/{{.TotalFunctions}}</td>
                        <td>{{.CoveredLines}}/{{.TotalLines}}</td>
                        <td>{{.Complexity}}</td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
        </div>

        {{if .UncoveredFunctions}}
        <div class="section">
            <h2 class="section-title">Top Uncovered Functions</h2>
            <ul class="uncovered-list">
                {{range .TopUncoveredFunctions}}
                <li>
                    <strong>{{.Name}}</strong> in {{.Package}}
                    <br><small>{{.File}}:{{.StartLine}} | Complexity: <span class="{{getComplexityClass .Complexity}}">{{.Complexity}}</span></small>
                </li>
                {{end}}
            </ul>
        </div>
        {{end}}

        <div class="footer">
            <p>Report generated by gcov v{{.Metadata.Version}} in {{.Metadata.AnalysisTime}}</p>
        </div>
    </div>
</body>
</html>`

	t, err := template.New("report").Funcs(template.FuncMap{
		"getCoverageClass": func(coverage float64) string {
			if coverage >= 80 {
				return "coverage-good"
			} else if coverage >= 60 {
				return "coverage-warning"
			}
			return "coverage-danger"
		},
		"getCoverageColor": func(coverage float64) string {
			if coverage >= 80 {
				return "#28a745"
			} else if coverage >= 60 {
				return "#ffc107"
			}
			return "#dc3545"
		},
		"getComplexityClass": func(complexity int) string {
			if complexity > 10 {
				return "complexity-high"
			} else if complexity > 5 {
				return "complexity-medium"
			}
			return "complexity-low"
		},
	}).Parse(tmpl)

	if err != nil {
		return fmt.Sprintf("<html><body><h1>Error generating report: %v</h1></body></html>", err)
	}

	// Prepare template data
	data := struct {
		*models.AnalysisResult
		PackagesByName        []*models.Package
		TopUncoveredFunctions []*models.Function
	}{
		AnalysisResult:        result,
		PackagesByName:        getPackagesSortedByName(result),
		TopUncoveredFunctions: getTopUncoveredFunctions(result, 15),
	}

	var buf strings.Builder
	if err := t.Execute(&buf, data); err != nil {
		return fmt.Sprintf("<html><body><h1>Error executing template: %v</h1></body></html>", err)
	}

	return buf.String()
}

// generateXMLReport creates an XML report in JUnit format for CI/CD integration
func generateXMLReport(result *models.AnalysisResult, opts *Options) error {
	type TestCase struct {
		XMLName   xml.Name `xml:"testcase"`
		ClassName string   `xml:"classname,attr"`
		Name      string   `xml:"name,attr"`
		Time      string   `xml:"time,attr"`
		Failure   *struct {
			Message string `xml:"message,attr"`
			Text    string `xml:",chardata"`
		} `xml:"failure,omitempty"`
	}

	type TestSuite struct {
		XMLName   xml.Name   `xml:"testsuite"`
		Name      string     `xml:"name,attr"`
		Tests     int        `xml:"tests,attr"`
		Failures  int        `xml:"failures,attr"`
		Time      string     `xml:"time,attr"`
		TestCases []TestCase `xml:"testcase"`
	}

	var testCases []TestCase
	failures := 0

	// Create test cases for each package
	for _, pkg := range result.PackageCoverage {
		testCase := TestCase{
			ClassName: pkg.Name,
			Name:      "coverage",
			Time:      "0.0",
		}

		if pkg.Coverage < opts.Threshold {
			failures++
			testCase.Failure = &struct {
				Message string `xml:"message,attr"`
				Text    string `xml:",chardata"`
			}{
				Message: fmt.Sprintf("Coverage %.1f%% below threshold %.1f%%", pkg.Coverage, opts.Threshold),
				Text:    fmt.Sprintf("Package %s has coverage %.1f%% which is below the required threshold of %.1f%%", pkg.Name, pkg.Coverage, opts.Threshold),
			}
		}

		testCases = append(testCases, testCase)
	}

	testSuite := TestSuite{
		Name:      "Coverage Report",
		Tests:     len(testCases),
		Failures:  failures,
		Time:      fmt.Sprintf("%.2f", result.Metadata.AnalysisTime.Seconds()),
		TestCases: testCases,
	}

	xmlData, err := xml.MarshalIndent(testSuite, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal XML: %w", err)
	}

	xmlContent := xml.Header + string(xmlData)
	return writeOutput(xmlContent, opts.OutputFile)
}

// Helper functions

func getCoverageColor(coverage, threshold float64) string {
	if coverage >= threshold {
		return ColorGreen
	} else if coverage >= threshold*0.75 {
		return ColorYellow
	}
	return ColorRed
}

func truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length-3] + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func writeOutput(content, filename string) error {
	if filename == "" {
		fmt.Print(content)
		return nil
	}

	return os.WriteFile(filename, []byte(content), 0644)
}

func openInBrowser(filename string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", filename}
	case "darwin":
		cmd = "open"
		args = []string{filename}
	case "linux":
		cmd = "xdg-open"
		args = []string{filename}
	default:
		return fmt.Errorf("unsupported platform")
	}

	return exec.Command(cmd, args...).Start()
}

func getPackagesSortedByName(result *models.AnalysisResult) []*models.Package {
	packages := make([]*models.Package, 0, len(result.PackageCoverage))
	for _, pkg := range result.PackageCoverage {
		packages = append(packages, pkg)
	}

	sort.Slice(packages, func(i, j int) bool {
		return packages[i].Name < packages[j].Name
	})

	return packages
}

func getTopUncoveredFunctions(result *models.AnalysisResult, limit int) []*models.Function {
	if len(result.UncoveredFunctions) <= limit {
		return result.UncoveredFunctions
	}
	return result.UncoveredFunctions[:limit]
}
