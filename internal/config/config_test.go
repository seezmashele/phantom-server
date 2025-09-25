package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestGetDefaultConfig(t *testing.T) {
	config := GetDefaultConfig()

	if config.Server.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", config.Server.Port)
	}

	// Test hardcoded timeout values as required by task 2
	if config.Server.ShutdownTimeout != 30 {
		t.Errorf("Expected default shutdown timeout 30, got %d", config.Server.ShutdownTimeout)
	}

	if config.Server.ReadTimeout != 10 {
		t.Errorf("Expected default read timeout 10, got %d", config.Server.ReadTimeout)
	}

	if config.Server.WriteTimeout != 10 {
		t.Errorf("Expected default write timeout 10, got %d", config.Server.WriteTimeout)
	}

	// Test hardcoded HTTP methods as required by task 2
	expectedMethods := []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	if !reflect.DeepEqual(config.Server.AllowedMethods, expectedMethods) {
		t.Errorf("Expected default methods %v, got %v", expectedMethods, config.Server.AllowedMethods)
	}

	if !config.Server.EnableLogging {
		t.Error("Expected default logging to be enabled")
	}

	expectedOrigins := []string{"*"}
	if !reflect.DeepEqual(config.Server.AllowedOrigins, expectedOrigins) {
		t.Errorf("Expected default origins %v, got %v", expectedOrigins, config.Server.AllowedOrigins)
	}
}

func TestLoadConfig(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("Load valid JSON config", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "valid_config.json")
		configContent := `{
			"server": {
				"port": 3000,
				"shutdown_timeout_seconds": 45,
				"read_timeout_seconds": 15,
				"write_timeout_seconds": 15,
				"allowed_origins": ["http://localhost:3000"],
				"allowed_methods": ["GET", "POST"],
				"enable_logging": false
			}
		}`

		err := os.WriteFile(configPath, []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write test config file: %v", err)
		}

		config, err := LoadConfig(configPath)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		if config.Server.Port != 3000 {
			t.Errorf("Expected port 3000, got %d", config.Server.Port)
		}

		if config.Server.ShutdownTimeout != 45 {
			t.Errorf("Expected shutdown timeout 45, got %d", config.Server.ShutdownTimeout)
		}

		if config.Server.EnableLogging {
			t.Error("Expected logging to be disabled")
		}

		expectedOrigins := []string{"http://localhost:3000"}
		if !reflect.DeepEqual(config.Server.AllowedOrigins, expectedOrigins) {
			t.Errorf("Expected origins %v, got %v", expectedOrigins, config.Server.AllowedOrigins)
		}
	})

	t.Run("Load non-existent config file", func(t *testing.T) {
		_, err := LoadConfig("non_existent_file.json")
		if err == nil {
			t.Error("Expected error when loading non-existent file")
		}
	})

	t.Run("Load invalid JSON config", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "invalid_config.json")
		invalidContent := `{
			"server": {
				"port": "invalid_port"
			}
		`

		err := os.WriteFile(configPath, []byte(invalidContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write invalid test config file: %v", err)
		}

		_, err = LoadConfig(configPath)
		if err == nil {
			t.Error("Expected error when loading invalid JSON")
		}
	})
}

func TestWriteConfig(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("Write and read config", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "write_test.json")

		// Create a test config
		testConfig := &Config{
			Server: ServerConfig{
				Port:            9000,
				ShutdownTimeout: 60,
				ReadTimeout:     20,
				WriteTimeout:    20,
				AllowedOrigins:  []string{"https://example.com"},
				AllowedMethods:  []string{"GET"},
				EnableLogging:   false,
			},
		}

		// Write the config
		err := WriteConfig(configPath, testConfig)
		if err != nil {
			t.Fatalf("Failed to write config: %v", err)
		}

		// Read it back
		loadedConfig, err := LoadConfig(configPath)
		if err != nil {
			t.Fatalf("Failed to load written config: %v", err)
		}

		// Compare
		if !reflect.DeepEqual(testConfig, loadedConfig) {
			t.Errorf("Written and loaded configs don't match.\nExpected: %+v\nGot: %+v", testConfig, loadedConfig)
		}
	})
}

