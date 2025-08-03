package utils

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// SuccessResponse defines the structure for successful JSON responses
type SuccessResponse struct {
	Status    string      `json:"status"`              
	Message   string      `json:"message"`            
	Data      interface{} `json:"data,omitempty"`      
	Meta      interface{} `json:"meta,omitempty"`      
	Timestamp string      `json:"timestamp"`           
}

var successLogger *zap.Logger

// InitSuccessLogger sets the logger used by success responses
func InitSuccessLogger(l *zap.Logger) {
	if l != nil {
		successLogger = l
	} else {
		// Create a no-op logger if nil is passed (consistent with error handler)
		successLogger = zap.NewNop()
	}
}

// getSuccessLogger returns the global success logger or a no-op logger if not initialized
func getSuccessLogger() *zap.Logger {
	if successLogger != nil {
		return successLogger
	}
	return zap.NewNop()
}

// getClientIP extracts the real client IP from request headers (shared with error handler)
func getSuccessClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to remote address
	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return ip
	}

	return r.RemoteAddr
}

// validateSuccessMessage ensures message is not empty
func validateSuccessMessage(message string) string {
	if strings.TrimSpace(message) == "" {
		return "Operation completed successfully"
	}
	return strings.TrimSpace(message)
}

// WriteSuccessWithContext writes a structured success response with context support and enhanced logging
func WriteSuccessWithContext(ctx context.Context, w http.ResponseWriter, r *http.Request, statusCode int, message string, data interface{}, meta interface{}) {
	// Validate inputs
	message = validateSuccessMessage(message)

	response := SuccessResponse{
		Status:    "success",
		Message:   message,
		Data:      data,
		Meta:      meta,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	// Enhanced logging with additional context
	logFields := []zap.Field{
		zap.Int("status", statusCode),
		zap.String("message", message),
		zap.String("method", r.Method),
		zap.String("url", r.URL.Path),
		zap.String("client_ip", getSuccessClientIP(r)),
		zap.String("user_agent", r.UserAgent()),
	}

	if data != nil {
		logFields = append(logFields, zap.Any("data_type", getDataType(data)))
	}

	if meta != nil {
		logFields = append(logFields, zap.Any("meta", meta))
	}

	// Add request ID if available in context
	if reqID := ctx.Value("request_id"); reqID != nil {
		logFields = append(logFields, zap.Any("request_id", reqID))
	}

	// Check for context cancellation
	if ctx.Err() != nil {
		logFields = append(logFields, zap.Error(ctx.Err()))
	}

	// Set headers and write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	l := getSuccessLogger()

	// Handle JSON encoding failures gracefully
	if encErr := json.NewEncoder(w).Encode(response); encErr != nil {
		l.Error("Failed to encode success response",
			zap.Error(encErr),
			zap.Int("original_status", statusCode),
			zap.String("original_message", message),
		)

		// Fallback: write a simple success message
		w.Header().Set("Content-Type", "text/plain")
		if _, writeErr := w.Write([]byte("Operation completed successfully")); writeErr != nil {
			l.Error("Failed to write fallback success response", zap.Error(writeErr))
		}
		return
	}

	// Log successful response
	l.Info("Success response sent", logFields...)
}

// WriteSuccess writes a structured success response (backward compatibility)
func WriteSuccess(w http.ResponseWriter, r *http.Request, statusCode int, message string, data interface{}, meta interface{}) {
	WriteSuccessWithContext(context.Background(), w, r, statusCode, message, data, meta)
}

// getDataType returns a string representation of the data type for logging
func getDataType(data interface{}) string {
	if data == nil {
		return "nil"
	}
	
	switch data.(type) {
	case string:
		return "string"
	case int, int8, int16, int32, int64:
		return "integer"
	case float32, float64:
		return "float"
	case bool:
		return "boolean"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	default:
		return "unknown"
	}
}

// Enhanced convenience functions with context support

// 2xx Success responses
func WriteOKWithContext(ctx context.Context, w http.ResponseWriter, r *http.Request, message string, data interface{}) {
	WriteSuccessWithContext(ctx, w, r, http.StatusOK, message, data, nil)
}

func WriteCreatedWithContext(ctx context.Context, w http.ResponseWriter, r *http.Request, message string, data interface{}) {
	WriteSuccessWithContext(ctx, w, r, http.StatusCreated, message, data, nil)
}

func WriteAcceptedWithContext(ctx context.Context, w http.ResponseWriter, r *http.Request, message string, data interface{}) {
	WriteSuccessWithContext(ctx, w, r, http.StatusAccepted, message, data, nil)
}

func WriteNoContentWithContext(ctx context.Context, w http.ResponseWriter, r *http.Request, message string) {
	WriteSuccessWithContext(ctx, w, r, http.StatusNoContent, message, nil, nil)
}

func WritePartialContentWithContext(ctx context.Context, w http.ResponseWriter, r *http.Request, message string, data interface{}, meta interface{}) {
	WriteSuccessWithContext(ctx, w, r, http.StatusPartialContent, message, data, meta)
}

// Convenience functions with pagination support
func WriteOKWithPaginationContext(ctx context.Context, w http.ResponseWriter, r *http.Request, message string, data interface{}, pagination interface{}) {
	WriteSuccessWithContext(ctx, w, r, http.StatusOK, message, data, pagination)
}

func WriteCreatedWithMetaContext(ctx context.Context, w http.ResponseWriter, r *http.Request, message string, data interface{}, meta interface{}) {
	WriteSuccessWithContext(ctx, w, r, http.StatusCreated, message, data, meta)
}

// Backward compatibility convenience functions (without context)
func WriteOK(w http.ResponseWriter, r *http.Request, message string, data interface{}) {
	WriteOKWithContext(context.Background(), w, r, message, data)
}

func WriteCreated(w http.ResponseWriter, r *http.Request, message string, data interface{}) {
	WriteCreatedWithContext(context.Background(), w, r, message, data)
}

func WriteAccepted(w http.ResponseWriter, r *http.Request, message string, data interface{}) {
	WriteAcceptedWithContext(context.Background(), w, r, message, data)
}

func WriteNoContent(w http.ResponseWriter, r *http.Request, message string) {
	WriteNoContentWithContext(context.Background(), w, r, message)
}

func WritePartialContent(w http.ResponseWriter, r *http.Request, message string, data interface{}, meta interface{}) {
	WritePartialContentWithContext(context.Background(), w, r, message, data, meta)
}

func WriteOKWithPagination(w http.ResponseWriter, r *http.Request, message string, data interface{}, pagination interface{}) {
	WriteOKWithPaginationContext(context.Background(), w, r, message, data, pagination)
}

func WriteCreatedWithMeta(w http.ResponseWriter, r *http.Request, message string, data interface{}, meta interface{}) {
	WriteCreatedWithMetaContext(context.Background(), w, r, message, data, meta)
}

// Specialized success handlers for common API patterns
func WriteResourceCreated(w http.ResponseWriter, r *http.Request, resourceType string, data interface{}) {
	message := resourceType + " created successfully"
	WriteCreated(w, r, message, data)
}

func WriteResourceUpdated(w http.ResponseWriter, r *http.Request, resourceType string, data interface{}) {
	message := resourceType + " updated successfully"
	WriteOK(w, r, message, data)
}

func WriteResourceDeleted(w http.ResponseWriter, r *http.Request, resourceType string) {
	message := resourceType + " deleted successfully"
	WriteNoContent(w, r, message)
}

func WriteResourceList(w http.ResponseWriter, r *http.Request, resourceType string, data interface{}, pagination interface{}) {
	message := resourceType + " retrieved successfully"
	WriteOKWithPagination(w, r, message, data, pagination)
}

func WriteResourceDetail(w http.ResponseWriter, r *http.Request, resourceType string, data interface{}) {
	message := resourceType + " retrieved successfully"
	WriteOK(w, r, message, data)
}