package generator

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/beck/go-coverage-analyzer/pkg/models"
)

// DataGenerator handles intelligent test data generation
type DataGenerator struct {
	rand    *rand.Rand
	verbose bool
}

// NewDataGenerator creates a new data generator
func NewDataGenerator(verbose bool) *DataGenerator {
	return &DataGenerator{
		rand:    rand.New(rand.NewSource(time.Now().UnixNano())),
		verbose: verbose,
	}
}

// GenerationStrategy defines how to generate test data
type GenerationStrategy string

const (
	StrategyPositive GenerationStrategy = "positive"
	StrategyNegative GenerationStrategy = "negative"
	StrategyEdge     GenerationStrategy = "edge"
	StrategyRandom   GenerationStrategy = "random"
	StrategyZero     GenerationStrategy = "zero"
)

// TestDataSet represents a complete set of test data for a function
type TestDataSet struct {
	FunctionName string
	TestCases    []GeneratedTestCase
}

// GeneratedTestCase represents a generated test case with inputs and expected outputs
type GeneratedTestCase struct {
	Name           string
	Description    string
	Strategy       GenerationStrategy
	Inputs         map[string]interface{}
	InputStrings   map[string]string // String representation for code generation
	ExpectedResult interface{}
	ExpectError    bool
	ErrorPattern   string
	Tags           []string
}

// GenerateTestData generates comprehensive test data for a function
func (dg *DataGenerator) GenerateTestData(function *models.Function, maxCases int) (*TestDataSet, error) {
	if dg.verbose {
		fmt.Printf("🎲 Generating test data for function: %s\n", function.Name)
	}

	testSet := &TestDataSet{
		FunctionName: function.Name,
		TestCases:    make([]GeneratedTestCase, 0),
	}

	// Generate different types of test cases
	strategies := []GenerationStrategy{StrategyPositive, StrategyEdge, StrategyZero}
	if function.HasErrorReturn {
		strategies = append(strategies, StrategyNegative)
	}

	casesPerStrategy := maxCases / len(strategies)
	if casesPerStrategy < 1 {
		casesPerStrategy = 1
	}

	for _, strategy := range strategies {
		cases := dg.generateCasesForStrategy(function, strategy, casesPerStrategy)
		testSet.TestCases = append(testSet.TestCases, cases...)

		if len(testSet.TestCases) >= maxCases {
			break
		}
	}

	// Trim to max cases
	if len(testSet.TestCases) > maxCases {
		testSet.TestCases = testSet.TestCases[:maxCases]
	}

	if dg.verbose {
		fmt.Printf("✅ Generated %d test cases for %s\n", len(testSet.TestCases), function.Name)
	}

	return testSet, nil
}

// generateCasesForStrategy generates test cases for a specific strategy
func (dg *DataGenerator) generateCasesForStrategy(function *models.Function, strategy GenerationStrategy, count int) []GeneratedTestCase {
	var cases []GeneratedTestCase

	for i := 0; i < count; i++ {
		testCase := GeneratedTestCase{
			Name:         fmt.Sprintf("%s_case_%d", strategy, i+1),
			Description:  dg.getStrategyDescription(strategy, function.Name),
			Strategy:     strategy,
			Inputs:       make(map[string]interface{}),
			InputStrings: make(map[string]string),
			Tags:         []string{string(strategy)},
		}

		// Generate inputs for each parameter
		for _, param := range function.Parameters {
			value, stringRepr := dg.generateValueForType(param.Type, strategy)
			testCase.Inputs[param.Name] = value
			testCase.InputStrings[param.Name] = stringRepr
		}

		// Generate expected result based on function analysis
		if len(function.ReturnTypes) > 0 {
			testCase.ExpectedResult = dg.generateExpectedResult(function, testCase.Inputs, strategy)
		}

		// Set error expectations
		testCase.ExpectError = dg.shouldExpectError(function, strategy, testCase.Inputs)
		if testCase.ExpectError {
			testCase.ErrorPattern = dg.generateErrorPattern(function, strategy)
		}

		cases = append(cases, testCase)
	}

	return cases
}

// generateValueForType generates a value for a specific Go type
func (dg *DataGenerator) generateValueForType(goType string, strategy GenerationStrategy) (interface{}, string) {
	switch {
	case goType == "string":
		return dg.generateStringValue(strategy)
	case isIntegerType(goType):
		return dg.generateIntValue(strategy, goType)
	case isFloatType(goType):
		return dg.generateFloatValue(strategy, goType)
	case goType == "bool":
		return dg.generateBoolValue(strategy)
	case strings.HasPrefix(goType, "[]"):
		return dg.generateSliceValue(goType, strategy)
	case strings.HasPrefix(goType, "map["):
		return dg.generateMapValue(goType, strategy)
	case strings.HasPrefix(goType, "*"):
		return dg.generatePointerValue(goType, strategy)
	case goType == "interface{}" || goType == "any":
		return dg.generateInterfaceValue(strategy)
	case strings.Contains(goType, "."):
		return dg.generateCustomTypeValue(goType, strategy)
	default:
		return dg.generateStructValue(goType, strategy)
	}
}

