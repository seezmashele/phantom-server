package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/goccy/go-json"
	"github.com/joho/godotenv"
)

// Config represents the application configuration
type Config struct {
	Server ServerConfig `json:"server"`
}

// ServerConfig represents the HTTP server configuration
type ServerConfig struct {
	Port            int      `json:"port"`
	ShutdownTimeout int      // Hardcoded timeout value, not configurable via JSON
	ReadTimeout     int      // Hardcoded timeout value, not configurable via JSON
	WriteTimeout    int      // Hardcoded timeout value, not configurable via JSON
	AllowedOrigins  []string `json:"allowed_origins"`
	AllowedMethods  []string // Hardcoded HTTP methods, not configurable via JSON
	EnableLogging   bool     `json:"enable_logging"`
}

// GetDefaultConfig returns the default configuration with sensible defaults
func GetDefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:            8080,
			ShutdownTimeout: 30,
			ReadTimeout:     10,
			WriteTimeout:    10,
			AllowedOrigins:  []string{"*"},
			AllowedMethods:  []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			EnableLogging:   true,
		},
	}
}

// LoadConfig loads configuration from a JSON file using goccy/go-json
func LoadConfig(path string) (*Config, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist: %s", path)
	}

	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse JSON config: %w", err)
	}

	return &config, nil
}

// WriteConfig writes configuration to a JSON file using goccy/go-json
func WriteConfig(path string, config *Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config to JSON: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// LoadEnvConfig loads configuration from .env files using godotenv
func LoadEnvConfig() (*Config, error) {
	// Load .env file if it exists (ignore error if file doesn't exist)
	envVars, err := godotenv.Read()
	if err != nil {
		// If .env file doesn't exist, return empty config (will use defaults)
		return GetDefaultConfig(), nil
	}

	config := GetDefaultConfig()

	// Parse PORT
	if portStr, exists := envVars["PORT"]; exists && portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			config.Server.Port = port
		}
	}

	// Parse ALLOWED_ORIGINS
	if originsStr, exists := envVars["ALLOWED_ORIGINS"]; exists && originsStr != "" {
		origins := strings.Split(originsStr, ",")
		for i, origin := range origins {
			origins[i] = strings.TrimSpace(origin)
		}
		config.Server.AllowedOrigins = origins
	}

	// Parse ENABLE_LOGGING
	if loggingStr, exists := envVars["ENABLE_LOGGING"]; exists && loggingStr != "" {
		if logging, err := strconv.ParseBool(loggingStr); err == nil {
			config.Server.EnableLogging = logging
		}
	}

	return config, nil
}

// MergeConfigs merges two configurations with the override config taking priority
// Timeout and methods values are never overridden (always use base/hardcoded values)
func MergeConfigs(base, override *Config) *Config {
	if base == nil {
		base = GetDefaultConfig()
	}
	if override == nil {
		return base
	}

	result := &Config{
		Server: ServerConfig{
			Port:            base.Server.Port,
			ShutdownTimeout: base.Server.ShutdownTimeout, // Always use base (hardcoded) values
			ReadTimeout:     base.Server.ReadTimeout,     // Always use base (hardcoded) values
			WriteTimeout:    base.Server.WriteTimeout,    // Always use base (hardcoded) values
			AllowedOrigins:  make([]string, len(base.Server.AllowedOrigins)),
			AllowedMethods:  make([]string, len(base.Server.AllowedMethods)), // Always use base (hardcoded) values
			EnableLogging:   base.Server.EnableLogging,
		},
	}

	// Copy slices from base (timeout and methods are never overridden)
	copy(result.Server.AllowedOrigins, base.Server.AllowedOrigins)
	copy(result.Server.AllowedMethods, base.Server.AllowedMethods)

	// Override with non-zero values from override config (excluding timeout and methods)
	if override.Server.Port != 0 {
		result.Server.Port = override.Server.Port
	}
	// Timeout values are intentionally NOT overridden - they remain hardcoded
	if len(override.Server.AllowedOrigins) > 0 {
		result.Server.AllowedOrigins = make([]string, len(override.Server.AllowedOrigins))
		copy(result.Server.AllowedOrigins, override.Server.AllowedOrigins)
	}
	// AllowedMethods are intentionally NOT overridden - they remain hardcoded
	// For boolean values, we need to check if they differ from the default
	// Since we can't distinguish between false and unset, we'll always use the override value
	result.Server.EnableLogging = override.Server.EnableLogging

	return result
}
