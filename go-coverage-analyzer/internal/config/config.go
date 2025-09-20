package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	// Analysis settings
	ExcludeDirs         []string  `mapstructure:"exclude_dirs"`
	IncludeTests        bool      `mapstructure:"include_tests"`
	CoverageThreshold   float64   `mapstructure:"coverage_threshold"`
	CalculateComplexity bool      `mapstructure:"calculate_complexity"`
	MinComplexity       int       `mapstructure:"min_complexity"`
	
	// Output and reporting settings
	OutputFormat        string    `mapstructure:"output_format"`
	OutputDir           string    `mapstructure:"output_dir"`
	Verbose             bool      `mapstructure:"verbose"`
	ProfileOutput       string    `mapstructure:"profile_output"`
	
	// Test generation settings
	TemplateStyle       string    `mapstructure:"template_style"`
	GenerateMocks       bool      `mapstructure:"generate_mocks"`
	TableDriven         bool      `mapstructure:"table_driven"`
	GenerateBenchmarks  bool      `mapstructure:"generate_benchmarks"`
	OverwriteTests      bool      `mapstructure:"overwrite_tests"`
	MaxTestCases        int       `mapstructure:"max_test_cases"`
	IgnoreFunctions     []string  `mapstructure:"ignore_functions"`
	
	// Template configuration
	Templates           TemplateConfig `mapstructure:"templates"`
	
	// Performance settings
	MaxConcurrency      int       `mapstructure:"max_concurrency"`
	EnableCaching       bool      `mapstructure:"enable_caching"`
	CacheDir            string    `mapstructure:"cache_dir"`
	
	// Advanced settings
	CustomPatterns      []string  `mapstructure:"custom_patterns"`
	GoVersions          []string  `mapstructure:"go_versions"`
	BuildTags           []string  `mapstructure:"build_tags"`
}

// TemplateConfig holds template-specific configuration
type TemplateConfig struct {
	CustomTemplatesDir  string            `mapstructure:"custom_templates_dir"`
	FunctionTemplate    string            `mapstructure:"function_template"`
	MethodTemplate      string            `mapstructure:"method_template"`
	TableTemplate       string            `mapstructure:"table_template"`
	BenchmarkTemplate   string            `mapstructure:"benchmark_template"`
	MockTemplate        string            `mapstructure:"mock_template"`
	Overrides          map[string]string `mapstructure:"overrides"`
}

// Load loads configuration from default locations
func Load() (*Config, error) {
	v := viper.New()
	setDefaults(v)
	
	// Set up configuration file search paths
	v.SetConfigName("gcov")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("$HOME/.config/gcov")
	v.AddConfigPath("/etc/gcov")
	
	// Enable environment variable support
	v.SetEnvPrefix("GCOV")
	v.AutomaticEnv()
	
	// Try to read config file (don't error if not found)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}
	
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}
	
	return &config, nil
}

// LoadFromFile loads configuration from a specific file
func LoadFromFile(configPath string) (*Config, error) {
	v := viper.New()
	setDefaults(v)
	
	// Set config file path
	v.SetConfigFile(configPath)
	
	// Enable environment variable support
	v.SetEnvPrefix("GCOV")
	v.AutomaticEnv()
	
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file %s: %w", configPath, err)
	}
	
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}
	
	return &config, nil
}

// Save saves the configuration to a file
func (c *Config) Save(configPath string) error {
	v := viper.New()
	
	// Set all configuration values
	v.Set("exclude_dirs", c.ExcludeDirs)
	v.Set("include_tests", c.IncludeTests)
	v.Set("coverage_threshold", c.CoverageThreshold)
	v.Set("calculate_complexity", c.CalculateComplexity)
	v.Set("min_complexity", c.MinComplexity)
	
	v.Set("output_format", c.OutputFormat)
	v.Set("output_dir", c.OutputDir)
	v.Set("verbose", c.Verbose)
	v.Set("profile_output", c.ProfileOutput)
	
	v.Set("template_style", c.TemplateStyle)
	v.Set("generate_mocks", c.GenerateMocks)
	v.Set("table_driven", c.TableDriven)
	v.Set("generate_benchmarks", c.GenerateBenchmarks)
	v.Set("overwrite_tests", c.OverwriteTests)
	v.Set("max_test_cases", c.MaxTestCases)
	v.Set("ignore_functions", c.IgnoreFunctions)
	
	v.Set("templates", c.Templates)
	
	v.Set("max_concurrency", c.MaxConcurrency)
	v.Set("enable_caching", c.EnableCaching)
	v.Set("cache_dir", c.CacheDir)
	
	v.Set("custom_patterns", c.CustomPatterns)
	v.Set("go_versions", c.GoVersions)
	v.Set("build_tags", c.BuildTags)
	
	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}
	
	return v.WriteConfigAs(configPath)
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate coverage threshold
	if c.CoverageThreshold < 0 || c.CoverageThreshold > 100 {
		return fmt.Errorf("coverage_threshold must be between 0 and 100, got %f", c.CoverageThreshold)
	}
	
	// Validate min complexity
	if c.MinComplexity < 1 {
		return fmt.Errorf("min_complexity must be at least 1, got %d", c.MinComplexity)
	}
	
	// Validate max test cases
	if c.MaxTestCases < 1 {
		return fmt.Errorf("max_test_cases must be at least 1, got %d", c.MaxTestCases)
	}
	
	// Validate max concurrency
	if c.MaxConcurrency < 1 {
		return fmt.Errorf("max_concurrency must be at least 1, got %d", c.MaxConcurrency)
	}
	
	// Validate output format
	validFormats := map[string]bool{
		"console": true,
		"json":    true,
		"html":    true,
		"xml":     true,
	}
	if !validFormats[c.OutputFormat] {
		return fmt.Errorf("invalid output_format: %s (valid: console, json, html, xml)", c.OutputFormat)
	}
	
	// Validate template style
	validStyles := map[string]bool{
		"standard": true,
		"testify":  true,
		"table":    true,
		"ginkgo":   true,
	}
	if !validStyles[c.TemplateStyle] {
		return fmt.Errorf("invalid template_style: %s (valid: standard, testify, table, ginkgo)", c.TemplateStyle)
	}
	
	// Validate custom templates directory if specified
	if c.Templates.CustomTemplatesDir != "" {
		if _, err := os.Stat(c.Templates.CustomTemplatesDir); os.IsNotExist(err) {
			return fmt.Errorf("custom_templates_dir does not exist: %s", c.Templates.CustomTemplatesDir)
		}
	}
	
	return nil
}