// generateStringValue generates string test values
func (dg *DataGenerator) generateStringValue(strategy GenerationStrategy) (interface{}, string) {
	switch strategy {
	case StrategyPositive:
		values := []string{"test", "hello", "example", "valid_input", "sample_data"}
		value := values[dg.rand.Intn(len(values))]
		return value, fmt.Sprintf(`"%s"`, value)
	case StrategyNegative:
		// Generate potentially problematic strings
		values := []string{"", "null", "undefined", "<script>", "'; DROP TABLE;", "../../etc/passwd"}
		value := values[dg.rand.Intn(len(values))]
		return value, fmt.Sprintf(`"%s"`, value)
	case StrategyEdge:
		values := []string{"", " ", "\n", "\t", string(rune(0)), strings.Repeat("a", 1000)}
		value := values[dg.rand.Intn(len(values))]
		return value, fmt.Sprintf(`"%s"`, value)
	case StrategyZero:
		return "", `""`
	case StrategyRandom:
		length := dg.rand.Intn(50) + 1
		value := dg.generateRandomString(length)
		return value, fmt.Sprintf(`"%s"`, value)
	default:
		return "test", `"test"`
	}
}

// generateIntValue generates integer test values
func (dg *DataGenerator) generateIntValue(strategy GenerationStrategy, goType string) (interface{}, string) {
	switch strategy {
	case StrategyPositive:
		value := dg.rand.Intn(1000) + 1
		return value, strconv.Itoa(value)
	case StrategyNegative:
		value := -(dg.rand.Intn(1000) + 1)
		return value, strconv.Itoa(value)
	case StrategyEdge:
		edges := dg.getIntegerEdgeValues(goType)
		value := edges[dg.rand.Intn(len(edges))]
		return value, strconv.FormatInt(value, 10)
	case StrategyZero:
		return 0, "0"
	case StrategyRandom:
		value := dg.rand.Intn(2000000) - 1000000
		return value, strconv.Itoa(value)
	default:
		return 42, "42"
	}
}

// generateFloatValue generates float test values
func (dg *DataGenerator) generateFloatValue(strategy GenerationStrategy, goType string) (interface{}, string) {
	switch strategy {
	case StrategyPositive:
		value := dg.rand.Float64() * 100
		if goType == "float32" {
			value32 := float32(value)
			return value32, fmt.Sprintf("%g", value32)
		}
		return value, fmt.Sprintf("%g", value)
	case StrategyNegative:
		value := -(dg.rand.Float64() * 100)
		if goType == "float32" {
			value32 := float32(value)
			return value32, fmt.Sprintf("%g", value32)
		}
		return value, fmt.Sprintf("%g", value)
	case StrategyEdge:
		edges := []string{"0.0", "0.1", "-0.1", "1.0", "-1.0"}
		if goType == "float32" {
			edges = append(edges, "3.4028235e+38", "-3.4028235e+38", "1.175494e-38")
		} else {
			edges = append(edges, "1.7976931348623157e+308", "-1.7976931348623157e+308", "4.9406564584124654e-324")
		}
		edge := edges[dg.rand.Intn(len(edges))]
		if goType == "float32" {
			value, _ := strconv.ParseFloat(edge, 32)
			return float32(value), edge
		}
		value, _ := strconv.ParseFloat(edge, 64)
		return value, edge
	case StrategyZero:
		return 0.0, "0.0"
	case StrategyRandom:
		value := (dg.rand.Float64() - 0.5) * 2000000
		if goType == "float32" {
			value32 := float32(value)
			return value32, fmt.Sprintf("%g", value32)
		}
		return value, fmt.Sprintf("%g", value)
	default:
		return 3.14, "3.14"
	}
}

// generateBoolValue generates boolean test values
func (dg *DataGenerator) generateBoolValue(strategy GenerationStrategy) (interface{}, string) {
	switch strategy {
	case StrategyPositive:
		return true, "true"
	case StrategyNegative, StrategyEdge:
		return false, "false"
	case StrategyZero:
		return false, "false"
	case StrategyRandom:
		value := dg.rand.Intn(2) == 1
		return value, strconv.FormatBool(value)
	default:
		return true, "true"
	}
}