func TestLoadEnvConfig(t *testing.T) {
	// Create a temporary directory for test .env files
	tempDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Save current working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	defer os.Chdir(originalWd)

	t.Run("Load default config when no .env file exists", func(t *testing.T) {
		// Change to temp directory where no .env file exists
		os.Chdir(tempDir)

		config, err := LoadEnvConfig()
		if err != nil {
			t.Fatalf("Failed to load env config: %v", err)
		}

		defaultConfig := GetDefaultConfig()
		if !reflect.DeepEqual(config, defaultConfig) {
			t.Errorf("Expected default config when no .env file exists.\nExpected: %+v\nGot: %+v", defaultConfig, config)
		}
	})

	t.Run("Load config from .env file", func(t *testing.T) {
		// Create .env file in temp directory
		envPath := filepath.Join(tempDir, ".env")
		envContent := `PORT=3000
SHUTDOWN_TIMEOUT=45
READ_TIMEOUT=15
WRITE_TIMEOUT=15
ALLOWED_ORIGINS=http://localhost:3000,https://example.com
ALLOWED_METHODS=GET,POST,PUT
ENABLE_LOGGING=false`

		err := os.WriteFile(envPath, []byte(envContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write .env file: %v", err)
		}

		// Change to temp directory
		os.Chdir(tempDir)

		config, err := LoadEnvConfig()
		if err != nil {
			t.Fatalf("Failed to load env config: %v", err)
		}

		if config.Server.Port != 3000 {
			t.Errorf("Expected port 3000, got %d", config.Server.Port)
		}

		if config.Server.ShutdownTimeout != 45 {
			t.Errorf("Expected shutdown timeout 45, got %d", config.Server.ShutdownTimeout)
		}

		if config.Server.ReadTimeout != 15 {
			t.Errorf("Expected read timeout 15, got %d", config.Server.ReadTimeout)
		}

		if config.Server.WriteTimeout != 15 {
			t.Errorf("Expected write timeout 15, got %d", config.Server.WriteTimeout)
		}

		expectedOrigins := []string{"http://localhost:3000", "https://example.com"}
		if !reflect.DeepEqual(config.Server.AllowedOrigins, expectedOrigins) {
			t.Errorf("Expected origins %v, got %v", expectedOrigins, config.Server.AllowedOrigins)
		}

		expectedMethods := []string{"GET", "POST", "PUT"}
		if !reflect.DeepEqual(config.Server.AllowedMethods, expectedMethods) {
			t.Errorf("Expected methods %v, got %v", expectedMethods, config.Server.AllowedMethods)
		}

		if config.Server.EnableLogging {
			t.Error("Expected logging to be disabled")
		}
	})

	t.Run("Handle invalid .env file values gracefully", func(t *testing.T) {
		// Create .env file with invalid values
		envPath := filepath.Join(tempDir, ".env")
		envContent := `PORT=invalid_port
SHUTDOWN_TIMEOUT=invalid_timeout
ENABLE_LOGGING=invalid_bool`

		err := os.WriteFile(envPath, []byte(envContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write .env file: %v", err)
		}

		// Change to temp directory
		os.Chdir(tempDir)

		config, err := LoadEnvConfig()
		if err != nil {
			t.Fatalf("Failed to load env config: %v", err)
		}

		// Should fall back to defaults for invalid values
		defaultConfig := GetDefaultConfig()
		if config.Server.Port != defaultConfig.Server.Port {
			t.Errorf("Expected default port %d for invalid .env var, got %d", defaultConfig.Server.Port, config.Server.Port)
		}

		if config.Server.ShutdownTimeout != defaultConfig.Server.ShutdownTimeout {
			t.Errorf("Expected default shutdown timeout %d for invalid .env var, got %d", defaultConfig.Server.ShutdownTimeout, config.Server.ShutdownTimeout)
		}

		if config.Server.EnableLogging != defaultConfig.Server.EnableLogging {
			t.Errorf("Expected default logging %v for invalid .env var, got %v", defaultConfig.Server.EnableLogging, config.Server.EnableLogging)
		}
	})

	t.Run("Load partial config from .env file", func(t *testing.T) {
		// Create .env file with only some values
		envPath := filepath.Join(tempDir, ".env")
		envContent := `PORT=4000
ENABLE_LOGGING=false`

		err := os.WriteFile(envPath, []byte(envContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write .env file: %v", err)
		}

		// Change to temp directory
		os.Chdir(tempDir)

		config, err := LoadEnvConfig()
		if err != nil {
			t.Fatalf("Failed to load env config: %v", err)
		}

		// Should use .env values where provided
		if config.Server.Port != 4000 {
			t.Errorf("Expected port 4000, got %d", config.Server.Port)
		}

		if config.Server.EnableLogging {
			t.Error("Expected logging to be disabled")
		}

		// Should use defaults for non-provided values
		defaultConfig := GetDefaultConfig()
		if config.Server.ShutdownTimeout != defaultConfig.Server.ShutdownTimeout {
			t.Errorf("Expected default shutdown timeout %d, got %d", defaultConfig.Server.ShutdownTimeout, config.Server.ShutdownTimeout)
		}
	})
}

func TestMergeConfigs(t *testing.T) {
	t.Run("Merge with nil base config", func(t *testing.T) {
		override := &Config{
			Server: ServerConfig{
				Port: 3000,
			},
		}

		result := MergeConfigs(nil, override)

		// Should use default config as base
		if result.Server.Port != 3000 {
			t.Errorf("Expected port 3000, got %d", result.Server.Port)
		}

		// Should keep default values for non-overridden fields
		defaultConfig := GetDefaultConfig()
		if result.Server.ShutdownTimeout != defaultConfig.Server.ShutdownTimeout {
			t.Errorf("Expected default shutdown timeout %d, got %d", defaultConfig.Server.ShutdownTimeout, result.Server.ShutdownTimeout)
		}
	})

	t.Run("Merge with nil override config", func(t *testing.T) {
		base := &Config{
			Server: ServerConfig{
				Port:            9000,
				ShutdownTimeout: 60,
				EnableLogging:   false,
			},
		}

		result := MergeConfigs(base, nil)

		if !reflect.DeepEqual(result, base) {
			t.Errorf("Expected base config when override is nil.\nExpected: %+v\nGot: %+v", base, result)
		}
	})

	t.Run("Merge configs with override taking priority", func(t *testing.T) {
		base := &Config{
			Server: ServerConfig{
				Port:            8080,
				ShutdownTimeout: 30,
				ReadTimeout:     10,
				WriteTimeout:    10,
				AllowedOrigins:  []string{"*"},
				AllowedMethods:  []string{"GET", "POST"},
				EnableLogging:   true,
			},
		}

		override := &Config{
			Server: ServerConfig{
				Port:            3000,
				ShutdownTimeout: 45,
				AllowedOrigins:  []string{"http://localhost:3000"},
				EnableLogging:   false,
			},
		}

		result := MergeConfigs(base, override)

		// Override values should take priority
		if result.Server.Port != 3000 {
			t.Errorf("Expected port 3000, got %d", result.Server.Port)
		}

		if result.Server.ShutdownTimeout != 45 {
			t.Errorf("Expected shutdown timeout 45, got %d", result.Server.ShutdownTimeout)
		}

		expectedOrigins := []string{"http://localhost:3000"}
		if !reflect.DeepEqual(result.Server.AllowedOrigins, expectedOrigins) {
			t.Errorf("Expected origins %v, got %v", expectedOrigins, result.Server.AllowedOrigins)
		}

		if result.Server.EnableLogging != false {
			t.Errorf("Expected logging false, got %v", result.Server.EnableLogging)
		}

		// Base values should be kept for non-overridden fields
		if result.Server.ReadTimeout != 10 {
			t.Errorf("Expected read timeout 10, got %d", result.Server.ReadTimeout)
		}

		if result.Server.WriteTimeout != 10 {
			t.Errorf("Expected write timeout 10, got %d", result.Server.WriteTimeout)
		}

		expectedMethods := []string{"GET", "POST"}
		if !reflect.DeepEqual(result.Server.AllowedMethods, expectedMethods) {
			t.Errorf("Expected methods %v, got %v", expectedMethods, result.Server.AllowedMethods)
		}
	})

	t.Run("Merge configs with zero values in override", func(t *testing.T) {
		base := &Config{
			Server: ServerConfig{
				Port:            8080,
				ShutdownTimeout: 30,
				ReadTimeout:     10,
				WriteTimeout:    10,
				AllowedOrigins:  []string{"*"},
				AllowedMethods:  []string{"GET", "POST"},
				EnableLogging:   true,
			},
		}

		override := &Config{
			Server: ServerConfig{
				Port:            0,          // Zero value should not override
				ShutdownTimeout: 0,          // Zero value should not override
				AllowedOrigins:  []string{}, // Empty slice should not override
				EnableLogging:   false,      // Boolean false should override
			},
		}

		result := MergeConfigs(base, override)

		// Zero values should not override base values
		if result.Server.Port != 8080 {
			t.Errorf("Expected port 8080 (zero value should not override), got %d", result.Server.Port)
		}

		if result.Server.ShutdownTimeout != 30 {
			t.Errorf("Expected shutdown timeout 30 (zero value should not override), got %d", result.Server.ShutdownTimeout)
		}

		expectedOrigins := []string{"*"}
		if !reflect.DeepEqual(result.Server.AllowedOrigins, expectedOrigins) {
			t.Errorf("Expected origins %v (empty slice should not override), got %v", expectedOrigins, result.Server.AllowedOrigins)
		}

		// Boolean false should override
		if result.Server.EnableLogging != false {
			t.Errorf("Expected logging false (boolean should override), got %v", result.Server.EnableLogging)
		}
	})
}
