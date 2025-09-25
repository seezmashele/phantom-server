# Implementation Plan

- [x] 1. Update ServerConfig struct to remove timeout and methods from JSON tags
  - Modify the ServerConfig struct in internal/config/config.go to remove JSON tags from timeout fields and allowed_methods field
  - Keep the Go struct fields for internal use but prevent them from being marshaled/unmarshaled from JSON
  - _Requirements: 1.2, 2.2_

- [x] 2. Update GetDefaultConfig function with hardcoded timeout and method values
  - Modify GetDefaultConfig in internal/config/config.go to set hardcoded timeout values (shutdown: 30s, read: 10s, write: 10s)
  - Set hardcoded HTTP methods array ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
  - Ensure these values cannot be overridden by external configuration
  - _Requirements: 1.1, 1.4, 2.1, 2.4_

- [x] 3. Remove timeout and methods parsing from LoadEnvConfig function
  - Remove parsing of SHUTDOWN_TIMEOUT, READ_TIMEOUT, WRITE_TIMEOUT environment variables from LoadEnvConfig
  - Remove parsing of ALLOWED_METHODS environment variable from LoadEnvConfig
  - Remove CONFIG_PATH environment variable processing
  - _Requirements: 1.3, 2.3, 4.2_

- [ ] 4. Delete current config test file and create new config tests
  - Delete the existing internal/config/config_test.go file
  - Create new config tests that reflect the new hardcoded timeout and methods behavior
  - Write tests for GetDefaultConfig to verify hardcoded values are always returned
  - Write tests for LoadConfig to verify timeout and methods fields in JSON are ignored
  - Write tests for LoadEnvConfig to verify timeout and methods environment variables are ignored
  - Write tests for MergeConfigs to verify timeout and methods are never overridden
  - _Requirements: 5.1_

- [ ] 5. Update MergeConfigs function to handle reduced configuration surface
  - Modify MergeConfigs in internal/config/config.go to not merge timeout fields from override config
  - Modify MergeConfigs to not merge allowed_methods field from override config
  - Ensure timeout and methods values always come from the base (default) configuration
  - _Requirements: 1.2, 2.2_

- [ ] 6. Update main.go to use fixed config.json path
  - Modify loadConfiguration function in main.go to remove CONFIG_PATH environment variable lookup
  - Change JSON config loading to use fixed path "config.json" in application root
  - Remove godotenv.Read() call for CONFIG_PATH processing
  - _Requirements: 3.1, 3.3, 4.1, 4.3_

- [ ] 7. Update config.json to remove timeout and methods fields
  - Remove shutdown_timeout_seconds, read_timeout_seconds, write_timeout_seconds fields from config.json
  - Remove allowed_methods field from config.json
  - Keep port, allowed_origins, enable_logging fields
  - _Requirements: 1.2, 2.2_

- [ ] 8. Update .env.example to remove timeout and methods variables
  - Remove SHUTDOWN_TIMEOUT, READ_TIMEOUT, WRITE_TIMEOUT variables from .env.example
  - Remove ALLOWED_METHODS variable from .env.example
  - Remove CONFIG_PATH variable from .env.example
  - Add APP_ENV variable to .env.example for environment specification
  - Update comments to reflect the changes
  - _Requirements: 1.3, 2.3, 4.1_

- [ ] 9. Update integration tests for new configuration behavior
  - Modify TestServerStartupWithDifferentConfigurations to test fixed config.json path
  - Update test to verify timeout values are always hardcoded
  - Update test configuration files to remove timeout and methods fields
  - Test that applications work without CONFIG_PATH
  - _Requirements: 5.4_

- [ ] 10. Add configuration logging after server startup
  - Modify main.go to log the final configuration values after server starts
  - Log port, logging status, and allowed origins values
  - Add structured logging to show which configuration values are active
  - _Requirements: 5.1_

- [ ] 11. Run all tests to ensure no regressions
  - Execute all unit tests in config package and verify they pass
  - Execute all handler tests and verify they pass (should be unchanged)
  - Execute all middleware tests and verify they pass (should be unchanged)
  - Execute all integration tests and verify they pass
  - Fix any test failures and ensure clear error messages
  - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5_