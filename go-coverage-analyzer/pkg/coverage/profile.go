package coverage

import (
	"bufio"
	"fmt"
	"go/build"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/beck/go-coverage-analyzer/pkg/models"
)

// ProfileParser handles parsing and generation of Go coverage profiles
type ProfileParser struct {
	verbose bool
}

// NewProfileParser creates a new profile parser
func NewProfileParser(verbose bool) *ProfileParser {
	return &ProfileParser{
		verbose: verbose,
	}
}

// GenerateProfile runs go test with coverage and generates a coverage profile
func (p *ProfileParser) GenerateProfile(projectPath, outputFile string, packagePattern string) error {
	if p.verbose {
		fmt.Printf("🔍 Generating coverage profile for: %s\n", projectPath)
	}

	// Prepare the go test command
	args := []string{"test", "-coverprofile=" + outputFile}

	// Add coverage mode
	args = append(args, "-covermode=atomic")

	// Add package pattern or default to all packages
	if packagePattern != "" {
		args = append(args, packagePattern)
	} else {
		args = append(args, "./...")
	}

	// Ensure output file is absolute path
	if !filepath.IsAbs(outputFile) {
		outputFile = filepath.Join(projectPath, outputFile)
	}

	// Update the coverprofile argument with absolute path
	for i, arg := range args {
		if strings.HasPrefix(arg, "-coverprofile=") {
			args[i] = "-coverprofile=" + outputFile
			break
		}
	}

	// Execute go test command
	cmd := exec.Command("go", args...)
	cmd.Dir = projectPath

	if p.verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		fmt.Printf("Running: go %s (in %s)\n", strings.Join(args, " "), projectPath)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to generate coverage profile: %w", err)
	}

	// Verify profile was created
	if _, err := os.Stat(outputFile); err != nil {
		return fmt.Errorf("coverage profile not found at %s: %w", outputFile, err)
	}

	if p.verbose {
		fmt.Printf("✅ Coverage profile generated: %s\n", outputFile)
	}

	return nil
}

// ParseProfile parses a Go coverage profile file
func (p *ProfileParser) ParseProfile(profilePath string) (*models.CoverageProfile, error) {
	file, err := os.Open(profilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open profile file %s: %w", profilePath, err)
	}
	defer file.Close()

	profile := &models.CoverageProfile{
		Blocks: make([]*models.ProfileBlock, 0),
		Files:  make(map[string]*models.FileProfile),
	}

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineNum++

		if lineNum == 1 {
			// First line contains the coverage mode
			if strings.HasPrefix(line, "mode: ") {
				profile.Mode = strings.TrimPrefix(line, "mode: ")
				continue
			}
		}

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		block, err := p.parseProfileLine(line)
		if err != nil {
			return nil, fmt.Errorf("failed to parse line %d: %w", lineNum, err)
		}

		profile.Blocks = append(profile.Blocks, block)

		// Group blocks by file
		if _, exists := profile.Files[block.FileName]; !exists {
			profile.Files[block.FileName] = &models.FileProfile{
				FileName: block.FileName,
				Blocks:   make([]*models.ProfileBlock, 0),
			}
		}
		profile.Files[block.FileName].Blocks = append(profile.Files[block.FileName].Blocks, block)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading profile file: %w", err)
	}

	// Calculate coverage statistics for each file
	for _, fileProfile := range profile.Files {
		p.calculateFileCoverage(fileProfile)
	}

	if p.verbose {
		fmt.Printf("📊 Parsed profile: %d blocks across %d files\n", len(profile.Blocks), len(profile.Files))
	}

	return profile, nil
}

// parseProfileLine parses a single line from a coverage profile
// Format: filename:startLine.startCol,endLine.endCol numStmts count
func (p *ProfileParser) parseProfileLine(line string) (*models.ProfileBlock, error) {
	parts := strings.Fields(line)
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid profile line format: %s", line)
	}

	// Parse file and position information
	filePos := parts[0]
	colonIndex := strings.LastIndex(filePos, ":")
	if colonIndex == -1 {
		return nil, fmt.Errorf("invalid file:position format: %s", filePos)
	}

	fileName := filePos[:colonIndex]
	positions := filePos[colonIndex+1:]

	// Parse positions (startLine.startCol,endLine.endCol)
	commaIndex := strings.Index(positions, ",")
	if commaIndex == -1 {
		return nil, fmt.Errorf("invalid position format: %s", positions)
	}

	startPos := positions[:commaIndex]
	endPos := positions[commaIndex+1:]

	startLine, startCol, err := p.parsePosition(startPos)
	if err != nil {
		return nil, fmt.Errorf("invalid start position: %w", err)
	}

	endLine, endCol, err := p.parsePosition(endPos)
	if err != nil {
		return nil, fmt.Errorf("invalid end position: %w", err)
	}

	// Parse statement count
	numStmts, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid statement count: %w", err)
	}

	// Parse execution count
	count, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid execution count: %w", err)
	}

	return &models.ProfileBlock{
		FileName:  fileName,
		StartLine: startLine,
		StartCol:  startCol,
		EndLine:   endLine,
		EndCol:    endCol,
		NumStmts:  numStmts,
		Count:     count,
	}, nil
}

// parsePosition parses a position in the format "line.column"
func (p *ProfileParser) parsePosition(pos string) (line, col int, err error) {
	dotIndex := strings.Index(pos, ".")
	if dotIndex == -1 {
		return 0, 0, fmt.Errorf("invalid position format: %s", pos)
	}

	line, err = strconv.Atoi(pos[:dotIndex])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid line number: %w", err)
	}

	col, err = strconv.Atoi(pos[dotIndex+1:])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid column number: %w", err)
	}

	return line, col, nil
}