// generateSliceValue generates slice test values
func (dg *DataGenerator) generateSliceValue(goType string, strategy GenerationStrategy) (interface{}, string) {
	elementType := strings.TrimPrefix(goType, "[]")

	switch strategy {
	case StrategyPositive:
		length := dg.rand.Intn(5) + 1
		return dg.buildSlice(elementType, length, StrategyPositive)
	case StrategyNegative:
		length := dg.rand.Intn(3) + 1
		return dg.buildSlice(elementType, length, StrategyNegative)
	case StrategyEdge:
		// Empty slice or single element
		if dg.rand.Intn(2) == 0 {
			return dg.buildSlice(elementType, 0, strategy)
		}
		return dg.buildSlice(elementType, 1, StrategyEdge)
	case StrategyZero:
		return nil, "nil"
	case StrategyRandom:
		length := dg.rand.Intn(10)
		return dg.buildSlice(elementType, length, StrategyRandom)
	default:
		return dg.buildSlice(elementType, 2, StrategyPositive)
	}
}

// generateMapValue generates map test values
func (dg *DataGenerator) generateMapValue(goType string, strategy GenerationStrategy) (interface{}, string) {
	switch strategy {
	case StrategyZero:
		return nil, "nil"
	case StrategyEdge:
		return nil, "make(" + goType + ")"
	default:
		return nil, "make(" + goType + ")"
	}
}

// generatePointerValue generates pointer test values
func (dg *DataGenerator) generatePointerValue(goType string, strategy GenerationStrategy) (interface{}, string) {
	baseType := strings.TrimPrefix(goType, "*")

	switch strategy {
	case StrategyZero, StrategyEdge:
		return nil, "nil"
	default:
		_, valueStr := dg.generateValueForType(baseType, strategy)
		return nil, "&" + valueStr
	}
}

// generateInterfaceValue generates interface{} test values
func (dg *DataGenerator) generateInterfaceValue(strategy GenerationStrategy) (interface{}, string) {
	switch strategy {
	case StrategyPositive:
		return "test", `"test"`
	case StrategyNegative:
		return nil, "nil"
	case StrategyEdge:
		values := []string{"nil", `""`, "0", "false"}
		value := values[dg.rand.Intn(len(values))]
		return nil, value
	case StrategyZero:
		return nil, "nil"
	default:
		return 42, "42"
	}
}

// generateCustomTypeValue generates values for custom types
func (dg *DataGenerator) generateCustomTypeValue(goType string, strategy GenerationStrategy) (interface{}, string) {
	switch strategy {
	case StrategyZero:
		return nil, goType + "{}"
	default:
		return nil, goType + "{}"
	}
}

// generateStructValue generates struct test values
func (dg *DataGenerator) generateStructValue(goType string, strategy GenerationStrategy) (interface{}, string) {
	switch strategy {
	case StrategyZero:
		return nil, goType + "{}"
	default:
		return nil, goType + "{}"
	}
}

// buildSlice constructs a slice with generated elements
func (dg *DataGenerator) buildSlice(elementType string, length int, strategy GenerationStrategy) (interface{}, string) {
	if length == 0 {
		return nil, elementType + "{}"
	}

	var elements []string
	for i := 0; i < length; i++ {
		_, elementStr := dg.generateValueForType(elementType, strategy)
		elements = append(elements, elementStr)
	}

	sliceStr := fmt.Sprintf("[]%s{%s}", elementType, strings.Join(elements, ", "))
	return nil, sliceStr // Return nil for actual value, string for code generation
}

// generateExpectedResult generates expected results based on function analysis
func (dg *DataGenerator) generateExpectedResult(function *models.Function, inputs map[string]interface{}, strategy GenerationStrategy) interface{} {
	if len(function.ReturnTypes) == 0 {
		return nil
	}

	returnType := function.ReturnTypes[0]

	// Use function name heuristics to generate realistic expected values
	funcName := strings.ToLower(function.Name)

	switch {
	case strings.Contains(funcName, "add") || strings.Contains(funcName, "sum"):
		if isIntegerType(returnType) {
			return dg.generateArithmeticResult(inputs, "add")
		}
	case strings.Contains(funcName, "multiply") || strings.Contains(funcName, "mul"):
		if isIntegerType(returnType) {
			return dg.generateArithmeticResult(inputs, "multiply")
		}
	case strings.Contains(funcName, "length") || strings.Contains(funcName, "len"):
		return dg.generateLengthResult(inputs)
	case strings.Contains(funcName, "empty") || strings.Contains(funcName, "isempty"):
		return strategy == StrategyEdge || strategy == StrategyZero
	case strings.Contains(funcName, "valid") || strings.Contains(funcName, "isvalid"):
		return strategy == StrategyPositive
	}

	// Default expected value based on return type
	expectedValue, _ := dg.generateValueForType(returnType, StrategyPositive)
	return expectedValue
}

