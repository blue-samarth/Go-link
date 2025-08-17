# Config Package Test Suite Documentation

## Overview

This test suite provides comprehensive coverage for the `config` package, which handles application configuration loading, validation, and logging. The tests ensure configuration is properly loaded from environment variables and `.env` files, with appropriate validation and error handling.

## Test Reference Table

| Test Name | Target Function | Test Type | Purpose | Setup Requirements | Success Criteria | Common Failure Reasons | Fix Strategy |
|-----------|-----------------|-----------|---------|-------------------|------------------|----------------------|--------------|
| `TestLoadConfig_DefaultsOnly` | `LoadConfig()` | Unit/Parallel | Verify default values applied correctly | `DB_TYPE=postgres`<br>`JWT_SECRET_KEY=test-key`<br>`REFRESH_TOKEN_SECRET=test-key`<br>`SKIP_PORT_CHECK=true` | Config non-nil<br>Port in range 1-65535<br>DB type matches env | Missing env vars<br>Invalid defaults<br>Config not initialized | Check env var setup<br>Verify default constants<br>Debug config initialization |
| `TestLoadConfig_LoadFromEnvFile` | `LoadConfig()` | Sequential | Test .env file loading | Create .env file with config<br>Clear conflicting env vars | Values from .env loaded<br>DB config created | .env file not found<br>Env vars override .env<br>File parsing errors | Check file creation<br>Clear env vars first<br>Validate .env format |
| `TestLoadConfig_LoadFromEnvVars` | `LoadConfig()` | Sequential | Test env var precedence | Set env vars directly<br>No .env file | Env vars correctly parsed<br>Appropriate DB config created | Type conversion failures<br>Missing DB config<br>Parsing errors | Verify data types<br>Check DB type logic<br>Debug parsing |
| `TestLoadConfig_InvalidDBType` | `LoadConfig()` | Subprocess | Ensure invalid DB types fatal | `DB_TYPE=foobar`<br>Required secrets set | Subprocess exits non-zero | Subprocess doesn't exit<br>Zero exit code<br>Process start failure | Check validation logic<br>Verify log.Fatalf calls<br>Test subprocess setup |
| `TestLoadConfig_ValidationFailure` | `LoadConfig()` | Subprocess | Test production validation | `ENV=production`<br>`JWT_SECRET_KEY=short` | Subprocess exits non-zero | Validation not enforced<br>Environment detection broken<br>Secret validation bypassed | Review validation rules<br>Check env detection<br>Verify secret length checks |
| `TestLoadConfig_DbConfigRequired` | `LoadConfig()` | Unit/Parallel | Verify PostgreSQL config creation | `DB_TYPE=postgres`<br>Required env vars | `DbConfig` is non-nil | Config object not created<br>Wrong DB type detection | Check DB type switching<br>Verify config initialization |
| `TestLoadConfig_MongoConfigRequired` | `LoadConfig()` | Unit/Parallel | Verify MongoDB config creation | `DB_TYPE=mongodb`<br>Required env vars | `MongoConfig` is non-nil | Config object not created<br>DB type logic broken | Check DB type switching<br>Verify MongoDB initialization |
| `TestLoadConfig_LoggingRotationFileRequired` | `LoadConfig()` | Subprocess | Test log rotation validation | `ENABLE_ROTATION=true`<br>No `LOG_FILE` set | Subprocess exits non-zero | Validation not enforced<br>File requirement bypassed | Check rotation validation<br>Verify file requirement logic |
| `TestLoadConfig_LoggingRotationSizeInvalid` | `LoadConfig()` | Subprocess | Test rotation size validation | `ROTATION_MAX_SIZE=0` | Subprocess exits non-zero | Size validation bypassed<br>Invalid values accepted | Review size validation<br>Check boundary conditions |
| `TestLoadConfig_LoggingFlushIntervalInvalid` | `LoadConfig()` | Subprocess | Test flush interval validation | `LOG_FLUSH_INTERVAL=0` | Subprocess exits non-zero | Interval validation bypassed<br>Zero values accepted | Check interval validation<br>Verify positive number requirement |
| `TestLoadConfig_SkipPortAvailabilityCheck` | `LoadConfig()` | Unit/Parallel | Test port check bypass | `SKIP_PORT_CHECK=true`<br>`APP_PORT=65530` | Config loads successfully<br>Port correctly set | Port checking not bypassed<br>High port rejection | Verify skip logic<br>Check port validation bypass |
| `TestLoadConfig_CleanenvFallbackOnInvalidEnv` | `LoadConfig()` | Unit/Parallel | Test valid env parsing | All valid env vars set | Config loads with correct values<br>All values parsed | Parsing errors<br>Type conversion issues<br>Value assignment failures | Check cleanenv integration<br>Verify data types<br>Debug parsing logic |
| `TestLogConfigValues_SimpleStructLogging` | `LogConfigValues()` | Unit/Parallel | Test basic field logging | Simple struct with exported fields | All fields in output<br>Correct value formatting | Fields not logged<br>Formatting issues<br>Missing field names | Check reflection logic<br>Verify log formatting<br>Debug field iteration |
| `TestLogConfigValues_SecretMasking` | `LogConfigValues()` | Unit/Parallel | Test secret field masking | Struct with `secret:"true"` tag | Non-secrets logged normally<br>Secrets masked with asterisks<br>Actual secrets not in output | Secrets not masked<br>Complete omission<br>Masking logic broken | Verify tag detection<br>Check masking implementation<br>Test security compliance |
| `TestLogConfigValues_NestedStructLogging` | `LogConfigValues()` | Unit/Parallel | Test nested struct logging | Struct with nested struct fields | Dot notation paths<br>All nested values logged | Nesting not traversed<br>Path format incorrect<br>Missing values | Check recursion logic<br>Verify path construction<br>Debug nested traversal |
| `TestLogConfigValues_PointerStructLogging` | `LogConfigValues()` | Unit/Parallel | Test pointer dereferencing | Struct with pointer to nested struct | Same output as value structs<br>Correct dereferencing | Nil pointer panics<br>Values not followed<br>Memory addresses logged | Add nil checks<br>Fix pointer following<br>Handle dereferencing safely |
| `TestLogConfigValues_NilPointerField` | `LogConfigValues()` | Unit/Parallel | Test nil pointer handling | Struct with nil pointer field | Non-nil fields logged<br>Nil fields skipped gracefully | Nil pointer panics<br>Error messages in output<br>Test crashes | Add nil checking<br>Handle gracefully<br>Prevent panics |
| `TestLogConfigValues_UnexportedFields` | `LogConfigValues()` | Unit/Parallel | Test field visibility filtering | Struct with private fields | Exported fields logged<br>Private fields ignored | Private fields logged<br>Security risk<br>Reflection filtering broken | Fix field visibility check<br>Ensure security<br>Test reflection filters |
| `TestLogConfigValues_DeeplyNestedConfig` | `LogConfigValues()` | Unit/Parallel | Test deep nesting handling | Multi-level nested structs | Full path notation<br>All levels logged<br>Acceptable performance | Stack overflow<br>Path construction errors<br>Performance issues | Limit recursion depth<br>Optimize path building<br>Test performance |

