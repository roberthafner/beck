package generator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/beck/go-coverage-analyzer/pkg/models"
)

// MockGenerator handles the generation of mock interfaces for testing
type MockGenerator struct {
	fileSet *token.FileSet
	verbose bool
}

// NewMockGenerator creates a new mock generator
func NewMockGenerator(verbose bool) *MockGenerator {
	return &MockGenerator{
		fileSet: token.NewFileSet(),
		verbose: verbose,
	}
}

// MockInterface represents an interface that needs mocking
type MockInterface struct {
	Name       string
	Package    string
	Methods    []*MockMethod
	ImportPath string
	FilePath   string
}

// MockMethod represents a method in an interface
type MockMethod struct {
	Name       string
	Parameters []*MockParam
	Returns    []*MockReturn
	Signature  string
	CallArgs   string
	ReturnCall string
}

// MockParam represents a parameter in a mock method
type MockParam struct {
	Name string
	Type string
}

// MockReturn represents a return value in a mock method
type MockReturn struct {
	Name string
	Type string
}

// GeneratedMock represents a generated mock file
type GeneratedMock struct {
	Interface    *MockInterface
	FilePath     string
	Content      string
	TestFilePath string
}

// GenerateMocks generates mock implementations for interfaces used by functions
func (mg *MockGenerator) GenerateMocks(functions []*models.Function, projectPath string) ([]*GeneratedMock, error) {
	if mg.verbose {
		fmt.Println("🎭 Starting mock generation...")
	}

	// Find all interfaces that need mocking
	interfaces, err := mg.findInterfacesToMock(functions, projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to find interfaces to mock: %w", err)
	}

	if len(interfaces) == 0 {
		if mg.verbose {
			fmt.Println("🎭 No interfaces found that need mocking")
		}
		return []*GeneratedMock{}, nil
	}

	var generatedMocks []*GeneratedMock

	for _, iface := range interfaces {
		if mg.verbose {
			fmt.Printf("🎭 Generating mock for interface: %s\n", iface.Name)
		}

		mock, err := mg.generateMockFile(iface, projectPath)
		if err != nil {
			if mg.verbose {
				fmt.Printf("⚠️ Failed to generate mock for %s: %v\n", iface.Name, err)
			}
			continue
		}

		generatedMocks = append(generatedMocks, mock)
	}

	if mg.verbose {
		fmt.Printf("🎭 Generated %d mocks\n", len(generatedMocks))
	}

	return generatedMocks, nil
}

// findInterfacesToMock identifies interfaces used by functions that need mocking
func (mg *MockGenerator) findInterfacesToMock(functions []*models.Function, projectPath string) ([]*MockInterface, error) {
	interfaceMap := make(map[string]*MockInterface)

	for _, function := range functions {
		// Check function parameters for interface types
		for _, param := range function.Parameters {
			if mg.isInterfaceType(param.Type) {
				iface, err := mg.parseInterface(param.Type, function.File, projectPath)
				if err != nil {
					if mg.verbose {
						fmt.Printf("⚠️ Could not parse interface %s: %v\n", param.Type, err)
					}
					continue
				}
				if iface != nil {
					interfaceMap[iface.Name] = iface
				}
			}
		}

		// Check for interfaces in function dependencies
		for _, dep := range function.Dependencies {
			if mg.isInterfaceType(dep) {
				iface, err := mg.parseInterface(dep, function.File, projectPath)
				if err != nil {
					continue
				}
				if iface != nil {
					interfaceMap[iface.Name] = iface
				}
			}
		}
	}

	var interfaces []*MockInterface
	for _, iface := range interfaceMap {
		interfaces = append(interfaces, iface)
	}

	return interfaces, nil
}

// isInterfaceType determines if a type is likely an interface
func (mg *MockGenerator) isInterfaceType(typeName string) bool {
	// Skip basic types
	basicTypes := []string{"string", "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64", "float32", "float64",
		"bool", "byte", "rune", "error"}

	for _, basic := range basicTypes {
		if typeName == basic || strings.HasPrefix(typeName, "[]"+basic) ||
			strings.HasPrefix(typeName, "map[") {
			return false
		}
	}

	// Skip pointers to basic types
	if strings.HasPrefix(typeName, "*") {
		return mg.isInterfaceType(strings.TrimPrefix(typeName, "*"))
	}

	// Skip slices and maps of basic types
	if strings.HasPrefix(typeName, "[]") || strings.HasPrefix(typeName, "map[") {
		return false
	}

	// Likely an interface if it's a custom type
	return strings.Contains(typeName, ".") ||
		(len(typeName) > 0 && strings.ToUpper(typeName[:1]) == typeName[:1])
}

