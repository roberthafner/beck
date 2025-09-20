# Sample Projects for Go Coverage Analyzer

This directory contains sample Go projects designed to demonstrate and test the functionality of the Go Coverage Analyzer & Test Generator tool (`gcov`).

## Project Organization

### 📁 Project Structure
```
sample-projects/
├── simple-calculator/     # Basic Go project with minimal dependencies
├── user-service/          # Complex project with interfaces and dependency injection
└── README.md              # This file
```

## Available Sample Projects

### 🧮 simple-calculator
**Type:** Basic Go Project  
**Complexity:** Low  
**Use Case:** Demonstrates basic coverage analysis on a straightforward codebase

**Features:**
- Single file Go project (`calculator.go`)
- Basic arithmetic operations with method receivers
- Mix of simple and complex functions (varying cyclomatic complexity)
- Existing test coverage (~27.4% with current tests)
- Perfect for learning `gcov` basics

**Key Functions:**
- `NewCalculator()` - Constructor function
- `Add()`, `Subtract()`, `Multiply()`, `Divide()` - Basic arithmetic 
- `Power()`, `Sqrt()` - Mathematical operations
- `IsPrime()`, `Fibonacci()`, `Factorial()` - Algorithmic functions
- Utility functions: `IsEven()`, `Max()`, `Min()`, `Abs()`

**Coverage Highlights:**
- Well-tested: Basic arithmetic operations
- Untested: Advanced math functions and edge cases
- Good for demonstrating test generation capabilities

---

### 🏗️ user-service  
**Type:** Interface-Based Architecture  
**Complexity:** High  
**Use Case:** Demonstrates coverage analysis on projects with dependency injection, interfaces, and mocks

**Features:**
- Multiple Go files with clear separation of concerns
- Interface definitions for dependency injection
- Service layer with business logic
- Generated mock implementations
- Comprehensive test suite with table-driven tests
- Real-world patterns and practices

**Architecture:**
```
user-service/
├── interfaces.go          # Interface definitions
├── service.go            # Business logic (UserService, FileService)
├── main.go               # Application entry point + utilities
├── *_test.go             # Test files
├── mock_*.go             # Generated mock implementations
├── go.mod/go.sum         # Dependencies
└── coverage.out          # Coverage profile
```

**Key Interfaces:**
- `Repository` - Data storage abstraction
- `Logger` - Logging interface
- `EmailSender` - Email service interface  
- `Cache` - Caching interface
- `HTTPClient` - HTTP client interface
- `FileProcessor` - File processing interface

**Services:**
- `UserService` - User management with CRUD operations
- `FileService` - File processing service

**Coverage Highlights:**
- Current coverage: ~38.8%
- Well-tested: Service constructors, basic CRUD with mocks
- Untested: Complex service methods, error handling paths
- Excellent for demonstrating mock-based test generation

## Usage Examples

### Basic Analysis
```bash
# Analyze simple calculator project
./gcov analyze sample-projects/simple-calculator --verbose

# Analyze complex user service project  
./gcov analyze sample-projects/user-service --verbose
```

### Coverage Profile Generation
```bash
# Generate fresh coverage profile for simple calculator
./gcov analyze sample-projects/simple-calculator --profile --verbose

# Generate coverage for user service
./gcov analyze sample-projects/user-service --profile --verbose
```

### Different Output Formats
```bash
# JSON output for CI/CD integration
./gcov analyze sample-projects/simple-calculator --output json

# HTML report for detailed review
./gcov analyze sample-projects/user-service --output html
```

### Test Generation (Dry Run)
```bash
# Preview test generation for simple calculator
./gcov generate sample-projects/simple-calculator --dry-run --verbose

# Generate tests for user service with mocks
./gcov generate sample-projects/user-service --dry-run --verbose --generate-mocks
```

### Complexity Analysis
```bash
# Focus on high complexity functions only
./gcov analyze sample-projects/simple-calculator --min-complexity 3 --verbose

# Analyze complex functions in user service
./gcov analyze sample-projects/user-service --min-complexity 3 --verbose
```

## Project Characteristics

| Project | Files | Functions | Interfaces | Mocks | Coverage | Complexity |
|---------|-------|-----------|------------|-------|----------|------------|
| simple-calculator | 1 | 15 | 0 | 0 | ~27% | Low-Medium |
| user-service | 9 | 42 | 6 | 6 | ~39% | Medium-High |

## Learning Path

### 1. Start with simple-calculator
- Learn basic `gcov analyze` commands
- Understand coverage reports and metrics
- Practice with different output formats
- Experiment with threshold settings

### 2. Progress to user-service
- Explore interface-based architecture analysis
- Understand mock coverage patterns
- Learn about complex dependency analysis
- Practice with table-driven test patterns

### 3. Advanced Usage
- Compare analysis results between projects
- Practice test generation workflows
- Experiment with CI/CD integration patterns
- Customize report formats and thresholds

## Maintenance Notes

These sample projects are designed to be:
- **Stable**: Code structure should remain consistent for reliable testing
- **Representative**: Cover common Go patterns and practices
- **Educational**: Include various complexity levels and architectural patterns
- **Testable**: Have existing tests to demonstrate coverage analysis

When modifying these projects:
1. Maintain existing function signatures for consistency
2. Keep test coverage at current levels for baseline comparisons
3. Update this README if adding new projects or changing structure
4. Ensure go.mod files are properly maintained

## Contributing

When adding new sample projects:
1. Create a descriptive directory name
2. Include a brief description in this README
3. Provide both covered and uncovered code for demonstration
4. Include appropriate go.mod file
5. Add usage examples for the new project