## Test Architecture

### Test Categories

The test suite is organized into two main categories:

1. **Configuration Loading Tests** (`TestLoadConfig_*`) - Test the `LoadConfig()` function
2. **Configuration Logging Tests** (`TestLogConfigValues_*`) - Test the `LogConfigValues()` function

### Test Isolation Strategy

- **Parallel Tests**: Tests that don't modify global state (environment variables, files) run in parallel using `t.Parallel()`
- **Sequential Tests**: Tests that create `.env` files or modify shared environment state run sequentially
- **Subprocess Tests**: Tests that verify `log.Fatalf` behavior run in isolated subprocesses to prevent test runner termination

## Test Helper Functions

### `cleanEnv(t *testing.T)`
Resets all configuration-related environment variables and removes any `.env` file to ensure test isolation.

**Usage**: Call at the beginning of each test to ensure clean state.

### `createEnvFile(t *testing.T, content string)`
Creates a `.env` file with the specified content for testing file-based configuration loading.

**Cleanup**: Automatically removes `.env` file after test completion.

### `expectFatal(t *testing.T, fn func()) (bool, string)`
Captures log output and detects if a function would cause a fatal error without actually terminating the test.

**Returns**: 
- `didPanic`: Whether a panic occurred
- `panicMsg`: The panic message or log output containing fatal/error keywords