// parseInterface parses an interface from source code
func (mg *MockGenerator) parseInterface(interfaceType, sourceFile, projectPath string) (*MockInterface, error) {
	// Find the source file containing the interface
	interfaceFile, err := mg.findInterfaceFile(interfaceType, sourceFile, projectPath)
	if err != nil {
		return nil, err
	}

	// Parse the source file
	src, err := os.ReadFile(interfaceFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", interfaceFile, err)
	}

	file, err := parser.ParseFile(mg.fileSet, interfaceFile, src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %s: %w", interfaceFile, err)
	}

	// Extract interface definition
	var mockInterface *MockInterface

	ast.Inspect(file, func(n ast.Node) bool {
		if typeSpec, ok := n.(*ast.TypeSpec); ok {
			if interfaceTypeName := mg.extractInterfaceName(interfaceType); typeSpec.Name.Name == interfaceTypeName {
				if interfaceTypeNode, ok := typeSpec.Type.(*ast.InterfaceType); ok {
					mockInterface = &MockInterface{
						Name:       typeSpec.Name.Name,
						Package:    file.Name.Name,
						FilePath:   interfaceFile,
						ImportPath: mg.getImportPath(interfaceFile, projectPath),
						Methods:    mg.parseInterfaceMethods(interfaceTypeNode),
					}
					return false
				}
			}
		}
		return true
	})

	return mockInterface, nil
}

// findInterfaceFile finds the file containing an interface definition
func (mg *MockGenerator) findInterfaceFile(interfaceType, sourceFile, projectPath string) (string, error) {
	// If interface is in same package, check the same directory
	if !strings.Contains(interfaceType, ".") {
		sourceDir := filepath.Dir(filepath.Join(projectPath, sourceFile))
		files, err := filepath.Glob(filepath.Join(sourceDir, "*.go"))
		if err != nil {
			return "", err
		}

		for _, file := range files {
			if strings.HasSuffix(file, "_test.go") {
				continue
			}
			if mg.containsInterface(file, interfaceType) {
				return file, nil
			}
		}
	}

	// For external interfaces, we'd need import resolution
	// For now, return the source file as fallback
	return filepath.Join(projectPath, sourceFile), nil
}

// containsInterface checks if a file contains an interface definition
func (mg *MockGenerator) containsInterface(filePath, interfaceName string) bool {
	src, err := os.ReadFile(filePath)
	if err != nil {
		return false
	}

	file, err := parser.ParseFile(mg.fileSet, filePath, src, 0)
	if err != nil {
		return false
	}

	found := false
	ast.Inspect(file, func(n ast.Node) bool {
		if typeSpec, ok := n.(*ast.TypeSpec); ok {
			if typeSpec.Name.Name == interfaceName {
				if _, ok := typeSpec.Type.(*ast.InterfaceType); ok {
					found = true
					return false
				}
			}
		}
		return true
	})

	return found
}

// parseInterfaceMethods extracts methods from an interface AST node
func (mg *MockGenerator) parseInterfaceMethods(interfaceType *ast.InterfaceType) []*MockMethod {
	var methods []*MockMethod

	for _, method := range interfaceType.Methods.List {
		if funcType, ok := method.Type.(*ast.FuncType); ok {
			for _, name := range method.Names {
				mockMethod := &MockMethod{
					Name:       name.Name,
					Parameters: mg.parseMethodParams(funcType.Params),
					Returns:    mg.parseMethodReturns(funcType.Results),
				}
				mockMethod.Signature = mg.buildMethodSignature(mockMethod)
				mockMethod.CallArgs = mg.buildCallArgs(mockMethod.Parameters)
				mockMethod.ReturnCall = mg.buildReturnCall(mockMethod.Returns)
				methods = append(methods, mockMethod)
			}
		}
	}

	return methods
}

