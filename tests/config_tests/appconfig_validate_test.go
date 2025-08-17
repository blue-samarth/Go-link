package tests

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/blue-samarth/go-link/config"
)

// Test helper for environment variable isolation
func setEnvForTest(t *testing.T, key, value string) {
	t.Helper()
	
	// Save original value (if any)
	originalValue, hadOriginal := os.LookupEnv(key)
	
	// Set new value
	if value == "" {
		os.Unsetenv(key)
	} else {
		os.Setenv(key, value)
	}
	
	// Restore on cleanup
	t.Cleanup(func() {
		if hadOriginal {
			os.Setenv(key, originalValue)
		} else {
			os.Unsetenv(key)
		}
	})
}

// Helper to generate secrets of specific length for testing
func makeSecret(length int) string {
	return strings.Repeat("x", length)
}

// Helper to create a minimal valid config for testing
func createValidBaseConfig() *config.AppConfig {
    return &config.AppConfig{
        AppPort: 8080,
        DbType:  config.SQLite,
        DbConfig: &config.DbConfig{DbPort: 5432}, // minimal config, since Validate() demands it
        Env:     "development",
        JWTSecretKey:       makeSecret(32),
        RefreshTokenSecret: makeSecret(32),
        LogConfig: config.LogConfig{
            LogFlushInterval: 1000,
            RotationMaxSize:  100,
            EnableRotation:   false,
            LogFile:          "",
        },
    }
}


func TestAppConfig_Validate_SucceedsWithValidConfiguration(t *testing.T) {
	t.Parallel()
	setEnvForTest(t, "SKIP_PORT_CHECK", "true")
	
	cfg := createValidBaseConfig()
	err := cfg.Validate()
	
	require.NoError(t, err, "valid configuration should pass validation")
}

func TestAppConfig_Validate_PortValidation(t *testing.T) {
	t.Parallel()
	setEnvForTest(t, "SKIP_PORT_CHECK", "true")
	
	testCases := []struct {
		name    string
		port    int
		wantErr bool
		errText string
	}{
		{"minimum valid port", 1, false, ""},
		{"maximum valid port", 65535, false, ""},
		{"below valid range", 0, true, "APP_PORT"},
		{"above valid range", 65536, true, "APP_PORT"},
		{"negative port", -1, true, "APP_PORT"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := createValidBaseConfig()
			cfg.AppPort = tc.port
			
			err := cfg.Validate()
			if tc.wantErr {
				require.Error(t, err, "APP_PORT=%d should fail validation", tc.port)
				require.Contains(t, err.Error(), tc.errText)
			} else {
				require.NoError(t, err, "APP_PORT=%d should be valid", tc.port)
			}
		})
	}
}

func TestAppConfig_Validate_FailsWithUnsupportedDbType(t *testing.T) {
	t.Parallel()
	setEnvForTest(t, "SKIP_PORT_CHECK", "true")
	
	cfg := createValidBaseConfig()
	cfg.DbType = config.DbType("Oracle") // Unsupported
	
	err := cfg.Validate()
	require.Error(t, err, "unsupported DbType should fail validation")
	require.Contains(t, err.Error(), "DB_TYPE")
}