### `captureLogOutput(fn func()) string`
Captures and returns log output from a function execution for assertion purposes.

## Configuration Loading Tests

### Core Functionality Tests

#### `TestLoadConfig_DefaultsOnly`
**Purpose**: Verifies that default configuration values are properly applied when only required environment variables are set.

**Setup**:
```bash
DB_TYPE=postgres
JWT_SECRET_KEY=test-jwt-secret-key-for-testing-only-32-chars
REFRESH_TOKEN_SECRET=test-refresh-secret-key-for-testing-only-32-chars
SKIP_PORT_CHECK=true
```

**Assertions**:
- Config object is non-nil
- Default port is within valid range (1-65535)
- DB type matches environment variable

**Common Failures**:
- Missing required environment variables
- Invalid default port configuration
- Configuration object not initialized

#### `TestLoadConfig_LoadFromEnvFile`
**Purpose**: Tests loading configuration from `.env` file when `LoadConfig(false)` is called.

**Setup**: Creates `.env` file with configuration values and ensures no conflicting environment variables exist.

**Assertions**:
- Values loaded from `.env` file correctly
- Database configuration object created for specified DB type

**Common Failures**:
- `.env` file not found or malformed
- Environment variables overriding `.env` file values
- Incorrect parsing of `.env` file format

#### `TestLoadConfig_LoadFromEnvVars`
**Purpose**: Verifies that environment variables take precedence and are correctly parsed.

**Setup**: Sets environment variables directly without `.env` file.

**Assertions**:
- Environment variables correctly parsed
- Appropriate database configuration object created

**Common Failures**:
- Environment variable parsing errors
- Type conversion failures
- Missing database-specific configuration

### Validation Tests

#### `TestLoadConfig_InvalidDBType`
**Purpose**: Ensures that invalid database types cause application termination via `log.Fatalf`.

**Method**: Subprocess isolation - runs the failing code in a separate process.

**Setup**:
```bash
DB_TYPE=foobar  # Invalid value
```

**Expected Behavior**: Subprocess exits with non-zero code.

**Common Failures**:
- Subprocess doesn't exit (validation not working)
- Exit code is 0 (successful exit instead of error)
- Process doesn't start (test environment issues)

#### `TestLoadConfig_ValidationFailure`
**Purpose**: Tests that configuration validation failures (e.g., short JWT secret in production) cause application termination.

**Method**: Subprocess isolation.

**Setup**:
```bash
ENV=production
JWT_SECRET_KEY=short  # Too short for production environment
```

**Expected Behavior**: Subprocess exits with non-zero code due to validation failure.

**Common Failures**:
- Validation rules not enforced
- Production environment detection not working
- Secret key length validation bypassed

### Database Configuration Tests

#### `TestLoadConfig_DbConfigRequired`
**Purpose**: Verifies that PostgreSQL configuration object is created when `DB_TYPE=postgres`.

**Assertions**:
- `cfg.DbConfig` is non-nil
- Database type correctly set

**Common Failures**:
- Database configuration object not initialized
- Wrong database type detection

#### `TestLoadConfig_MongoConfigRequired`
**Purpose**: Verifies that MongoDB configuration object is created when `DB_TYPE=mongodb`.

**Assertions**:
- `cfg.MongoConfig` is non-nil
- Database type correctly set

**Common Failures**:
- MongoDB configuration object not initialized
- Database type switching logic broken

### Logging Configuration Tests

#### `TestLoadConfig_LoggingRotationFileRequired`
**Purpose**: Ensures that enabling log rotation without specifying a log file causes validation failure.

**Method**: Subprocess isolation.

**Setup**:
```bash
ENABLE_ROTATION=true
# LOG_FILE not set - should cause failure
```