// generateArithmeticResult generates expected results for arithmetic operations
func (dg *DataGenerator) generateArithmeticResult(inputs map[string]interface{}, operation string) interface{} {
	// Simple implementation - in practice would be more sophisticated
	return 42
}

// generateLengthResult generates expected results for length operations
func (dg *DataGenerator) generateLengthResult(inputs map[string]interface{}) interface{} {
	// Analyze string or slice inputs to determine expected length
	for _, value := range inputs {
		if str, ok := value.(string); ok {
			return len(str)
		}
	}
	return 0
}

// shouldExpectError determines if a test case should expect an error
func (dg *DataGenerator) shouldExpectError(function *models.Function, strategy GenerationStrategy, inputs map[string]interface{}) bool {
	if !function.HasErrorReturn {
		return false
	}

	// Error expectations based on strategy and function characteristics
	switch strategy {
	case StrategyNegative:
		return true
	case StrategyEdge:
		return dg.hasProblematicInputs(function, inputs)
	default:
		return false
	}
}

// hasProblematicInputs checks if inputs might cause errors
func (dg *DataGenerator) hasProblematicInputs(function *models.Function, inputs map[string]interface{}) bool {
	funcName := strings.ToLower(function.Name)

	// Check for division by zero
	if strings.Contains(funcName, "div") {
		for name, value := range inputs {
			if strings.Contains(strings.ToLower(name), "divisor") ||
				strings.Contains(strings.ToLower(name), "denominator") {
				if intVal, ok := value.(int); ok && intVal == 0 {
					return true
				}
				if floatVal, ok := value.(float64); ok && floatVal == 0.0 {
					return true
				}
			}
		}
	}

	// Check for nil pointers
	for _, value := range inputs {
		if value == nil {
			return true
		}
	}

	return false
}

// generateErrorPattern generates expected error patterns
func (dg *DataGenerator) generateErrorPattern(function *models.Function, strategy GenerationStrategy) string {
	funcName := strings.ToLower(function.Name)

	switch {
	case strings.Contains(funcName, "div"):
		return "division by zero"
	case strings.Contains(funcName, "parse"):
		return "invalid"
	case strings.Contains(funcName, "open") || strings.Contains(funcName, "read"):
		return "no such file"
	case strings.Contains(funcName, "connect"):
		return "connection"
	default:
		return "error"
	}
}

// getStrategyDescription returns a human-readable description for a strategy
func (dg *DataGenerator) getStrategyDescription(strategy GenerationStrategy, functionName string) string {
	switch strategy {
	case StrategyPositive:
		return fmt.Sprintf("Test %s with valid positive inputs", functionName)
	case StrategyNegative:
		return fmt.Sprintf("Test %s with invalid/negative inputs", functionName)
	case StrategyEdge:
		return fmt.Sprintf("Test %s with edge case inputs", functionName)
	case StrategyZero:
		return fmt.Sprintf("Test %s with zero/empty inputs", functionName)
	case StrategyRandom:
		return fmt.Sprintf("Test %s with random inputs", functionName)
	default:
		return fmt.Sprintf("Test %s", functionName)
	}
}

// generateRandomString generates a random string of specified length
func (dg *DataGenerator) generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[dg.rand.Intn(len(charset))]
	}
	return string(b)
}

// getIntegerEdgeValues returns edge values for integer types
func (dg *DataGenerator) getIntegerEdgeValues(goType string) []int64 {
	switch goType {
	case "int8":
		return []int64{-128, -1, 0, 1, 127}
	case "uint8", "byte":
		return []int64{0, 1, 127, 255}
	case "int16":
		return []int64{-32768, -1, 0, 1, 32767}
	case "uint16":
		return []int64{0, 1, 32767, 65535}
	case "int32", "rune":
		return []int64{-2147483648, -1, 0, 1, 2147483647}
	case "uint32":
		return []int64{0, 1, 2147483647, 4294967295}
	case "int", "int64":
		return []int64{-9223372036854775808, -1, 0, 1, 9223372036854775807}
	case "uint", "uint64":
		return []int64{0, 1, 9223372036854775807} // Max safe value for int64
	default:
		return []int64{-1, 0, 1, 42, 100}
	}
}

// Type checking helper functions
func isIntegerType(t string) bool {
	intTypes := []string{"int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "byte", "rune"}
	for _, it := range intTypes {
		if t == it {
			return true
		}
	}
	return false
}

func isFloatType(t string) bool {
	return t == "float32" || t == "float64"
}