// parseMethodParams extracts parameters from a function type
func (mg *MockGenerator) parseMethodParams(params *ast.FieldList) []*MockParam {
	if params == nil {
		return []*MockParam{}
	}

	var mockParams []*MockParam
	paramIndex := 0

	for _, field := range params.List {
		typeStr := mg.typeToString(field.Type)

		if len(field.Names) == 0 {
			// Unnamed parameter
			mockParams = append(mockParams, &MockParam{
				Name: fmt.Sprintf("arg%d", paramIndex),
				Type: typeStr,
			})
			paramIndex++
		} else {
			// Named parameters
			for _, name := range field.Names {
				mockParams = append(mockParams, &MockParam{
					Name: name.Name,
					Type: typeStr,
				})
				paramIndex++
			}
		}
	}

	return mockParams
}

// parseMethodReturns extracts return types from a function type
func (mg *MockGenerator) parseMethodReturns(returns *ast.FieldList) []*MockReturn {
	if returns == nil {
		return []*MockReturn{}
	}

	var mockReturns []*MockReturn
	returnIndex := 0

	for _, field := range returns.List {
		typeStr := mg.typeToString(field.Type)

		if len(field.Names) == 0 {
			// Unnamed return
			mockReturns = append(mockReturns, &MockReturn{
				Name: fmt.Sprintf("ret%d", returnIndex),
				Type: typeStr,
			})
			returnIndex++
		} else {
			// Named returns
			for _, name := range field.Names {
				mockReturns = append(mockReturns, &MockReturn{
					Name: name.Name,
					Type: typeStr,
				})
				returnIndex++
			}
		}
	}

	return mockReturns
}

// typeToString converts an AST type to its string representation
func (mg *MockGenerator) typeToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + mg.typeToString(t.X)
	case *ast.ArrayType:
		return "[]" + mg.typeToString(t.Elt)
	case *ast.MapType:
		return "map[" + mg.typeToString(t.Key) + "]" + mg.typeToString(t.Value)
	case *ast.SelectorExpr:
		return mg.typeToString(t.X) + "." + t.Sel.Name
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.FuncType:
		return "func" // Simplified
	default:
		return "interface{}" // Fallback
	}
}

// buildMethodSignature creates the method signature string
func (mg *MockGenerator) buildMethodSignature(method *MockMethod) string {
	var params []string
	for _, param := range method.Parameters {
		params = append(params, fmt.Sprintf("%s %s", param.Name, param.Type))
	}

	signature := fmt.Sprintf("(%s)", strings.Join(params, ", "))

	if len(method.Returns) > 0 {
		var returns []string
		for _, ret := range method.Returns {
			returns = append(returns, ret.Type)
		}

		if len(returns) == 1 {
			signature += " " + returns[0]
		} else {
			signature += " (" + strings.Join(returns, ", ") + ")"
		}
	}

	return signature
}

// buildCallArgs creates the arguments string for mock calls
func (mg *MockGenerator) buildCallArgs(params []*MockParam) string {
	var args []string
	for _, param := range params {
		args = append(args, param.Name)
	}
	return strings.Join(args, ", ")
}

// buildReturnCall creates the return call string for testify mocks
func (mg *MockGenerator) buildReturnCall(returns []*MockReturn) string {
	if len(returns) == 0 {
		return ""
	}

	if len(returns) == 1 {
		switch returns[0].Type {
		case "string":
			return "args.String(0)"
		case "int", "int32", "int64":
			return "args.Int(0)"
		case "bool":
			return "args.Bool(0)"
		case "error":
			return "args.Error(0)"
		case "interface{}":
			return "args.Get(0)"
		default:
			return "args.Get(0).(" + returns[0].Type + ")"
		}
	}

	// Multiple returns
	var calls []string
	for i, ret := range returns {
		switch ret.Type {
		case "string":
			calls = append(calls, fmt.Sprintf("args.String(%d)", i))
		case "int", "int32", "int64":
			calls = append(calls, fmt.Sprintf("args.Int(%d)", i))
		case "bool":
			calls = append(calls, fmt.Sprintf("args.Bool(%d)", i))
		case "error":
			calls = append(calls, fmt.Sprintf("args.Error(%d)", i))
		default:
			calls = append(calls, fmt.Sprintf("args.Get(%d).(%s)", i, ret.Type))
		}
	}

	return strings.Join(calls, ", ")
}