**Expected Behavior**: Subprocess exits with non-zero code.

**Common Failures**:
- Log rotation validation not enforced
- Missing file validation bypassed

#### `TestLoadConfig_LoggingRotationSizeInvalid`
**Purpose**: Tests validation of log rotation size parameter.

**Method**: Subprocess isolation.

**Setup**:
```bash
ROTATION_MAX_SIZE=0  # Invalid size
```

**Expected Behavior**: Subprocess exits with non-zero code.

#### `TestLoadConfig_LoggingFlushIntervalInvalid`
**Purpose**: Tests validation of log flush interval parameter.

**Method**: Subprocess isolation.

**Setup**:
```bash
LOG_FLUSH_INTERVAL=0  # Invalid interval
```

**Expected Behavior**: Subprocess exits with non-zero code.

### Network Configuration Tests

#### `TestLoadConfig_SkipPortAvailabilityCheck`
**Purpose**: Verifies that port availability checking can be bypassed for testing.

**Setup**:
```bash
SKIP_PORT_CHECK=true
APP_PORT=65530  # High port number
```

**Assertions**:
- Configuration loads successfully despite potentially unavailable port
- Port number correctly set

**Common Failures**:
- Port checking not properly bypassed
- Port validation still enforced

#### `TestLoadConfig_CleanenvFallbackOnInvalidEnv`
**Purpose**: Tests that configuration loading handles environment parsing gracefully.

**Setup**: Valid environment variables to ensure successful parsing.

**Assertions**:
- Configuration loads with valid environment values
- All values correctly parsed and assigned

## Configuration Logging Tests

### Basic Logging Tests

#### `TestLogConfigValues_SimpleStructLogging`
**Purpose**: Tests basic struct field logging functionality.

**Test Structure**:
```go
testStruct := struct {
    AppPort int
    DbType  string
}{
    AppPort: 8080,
    DbType:  "postgres",
}
```

**Assertions**:
- All exported fields appear in log output
- Values are correctly formatted

**Common Failures**:
- Fields not logged
- Incorrect value formatting
- Missing field names in output

#### `TestLogConfigValues_SecretMasking`
**Purpose**: Ensures that fields tagged with `secret:"true"` are masked in log output.

**Test Structure**:
```go
testStruct := struct {
    Username string
    Password string `secret:"true"`
}{
    Username: "admin",
    Password: "secret123",
}
```

**Assertions**:
- Non-secret fields logged normally
- Secret fields present but masked (contains asterisks)
- Actual secret values not present in output

**Common Failures**:
- Secret values not masked (security risk)
- Secret fields completely omitted
- Masking mechanism not working

### Nested Structure Tests

#### `TestLogConfigValues_NestedStructLogging`
**Purpose**: Tests logging of nested struct configurations.

**Expected Output Format**:
```
AppPort: 8080
DbConfig.Host: localhost
DbConfig.Port: 5432
```

**Assertions**:
- Nested fields use dot notation for field paths
- All nested values correctly logged

**Common Failures**:
- Nested structures not traversed
- Incorrect field path formatting
- Missing nested values

#### `TestLogConfigValues_PointerStructLogging`
**Purpose**: Tests logging when nested structs are accessed via pointers.

**Assertions**:
- Pointer dereferencing works correctly
- Same output format as value-based nested structs

**Common Failures**:
- Nil pointer dereference panics
- Pointer values not followed
- Memory address logged instead of values

#### `TestLogConfigValues_NilPointerField`
**Purpose**: Ensures that nil pointer fields are handled gracefully.

**Assertions**:
- Non-nil fields logged normally
- Nil pointer fields skipped without errors

**Common Failures**:
- Nil pointer dereference causing panics
- Error messages in log output
- Test crashes due to unhandled nil pointers

### Advanced Logging Tests

#### `TestLogConfigValues_UnexportedFields`
**Purpose**: Verifies that unexported (private) fields are ignored during logging.

**Test Structure**:
```go
testStruct := struct {
    AppPort    int      // Exported - should appear
    privateKey string   // Unexported - should be ignored
}
```

**Assertions**:
- Exported fields logged normally
- Unexported fields not present in output