// GetCacheDir returns the cache directory, creating it if necessary
func (c *Config) GetCacheDir() (string, error) {
	cacheDir := c.CacheDir
	if cacheDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("error getting home directory: %w", err)
		}
		cacheDir = filepath.Join(homeDir, ".cache", "gcov")
	}
	
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("error creating cache directory: %w", err)
	}
	
	return cacheDir, nil
}

// GetTemplatesDir returns the templates directory
func (c *Config) GetTemplatesDir() string {
	if c.Templates.CustomTemplatesDir != "" {
		return c.Templates.CustomTemplatesDir
	}
	
	// Return default templates directory (embedded or relative to binary)
	return "templates"
}

// IsIgnoredFunction checks if a function should be ignored based on patterns
func (c *Config) IsIgnoredFunction(functionName string) bool {
	for _, pattern := range c.IgnoreFunctions {
		if matched, _ := filepath.Match(pattern, functionName); matched {
			return true
		}
	}
	return false
}

// GetTemplate returns the template for a specific type
func (c *Config) GetTemplate(templateType string) string {
	// Check for override first
	if override, exists := c.Templates.Overrides[templateType]; exists {
		return override
	}
	
	// Return specific template or default
	switch templateType {
	case "function":
		if c.Templates.FunctionTemplate != "" {
			return c.Templates.FunctionTemplate
		}
		return "function_test.tmpl"
	case "method":
		if c.Templates.MethodTemplate != "" {
			return c.Templates.MethodTemplate
		}
		return "method_test.tmpl"
	case "table":
		if c.Templates.TableTemplate != "" {
			return c.Templates.TableTemplate
		}
		return "table_test.tmpl"
	case "benchmark":
		if c.Templates.BenchmarkTemplate != "" {
			return c.Templates.BenchmarkTemplate
		}
		return "benchmark_test.tmpl"
	case "mock":
		if c.Templates.MockTemplate != "" {
			return c.Templates.MockTemplate
		}
		return "mock.tmpl"
	default:
		return "function_test.tmpl"
	}
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Analysis defaults
	v.SetDefault("exclude_dirs", []string{"vendor", "testdata", ".git", "node_modules"})
	v.SetDefault("include_tests", false)
	v.SetDefault("coverage_threshold", 80.0)
	v.SetDefault("calculate_complexity", true)
	v.SetDefault("min_complexity", 1)
	
	// Output defaults
	v.SetDefault("output_format", "console")
	v.SetDefault("output_dir", ".")
	v.SetDefault("verbose", false)
	v.SetDefault("profile_output", "coverage.out")
	
	// Generation defaults
	v.SetDefault("template_style", "standard")
	v.SetDefault("generate_mocks", true)
	v.SetDefault("table_driven", true)
	v.SetDefault("generate_benchmarks", false)
	v.SetDefault("overwrite_tests", false)
	v.SetDefault("max_test_cases", 10)
	v.SetDefault("ignore_functions", []string{})
	
	// Template defaults
	v.SetDefault("templates.custom_templates_dir", "")
	v.SetDefault("templates.function_template", "")
	v.SetDefault("templates.method_template", "")
	v.SetDefault("templates.table_template", "")
	v.SetDefault("templates.benchmark_template", "")
	v.SetDefault("templates.mock_template", "")
	v.SetDefault("templates.overrides", map[string]string{})
	
	// Performance defaults
	v.SetDefault("max_concurrency", 4)
	v.SetDefault("enable_caching", true)
	v.SetDefault("cache_dir", "")
	
	// Advanced defaults
	v.SetDefault("custom_patterns", []string{})
	v.SetDefault("go_versions", []string{})
	v.SetDefault("build_tags", []string{})
}