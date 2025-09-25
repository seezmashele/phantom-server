# Requirements Document

## Introduction

This feature involves refactoring the HTTP server configuration management to consolidate timeout settings and HTTP methods as hardcoded defaults in the Go configuration code, while removing them from external configuration files. The goal is to simplify configuration management by moving static configuration values into the codebase and removing the CONFIG_PATH environment variable dependency.

## Requirements

### Requirement 1

**User Story:** As a developer, I want timeout settings to be hardcoded in the Go configuration code, so that I don't need to manage these values in external configuration files.

#### Acceptance Criteria

1. WHEN the application starts THEN the system SHALL use hardcoded timeout values defined in config.go
2. WHEN loading configuration THEN the system SHALL NOT read timeout settings from config.json
3. WHEN loading configuration THEN the system SHALL NOT read timeout settings from .env files
4. WHEN the application runs THEN the system SHALL use the following hardcoded timeout values:
   - Shutdown timeout: 30 seconds
   - Read timeout: 10 seconds  
   - Write timeout: 10 seconds

### Requirement 2

**User Story:** As a developer, I want HTTP methods to be hardcoded in the Go configuration code, so that the allowed methods are consistent and not configurable externally.

#### Acceptance Criteria

1. WHEN the application starts THEN the system SHALL use hardcoded HTTP methods defined in config.go
2. WHEN loading configuration THEN the system SHALL NOT read allowed_methods from config.json
3. WHEN loading configuration THEN the system SHALL NOT read ALLOWED_METHODS from .env files
4. WHEN the application runs THEN the system SHALL use the following hardcoded HTTP methods: GET, POST, PUT, DELETE, OPTIONS

### Requirement 3

**User Story:** As a developer, I want the config.json file to be located in the application root directory by default, so that I don't need to specify a CONFIG_PATH environment variable.

#### Acceptance Criteria

1. WHEN the application starts THEN the system SHALL look for config.json in the root directory by default
2. WHEN loading configuration THEN the system SHALL NOT require CONFIG_PATH environment variable
3. WHEN CONFIG_PATH is not specified THEN the system SHALL attempt to load config.json from the application root
4. WHEN config.json does not exist in the root THEN the system SHALL continue with default values without error

### Requirement 4

**User Story:** As a developer, I want to remove CONFIG_PATH from environment configuration, so that the configuration loading is simplified and more predictable.

#### Acceptance Criteria

1. WHEN updating .env.example THEN the system SHALL NOT include CONFIG_PATH variable
2. WHEN loading environment configuration THEN the system SHALL NOT process CONFIG_PATH variable
3. WHEN the application starts THEN the system SHALL use a fixed path for config.json in the root directory

### Requirement 5

**User Story:** As a developer, I want all existing tests to pass after the refactoring, so that I can ensure the changes don't break existing functionality.

#### Acceptance Criteria

1. WHEN running unit tests THEN all config package tests SHALL pass
2. WHEN running handler tests THEN all handler tests SHALL pass  
3. WHEN running middleware tests THEN all middleware tests SHALL pass
4. WHEN running integration tests THEN all integration tests SHALL pass
5. WHEN tests fail THEN the system SHALL provide clear error messages indicating what needs to be fixed