// calculateFileCoverage calculates coverage statistics for a file
func (p *ProfileParser) calculateFileCoverage(fileProfile *models.FileProfile) {
	totalStmts := 0
	coveredStmts := 0

	for _, block := range fileProfile.Blocks {
		totalStmts += block.NumStmts
		if block.Count > 0 {
			coveredStmts += block.NumStmts
		}
	}

	fileProfile.TotalStmts = totalStmts
	fileProfile.CoveredStmts = coveredStmts

	if totalStmts > 0 {
		fileProfile.Coverage = float64(coveredStmts) / float64(totalStmts) * 100.0
	}
}

// ConvertToBlocks converts profile blocks to internal Block format
func (p *ProfileParser) ConvertToBlocks(profileBlocks []*models.ProfileBlock) []*models.Block {
	blocks := make([]*models.Block, len(profileBlocks))

	for i, pb := range profileBlocks {
		blocks[i] = &models.Block{
			StartLine: pb.StartLine,
			StartCol:  pb.StartCol,
			EndLine:   pb.EndLine,
			EndCol:    pb.EndCol,
			NumStmts:  pb.NumStmts,
			Count:     pb.Count,
			IsCovered: pb.Count > 0,
		}
	}

	return blocks
}

// GetProjectInfo extracts information about the Go project
func (p *ProfileParser) GetProjectInfo(projectPath string) (*models.ProjectInfo, error) {
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	info := &models.ProjectInfo{
		RootDir:      absPath,
		Packages:     make([]string, 0),
		TestFiles:    make([]string, 0),
		Dependencies: make([]string, 0),
	}

	// Try to get module information
	if modPath, goVersion := p.getModuleInfo(absPath); modPath != "" {
		info.ModulePath = modPath
		info.GoVersion = goVersion
	}

	// Walk the project directory to collect information
	err = filepath.Walk(absPath, func(path string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fileInfo.IsDir() {
			// Skip vendor, .git, and other common directories
			name := fileInfo.Name()
			if name == "vendor" || name == ".git" || name == "node_modules" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}

		if strings.HasSuffix(path, ".go") {
			relPath, _ := filepath.Rel(absPath, path)

			if strings.HasSuffix(path, "_test.go") {
				info.TestFiles = append(info.TestFiles, relPath)
				info.HasTests = true
			} else {
				info.TotalFiles++
			}

			// Count lines and functions (simplified)
			if lines, functions := p.countLinesAndFunctions(path); lines > 0 {
				info.TotalLines += lines
				info.TotalFunctions += functions
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk project directory: %w", err)
	}

	return info, nil
}

// getModuleInfo extracts Go module information
func (p *ProfileParser) getModuleInfo(projectPath string) (string, string) {
	// Try to read go.mod file
	modFile := filepath.Join(projectPath, "go.mod")
	if content, err := os.ReadFile(modFile); err == nil {
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "module ") {
				module := strings.TrimSpace(strings.TrimPrefix(line, "module"))

				// Try to get Go version
				for _, l := range lines {
					if strings.HasPrefix(strings.TrimSpace(l), "go ") {
						version := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(l), "go"))
						return module, version
					}
				}
				return module, ""
			}
		}
	}

	// Fallback: try to get from build context
	ctx := build.Default
	if pkg, err := ctx.ImportDir(projectPath, 0); err == nil {
		return pkg.ImportPath, ""
	}

	return "", ""
}

// countLinesAndFunctions provides a simple count of lines and functions in a Go file
func (p *ProfileParser) countLinesAndFunctions(filePath string) (int, int) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return 0, 0
	}

	lines := strings.Split(string(content), "\n")
	lineCount := len(lines)
	functionCount := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "func ") {
			functionCount++
		}
	}

	return lineCount, functionCount
}

// ValidateProfile checks if a coverage profile is valid
func (p *ProfileParser) ValidateProfile(profile *models.CoverageProfile) error {
	if profile == nil {
		return fmt.Errorf("profile is nil")
	}

	if profile.Mode == "" {
		return fmt.Errorf("profile mode is empty")
	}

	validModes := map[string]bool{
		"set":    true,
		"count":  true,
		"atomic": true,
	}

	if !validModes[profile.Mode] {
		return fmt.Errorf("invalid coverage mode: %s", profile.Mode)
	}

	if len(profile.Blocks) == 0 {
		return fmt.Errorf("profile contains no coverage blocks")
	}

	// Validate individual blocks
	for i, block := range profile.Blocks {
		if err := p.validateBlock(block); err != nil {
			return fmt.Errorf("invalid block at index %d: %w", i, err)
		}
	}

	return nil
}

// validateBlock validates a single coverage block
func (p *ProfileParser) validateBlock(block *models.ProfileBlock) error {
	if block.FileName == "" {
		return fmt.Errorf("block filename is empty")
	}

	if block.StartLine <= 0 || block.EndLine <= 0 {
		return fmt.Errorf("invalid line numbers: start=%d, end=%d", block.StartLine, block.EndLine)
	}

	if block.StartLine > block.EndLine {
		return fmt.Errorf("start line %d is greater than end line %d", block.StartLine, block.EndLine)
	}

	if block.StartLine == block.EndLine && block.StartCol > block.EndCol {
		return fmt.Errorf("start column %d is greater than end column %d on same line", block.StartCol, block.EndCol)
	}

	if block.NumStmts <= 0 {
		return fmt.Errorf("invalid statement count: %d", block.NumStmts)
	}

	if block.Count < 0 {
		return fmt.Errorf("invalid execution count: %d", block.Count)
	}

	return nil
}
