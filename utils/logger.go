package utils

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"config"
)

type LoggerConfig struct {
	Level         string
	Mode          string
	FilePath      string
	EnableRotate  bool
	MaxSizeMB     int
	MaxBackups    int
	MaxAgeDays    int
	Compress      bool
	BufferSize    int
	FlushInterval time.Duration
}

var (
	logger       *zap.Logger
	logBuffer    chan zapcore.Entry
	initOnce     sync.Once
	shutdownOnce sync.Once
	cancelFunc   context.CancelFunc
	mu           sync.RWMutex

	cfg       *config.AppConfig
)

// customLevelEncoder adds color to log levels for console output
func customLevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	var colorCode string
	switch l {
	case zapcore.DebugLevel:
		colorCode = "\033[36m" // Cyan
	case zapcore.InfoLevel:
		colorCode = "\033[32m" // Green
	case zapcore.WarnLevel:
		colorCode = "\033[33m" // Yellow
	case zapcore.ErrorLevel:
		colorCode = "\033[31m" // Red
	case zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		colorCode = "\033[35m" // Magenta
	default:
		colorCode = "\033[0m" // Reset
	}
	enc.AppendString(fmt.Sprintf("%s[ %s ]\033[0m", colorCode, strings.ToUpper(l.String())))
}

// InitLogger sets up the global logger (safe singleton)
func InitLogger(cfg LoggerConfig) error {
	var initErr error
	initOnce.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		cancelFunc = cancel

		// Determine log level
		level := zap.InfoLevel
		if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
			level = zap.InfoLevel
		}

		// Build zap core
		var ws zapcore.WriteSyncer
		var useColor bool

		if cfg.FilePath != "" {
			// File logger with optional rotation
			lj := &lumberjack.Logger{
				Filename:   cfg.FilePath,
				MaxSize:    cfg.MaxSizeMB,
				MaxBackups: cfg.MaxBackups,
				MaxAge:     cfg.MaxAgeDays,
				Compress:   cfg.Compress,
			}
			ws = zapcore.AddSync(lj)
			useColor = false // disable color in files
		} else {
			// Console logger
			ws = zapcore.Lock(os.Stderr)
			useColor = true
		}

		encoderCfg := zap.NewProductionEncoderConfig()
		if cfg.Mode == "dev" {
			encoderCfg = zap.NewDevelopmentEncoderConfig()
		}
		encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

		// Color coding for console only
		if useColor {
			encoderCfg.EncodeLevel = customLevelEncoder
		} else {
			encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder
		}

		core := zapcore.NewCore(zapcore.NewConsoleEncoder(encoderCfg), ws, level)
		logger = zap.New(core, zap.AddCaller())

		// Set up buffered channel
		if cfg.BufferSize <= 0 {
			cfg.BufferSize = 1000
		}
		logBuffer = make(chan zapcore.Entry, cfg.BufferSize)

		// Start background flush worker
		go flushWorker(ctx, cfg.FlushInterval, core)

		// Setup signal handler
		go handleSignals()
	})
	return initErr
}

func flushWorker(ctx context.Context, interval time.Duration, core zapcore.Core) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			flushAll(core)
			return
		case <-ticker.C:
			flushAll(core)
		}
	}
}

func flushAll(core zapcore.Core) {
	for {
		select {
		case entry := <-logBuffer:
			if err := core.Write(entry, nil); err != nil {
				fmt.Fprintf(os.Stderr, "flush error: %v\n", err)
			}
		default:
			return
		}
	}
}

func handleSignals() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	for sig := range sigCh {
		fmt.Fprintf(os.Stderr, "Received signal: %v, shutting down logger\n", sig)
		ShutdownLogger()
		return
	}
}

// ShutdownLogger flushes and cleans up resources
func ShutdownLogger() {
	shutdownOnce.Do(func() {
		if cancelFunc != nil {
			cancelFunc()
		}
		if logger != nil {
			_ = logger.Sync()
		}
	})
}

// Safe accessor
func Logger() *zap.Logger {
	mu.RLock()
	defer mu.RUnlock()
	return logger
}

// Async logging API
func Log(level zapcore.Level, msg string, allowedModes []string, fields ...zap.Field) {
    if logger == nil {
        return
    }

    // Read current mode from environment
    currentMode := strings.ToLower(cfg.LogConfig.LogMode)

    allowed := false
    for _, m := range allowedModes {
        if strings.ToLower(m) == currentMode {
            allowed = true
            break
        }
    }
    if !allowed {
        return 
    }

    entry := zapcore.Entry{
        Level:   level,
        Time:    time.Now(),
        Message: msg,
    }
    select {
    case logBuffer <- entry:
    default:
        fmt.Fprintf(os.Stderr, "log buffer full, dropping log: %s\n", msg)
    }
}