**Common Failures**:
- Private fields accidentally logged (potential security issue)
- Reflection not properly filtering field visibility

#### `TestLogConfigValues_DeeplyNestedConfig`
**Purpose**: Tests logging of deeply nested configuration structures.

**Expected Output**:
```
Level1.Level2.Level3.Value: deep
```

**Assertions**:
- Deep nesting handled correctly
- Full path notation maintained
- Performance acceptable for deep structures

**Common Failures**:
- Stack overflow on deep nesting
- Incorrect path construction
- Performance issues with reflection

## Troubleshooting Guide

### Common Test Failures

#### Environment Variable Issues
**Symptoms**: Tests fail with missing configuration or default values applied incorrectly.

**Solutions**:
1. Verify `cleanEnv()` is called at test start
2. Check that required environment variables are set
3. Ensure environment variable names match exactly
4. Verify no conflicting variables from previous tests

#### Subprocess Test Failures
**Symptoms**: Tests expecting `log.Fatalf` behavior pass when they should fail, or vice versa.

**Solutions**:
1. Check that `TEST_SUBPROCESS` environment variable is properly set
2. Verify subprocess is running the correct test function
3. Ensure parent process correctly interprets exit codes
4. Check for subprocess execution permissions

#### File System Issues
**Symptoms**: Tests involving `.env` files fail intermittently.

**Solutions**:
1. Verify test has write permissions in working directory
2. Check for file system race conditions
3. Ensure `.env` file cleanup is working properly
4. Verify no concurrent tests are modifying the same files

#### Parallel Execution Issues
**Symptoms**: Tests fail when run in parallel but pass individually.

**Solutions**:
1. Review which tests are marked with `t.Parallel()`
2. Check for shared state between parallel tests
3. Verify environment variable isolation
4. Consider removing `t.Parallel()` for problematic tests

### Debugging Tips

#### Log Output Analysis
When tests fail due to unexpected log content:
1. Examine the captured log output string
2. Check for whitespace or formatting differences
3. Verify expected keywords are present
4. Look for additional unexpected log entries

#### Configuration Object Inspection
For configuration loading failures:
1. Add debug prints of the loaded configuration
2. Check each field value against expectations
3. Verify correct data types and conversions
4. Examine nested object initialization

#### Subprocess Debugging
For subprocess test issues:
1. Run the failing command manually
2. Check stderr output for additional error information
3. Verify exit codes match expectations
4. Test subprocess isolation by running tests individually

## Test Maintenance

### Adding New Tests

When adding new configuration fields or validation rules:

1. **Add Unit Tests**: Create tests for both valid and invalid configurations
2. **Update Validation Tests**: Add subprocess tests for new validation failures
3. **Update Logging Tests**: Add tests for new fields in logging output
4. **Consider Edge Cases**: Test nil values, empty strings, and boundary conditions

### Modifying Existing Tests

When changing configuration behavior:

1. **Update Expected Values**: Modify test assertions to match new behavior
2. **Review Test Categories**: Ensure tests are still properly categorized
3. **Check Parallel Safety**: Verify parallel execution is still safe
4. **Update Documentation**: Reflect changes in this documentation

### Performance Considerations

- **Subprocess Overhead**: Minimize subprocess tests as they have higher execution time
- **File I/O**: Limit `.env` file operations to necessary tests only
- **Parallel Execution**: Maximize parallel test execution for faster runs
- **Resource Cleanup**: Ensure all tests properly clean up resources

## Integration with CI/CD

### Required Environment
- Go 1.19 or later
- Write permissions for `.env` file creation
- Ability to execute subprocess commands
- Access to environment variable modification

### Recommended Test Execution
```bash
# Run all tests with race detection
go test -race -v ./config

# Run specific test category
go test -run="TestLoadConfig_" -v ./config

# Run with coverage report
go test -cover -coverprofile=coverage.out ./config
```

### CI/CD Considerations
- Ensure clean environment for each test run
- Consider test timeout values for subprocess tests
- Monitor test execution time for performance regressions
- Include both sequential and parallel test execution in pipeline