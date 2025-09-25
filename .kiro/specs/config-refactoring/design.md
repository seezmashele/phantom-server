# Design Document

## Overview

This design outlines the refactoring of the HTTP server configuration management system to consolidate timeout settings and HTTP methods as hardcoded defaults in the Go configuration code. The refactoring will simplify configuration management by removing these values from external configuration files (config.json and .env) and eliminating the CONFIG_PATH environment variable dependency.

## Architecture

The current configuration system uses a three-tier priority system:
1. Environment variables (highest priority)
2. JSON configuration file (medium priority) 
3. Default values in Go code (lowest priority)

After refactoring, the system will maintain the same architecture but with a simplified configuration surface:
- **Hardcoded in Go**: Timeout settings and HTTP methods
- **Configurable via JSON/env**: Port, CORS origins, logging settings
- **Fixed path**: config.json always loaded from application root

## Components and Interfaces

### Modified Components

#### 1. Config Package (`internal/config/config.go`)

**Changes to ServerConfig struct:**
- Remove timeout fields from JSON tags (keep Go fields for internal use)
- Remove allowed_methods field from JSON tags (keep Go field for internal use)

**Changes to GetDefaultConfig function:**
- Hardcode timeout values that cannot be overridden
- Hardcode HTTP methods that cannot be overridden

**Changes to LoadEnvConfig function:**
- Remove parsing of timeout environment variables
- Remove parsing of ALLOWED_METHODS environment variable
- Remove CONFIG_PATH processing

**New behavior:**
- Timeout values always use hardcoded defaults
- HTTP methods always use hardcoded defaults
- JSON config loading uses fixed "config.json" path in root

#### 2. Main Package (`main.go`)

**Changes to loadConfiguration function:**
- Remove CONFIG_PATH environment variable lookup
- Use fixed path "config.json" for JSON configuration loading
- Maintain same merge priority but with reduced configuration surface

#### 3. Configuration Files

**config.json changes:**
- Remove timeout-related fields
- Remove allowed_methods field
- Keep port, allowed_origins, enable_logging

**.env.example changes:**
- Remove timeout-related variables
- Remove ALLOWED_METHODS variable
- Remove CONFIG_PATH variable
- Keep PORT, ALLOWED_ORIGINS, ENABLE_LOGGING

## Data Models

### Updated ServerConfig Structure

```go
type ServerConfig struct {
    // Configurable fields (can be set via JSON/env)
    Port           int      `json:"port"`
    AllowedOrigins []string `json:"allowed_origins"`
    EnableLogging  bool     `json:"enable_logging"`
    
    // Hardcoded fields (not in JSON, set by defaults only)
    ShutdownTimeout int      // Always 30 seconds
    ReadTimeout     int      // Always 10 seconds  
    WriteTimeout    int      // Always 10 seconds
    AllowedMethods  []string // Always ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
}
```

### Configuration Loading Flow

```
Application Start
       ↓
Load Default Config (with hardcoded timeouts/methods)
       ↓
Try Load config.json from root (optional)
       ↓
Merge JSON config (only configurable fields)
       ↓
Load .env config (optional)
       ↓
Merge env config (only configurable fields)
       ↓
Final Configuration
```

## Error Handling

### Configuration Loading Errors

1. **Missing config.json**: Continue with defaults (no error)
2. **Invalid JSON in config.json**: Log warning, continue with defaults
3. **Invalid .env values**: Use defaults for invalid fields, continue
4. **Missing .env file**: Continue with defaults (no error)

### Backward Compatibility

- Applications currently using timeout settings in config.json will see those settings ignored (logged as warning)
- Applications using timeout environment variables will see those ignored (no warning needed)
- Applications using CONFIG_PATH will see it ignored (no warning needed)
- All applications will continue to function with hardcoded timeout values

## Testing Strategy

### Unit Tests

1. **Config Package Tests**
   - Update `TestGetDefaultConfig` to verify hardcoded values
   - Update `TestLoadConfig` to ignore timeout fields in JSON
   - Update `TestLoadEnvConfig` to ignore timeout environment variables
   - Update `TestMergeConfigs` to handle reduced configuration surface
   - Add tests for fixed config.json path loading

2. **Integration Tests**
   - Update configuration loading tests to use fixed path
   - Verify timeout values are always hardcoded regardless of config files
   - Test that applications work without CONFIG_PATH

### Test Data Updates

1. **Test JSON files**: Remove timeout and methods fields
2. **Test .env files**: Remove timeout and methods variables
3. **Mock configurations**: Update to reflect new structure

### Regression Testing

1. **Handler Tests**: Should pass unchanged (no configuration dependency)
2. **Middleware Tests**: Should pass unchanged (no configuration dependency)
3. **Integration Tests**: Update to reflect new configuration behavior

## Implementation Phases

### Phase 1: Update Configuration Structure
- Modify ServerConfig JSON tags
- Update GetDefaultConfig with hardcoded values
- Update LoadEnvConfig to ignore timeout variables

### Phase 2: Update Configuration Loading
- Modify main.go to use fixed config.json path
- Remove CONFIG_PATH processing
- Update configuration merging logic

### Phase 3: Update Configuration Files
- Update config.json to remove timeout/methods fields
- Update .env.example to remove timeout/methods variables

### Phase 4: Update Tests
- Modify all configuration tests
- Update test data files
- Verify all tests pass

### Phase 5: Validation
- Run full test suite
- Verify application behavior with various configuration scenarios
- Confirm backward compatibility for non-timeout settings