// generateMockFile generates the complete mock file content
func (mg *MockGenerator) generateMockFile(iface *MockInterface, projectPath string) (*GeneratedMock, error) {
	// Collect all necessary imports
	imports := []string{"github.com/stretchr/testify/mock"}

	// Check if we need context import
	needsContext := false
	needsIO := false

	for _, method := range iface.Methods {
		for _, param := range method.Parameters {
			if strings.Contains(param.Type, "context.Context") {
				needsContext = true
			}
			if strings.Contains(param.Type, "io.") {
				needsIO = true
			}
		}
		for _, ret := range method.Returns {
			if strings.Contains(ret.Type, "context.Context") {
				needsContext = true
			}
			if strings.Contains(ret.Type, "io.") {
				needsIO = true
			}
		}
	}

	if needsContext {
		imports = append(imports, "context")
	}
	if needsIO {
		imports = append(imports, "io")
	}

	// Define the template
	mockTemplate := `package {{.Package}}

import (
{{range .Imports}}	"{{.}}"
{{end}}
)

// Mock{{.Name}} is a mock implementation of {{.Name}}
type Mock{{.Name}} struct {
	mock.Mock
}

{{range .Methods}}// {{.Name}} mocks the {{.Name}} method
func (m *Mock{{$.Name}}) {{.Name}}{{.Signature}} {
	{{if .Returns}}{{if .Parameters}}args := m.Called({{.CallArgs}}){{else}}args := m.Called(){{end}}
	return {{.ReturnCall}}{{else}}{{if .Parameters}}m.Called({{.CallArgs}}){{else}}m.Called(){{end}}{{end}}
}

{{end}}`

	// Parse and execute template
	tmpl, err := template.New("mock").Parse(mockTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse mock template: %w", err)
	}

	// Create template data with imports
	templateData := struct {
		*MockInterface
		Imports []string
	}{
		MockInterface: iface,
		Imports:       imports,
	}

	var content strings.Builder
	if err := tmpl.Execute(&content, templateData); err != nil {
		return nil, fmt.Errorf("failed to execute mock template: %w", err)
	}

	// Determine output file path
	mockFileName := fmt.Sprintf("mock_%s.go", strings.ToLower(iface.Name))
	mockFilePath := filepath.Join(filepath.Dir(iface.FilePath), mockFileName)

	// Make path relative to project
	if strings.HasPrefix(mockFilePath, projectPath) {
		mockFilePath = strings.TrimPrefix(mockFilePath, projectPath)
		mockFilePath = strings.TrimPrefix(mockFilePath, "/")
	}

	return &GeneratedMock{
		Interface:    iface,
		FilePath:     mockFilePath,
		Content:      content.String(),
		TestFilePath: iface.FilePath,
	}, nil
}

// extractInterfaceName extracts interface name from a type string
func (mg *MockGenerator) extractInterfaceName(interfaceType string) string {
	// Handle qualified names like "package.Interface"
	parts := strings.Split(interfaceType, ".")
	return parts[len(parts)-1]
}

// getImportPath determines the import path for an interface file
func (mg *MockGenerator) getImportPath(filePath, projectPath string) string {
	relativePath := strings.TrimPrefix(filePath, projectPath)
	relativePath = strings.TrimPrefix(relativePath, "/")
	return filepath.Dir(relativePath)
}

// WriteMocks writes generated mocks to files
func (mg *MockGenerator) WriteMocks(mocks []*GeneratedMock, projectPath string, dryRun bool) error {
	if dryRun {
		if mg.verbose {
			fmt.Printf("🎭 [DRY RUN] Would write %d mock files\n", len(mocks))
		}
		return nil
	}

	for _, mock := range mocks {
		fullPath := filepath.Join(projectPath, mock.FilePath)

		// Create directory if needed
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}

		// Write mock file
		if err := os.WriteFile(fullPath, []byte(mock.Content), 0644); err != nil {
			return fmt.Errorf("failed to write mock file %s: %w", fullPath, err)
		}

		if mg.verbose {
			fmt.Printf("🎭 Generated mock file: %s\n", mock.FilePath)
		}
	}

	return nil
}