func TestAppConfig_Validate_DatabaseConfigValidation(t *testing.T) {
	t.Parallel()
	setEnvForTest(t, "SKIP_PORT_CHECK", "true")
	
	testCases := []struct {
		name        string
		dbType      config.DbType
		dbConfig    *config.DbConfig
		mongoConfig *config.MongoConfig
		wantErr     bool
		errText     string
	}{
		{"Postgres missing DbConfig", config.Postgres, nil, nil, true, "DbConfig must not be nil"},
		{"MySQL missing DbConfig", config.MySQL, nil, nil, true, "DbConfig must not be nil"},
		{"MsSQL missing DbConfig", config.MsSQL, nil, nil, true, "DbConfig must not be nil"},
		{"MongoDB missing MongoConfig", config.MongoDb, nil, nil, true, "MongoConfig must not be nil"},
		{"SQLite missing SQLiteConfig", config.SQLite, nil, nil, true, "DbConfig must not be nil for database type sqlite"},
		{"Postgres with valid DbConfig", config.Postgres, &config.DbConfig{DbPort: 5432}, nil, false, ""},
		{"MongoDB with valid MongoConfig", config.MongoDb, nil, &config.MongoConfig{Port: 27017}, false, ""},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := createValidBaseConfig()
			cfg.DbType = tc.dbType
			cfg.DbConfig = tc.dbConfig
			cfg.MongoConfig = tc.mongoConfig
			
			err := cfg.Validate()
			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errText)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAppConfig_Validate_ProductionSecretValidation(t *testing.T) {
	t.Parallel()
	setEnvForTest(t, "SKIP_PORT_CHECK", "true")
	
	testCases := []struct {
		name          string
		jwtSecret     string
		refreshSecret string
		wantErr       bool
		errText       string
	}{
		{"weak JWT secret", "secret", makeSecret(32), true, "JWT_SECRET_KEY"},
		{"short JWT secret", "short", makeSecret(32), true, "JWT_SECRET_KEY"},
		{"weak refresh secret", makeSecret(32), "refresh_secret", true, "REFRESH_TOKEN_SECRET"},
		{"short refresh secret", makeSecret(32), "short", true, "REFRESH_TOKEN_SECRET"},
		{"valid secrets", makeSecret(32), makeSecret(32), false, ""},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := createValidBaseConfig()
			cfg.Env = "production"
			cfg.JWTSecretKey = tc.jwtSecret
			cfg.RefreshTokenSecret = tc.refreshSecret
			
			err := cfg.Validate()
			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errText)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAppConfig_Validate_LogConfigValidation(t *testing.T) {
	t.Parallel()
	setEnvForTest(t, "SKIP_PORT_CHECK", "true")
	
	testCases := []struct {
		name             string
		enableRotation   bool
		logFile          string
		rotationMaxSize  int
		logFlushInterval int
		wantErr          bool
		errText          string
	}{
		{"rotation enabled without log file", true, "", 100, 1000, true, "LOG_FILE must be set"},
		{"rotation enabled with log file", true, "/tmp/app.log", 100, 1000, false, ""},
		{"zero rotation max size", false, "", 0, 1000, true, "ROTATION_MAX_SIZE"},
		{"negative rotation max size", false, "", -1, 1000, true, "ROTATION_MAX_SIZE"},
		{"zero flush interval", false, "", 100, 0, true, "LOG_FLUSH_INTERVAL"},
		{"negative flush interval", false, "", 100, -1, true, "LOG_FLUSH_INTERVAL"},
		{"valid log config", false, "", 100, 1000, false, ""},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := createValidBaseConfig()
			cfg.LogConfig.EnableRotation = tc.enableRotation
			cfg.LogConfig.LogFile = tc.logFile
			cfg.LogConfig.RotationMaxSize = tc.rotationMaxSize
			cfg.LogConfig.LogFlushInterval = tc.logFlushInterval
			
			err := cfg.Validate()
			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errText)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAppConfig_Validate_SkipsPortCheckWhenEnvironmentVariableSet(t *testing.T) {
	t.Parallel()
	setEnvForTest(t, "SKIP_PORT_CHECK", "true")
	
	cfg := createValidBaseConfig()
	cfg.AppPort = 80 // System port, might be restricted
	
	err := cfg.Validate()
	require.NoError(t, err, "should skip port availability check when SKIP_PORT_CHECK=true")
}

func TestAppConfig_Validate_DatabasePortValidation(t *testing.T) {
	t.Parallel()
	setEnvForTest(t, "SKIP_PORT_CHECK", "true")
	
	testCases := []struct {
		name        string
		dbType      config.DbType
		port        int
		setupConfig func(*config.AppConfig, int)
		errText     string
	}{
		{
			"invalid Postgres port",
			config.Postgres,
			0,
			func(cfg *config.AppConfig, port int) {
				cfg.DbConfig = &config.DbConfig{DbPort: port}
			},
			"DB_PORT",
		},
		{
			"invalid MongoDB port",
			config.MongoDb,
			70000,
			func(cfg *config.AppConfig, port int) {
				cfg.MongoConfig = &config.MongoConfig{Port: port}
			},
			"MONGO_PORT",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := createValidBaseConfig()
			cfg.DbType = tc.dbType
			tc.setupConfig(cfg, tc.port)
			
			err := cfg.Validate()
			require.Error(t, err, "invalid database port should fail validation")
			require.Contains(t, err.Error(), tc.errText)
		})
	}
}