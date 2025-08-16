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

// ErrorResponse defines the structure for error JSON responses
type ErrorResponse struct {
	Status    string      `json:"status"`
	Code      string      `json:"code"`
	Message   string      `json:"message"`
	Details   interface{} `json:"details,omitempty"`
	Timestamp string      `json:"timestamp"`
}

// getClientIP extracts the real client IP from request headers
func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return ip
	}

	return r.RemoteAddr
}

func validateMessage(message string) string {
	if strings.TrimSpace(message) == "" {
		return "An error occurred"
	}
	return strings.TrimSpace(message)
}

// HandleErrorWithContext provides enhanced error handling and logging
func HandleErrorWithContext(ctx context.Context, w http.ResponseWriter, r *http.Request, status int, message string, err error, details ...interface{}) {
	message = validateMessage(message)

	var detail interface{}
	if len(details) == 1 {
		detail = details[0]
	} else if len(details) > 1 {
		detail = details
	}

	resp := ErrorResponse{
		Status:    "error",
		Code:      mapStatusToCode(status),
		Message:   message,
		Details:   detail,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	logFields := []zap.Field{
		zap.Int("status", status),
		zap.String("message", message),
		zap.String("method", r.Method),
		zap.String("url", r.URL.Path),
		zap.String("client_ip", getClientIP(r)),
		zap.String("user_agent", r.UserAgent()),
	}

	if detail != nil {
		logFields = append(logFields, zap.Any("details", detail))
	}

	if reqID := ctx.Value("request_id"); reqID != nil {
		logFields = append(logFields, zap.Any("request_id", reqID))
	}

	if ctx.Err() != nil {
		logFields = append(logFields, zap.Error(ctx.Err()))
	}

	l := Logger()
	if err != nil {
		logFields = append(logFields, zap.Error(err))
		l.Error("HTTP error", logFields...)
	} else {
		l.Warn("HTTP warning", logFields...)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if encErr := json.NewEncoder(w).Encode(resp); encErr != nil {
		l.Error("Failed to encode error response",
			zap.Error(encErr),
			zap.Int("original_status", status),
			zap.String("original_message", message),
		)
		w.Header().Set("Content-Type", "text/plain")
		if _, writeErr := w.Write([]byte("Internal server error: failed to encode response")); writeErr != nil {
			l.Error("Failed to write fallback error response", zap.Error(writeErr))
		}
	}
}

// Generic error handler (backward compatibility)
func HandleError(w http.ResponseWriter, r *http.Request, status int, message string, err error, details ...interface{}) {
	HandleErrorWithContext(context.Background(), w, r, status, message, err, details...)
}

// Enhanced shorthand handlers with context support
func HandleBadRequestWithContext(ctx context.Context, w http.ResponseWriter, r *http.Request, message string, err error, details ...interface{}) {
	HandleErrorWithContext(ctx, w, r, http.StatusBadRequest, message, err, details...)
}

func HandleUnauthorizedWithContext(ctx context.Context, w http.ResponseWriter, r *http.Request, message string, err error, details ...interface{}) {
	HandleErrorWithContext(ctx, w, r, http.StatusUnauthorized, message, err, details...)
}

func HandleForbiddenWithContext(ctx context.Context, w http.ResponseWriter, r *http.Request, message string, err error, details ...interface{}) {
	HandleErrorWithContext(ctx, w, r, http.StatusForbidden, message, err, details...)
}

func HandleNotFoundWithContext(ctx context.Context, w http.ResponseWriter, r *http.Request, message string, err error, details ...interface{}) {
	HandleErrorWithContext(ctx, w, r, http.StatusNotFound, message, err, details...)
}

func HandleConflictWithContext(ctx context.Context, w http.ResponseWriter, r *http.Request, message string, err error, details ...interface{}) {
	HandleErrorWithContext(ctx, w, r, http.StatusConflict, message, err, details...)
}

func HandleUnprocessableEntityWithContext(ctx context.Context, w http.ResponseWriter, r *http.Request, message string, err error, details ...interface{}) {
	HandleErrorWithContext(ctx, w, r, http.StatusUnprocessableEntity, message, err, details...)
}

func HandleTooManyRequestsWithContext(ctx context.Context, w http.ResponseWriter, r *http.Request, message string, err error, details ...interface{}) {
	HandleErrorWithContext(ctx, w, r, http.StatusTooManyRequests, message, err, details...)
}

func HandleInternalErrorWithContext(ctx context.Context, w http.ResponseWriter, r *http.Request, err error, details ...interface{}) {
	HandleErrorWithContext(ctx, w, r, http.StatusInternalServerError, "Internal server error", err, details...)
}

func HandleServiceUnavailableWithContext(ctx context.Context, w http.ResponseWriter, r *http.Request, message string, err error, details ...interface{}) {
	HandleErrorWithContext(ctx, w, r, http.StatusServiceUnavailable, message, err, details...)
}

// Backward compatibility shorthand handlers (without context)
func HandleBadRequest(w http.ResponseWriter, r *http.Request, message string, err error, details ...interface{}) {
	HandleBadRequestWithContext(context.Background(), w, r, message, err, details...)
}

func HandleUnauthorized(w http.ResponseWriter, r *http.Request, message string, err error, details ...interface{}) {
	HandleUnauthorizedWithContext(context.Background(), w, r, message, err, details...)
}

func HandleForbidden(w http.ResponseWriter, r *http.Request, message string, err error, details ...interface{}) {
	HandleForbiddenWithContext(context.Background(), w, r, message, err, details...)
}

func HandleNotFound(w http.ResponseWriter, r *http.Request, message string, err error, details ...interface{}) {
	HandleNotFoundWithContext(context.Background(), w, r, message, err, details...)
}

func HandleMethodNotAllowed(w http.ResponseWriter, r *http.Request, message string, err error, details ...interface{}) {
	HandleErrorWithContext(context.Background(), w, r, http.StatusMethodNotAllowed, message, err, details...)
}

func HandleConflict(w http.ResponseWriter, r *http.Request, message string, err error, details ...interface{}) {
	HandleConflictWithContext(context.Background(), w, r, message, err, details...)
}

func HandleUnprocessableEntity(w http.ResponseWriter, r *http.Request, message string, err error, details ...interface{}) {
	HandleUnprocessableEntityWithContext(context.Background(), w, r, message, err, details...)
}

func HandleTooManyRequests(w http.ResponseWriter, r *http.Request, message string, err error, details ...interface{}) {
	HandleTooManyRequestsWithContext(context.Background(), w, r, message, err, details...)
}

func HandleInternalError(w http.ResponseWriter, r *http.Request, err error, details ...interface{}) {
	HandleInternalErrorWithContext(context.Background(), w, r, err, details...)
}

func HandleServiceUnavailable(w http.ResponseWriter, r *http.Request, message string, err error, details ...interface{}) {
	HandleServiceUnavailableWithContext(context.Background(), w, r, message, err, details...)
}

func HandleRequestTimeout(w http.ResponseWriter, r *http.Request, message string, err error, details ...interface{}) {
	HandleErrorWithContext(context.Background(), w, r, http.StatusRequestTimeout, message, err, details...)
}

func HandleUnsupportedMediaType(w http.ResponseWriter, r *http.Request, message string, err error, details ...interface{}) {
	HandleErrorWithContext(context.Background(), w, r, http.StatusUnsupportedMediaType, message, err, details...)
}

func HandlePaymentRequired(w http.ResponseWriter, r *http.Request, message string, err error, details ...interface{}) {
	HandleErrorWithContext(context.Background(), w, r, http.StatusPaymentRequired, message, err, details...)
}

func HandleNotImplemented(w http.ResponseWriter, r *http.Request, message string, err error, details ...interface{}) {
	HandleErrorWithContext(context.Background(), w, r, http.StatusNotImplemented, message, err, details...)
}

func HandleBadGateway(w http.ResponseWriter, r *http.Request, message string, err error, details ...interface{}) {
	HandleErrorWithContext(context.Background(), w, r, http.StatusBadGateway, message, err, details...)
}

func HandleGatewayTimeout(w http.ResponseWriter, r *http.Request, message string, err error, details ...interface{}) {
	HandleErrorWithContext(context.Background(), w, r, http.StatusGatewayTimeout, message, err, details...)
}

func mapStatusToCode(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "BAD_REQUEST"
	case http.StatusUnauthorized:
		return "UNAUTHORIZED"
	case http.StatusPaymentRequired:
		return "PAYMENT_REQUIRED"
	case http.StatusForbidden:
		return "FORBIDDEN"
	case http.StatusNotFound:
		return "NOT_FOUND"
	case http.StatusMethodNotAllowed:
		return "METHOD_NOT_ALLOWED"
	case http.StatusNotAcceptable:
		return "NOT_ACCEPTABLE"
	case http.StatusProxyAuthRequired:
		return "PROXY_AUTH_REQUIRED"
	case http.StatusRequestTimeout:
		return "REQUEST_TIMEOUT"
	case http.StatusConflict:
		return "CONFLICT"
	case http.StatusGone:
		return "GONE"
	case http.StatusLengthRequired:
		return "LENGTH_REQUIRED"
	case http.StatusPreconditionFailed:
		return "PRECONDITION_FAILED"
	case http.StatusRequestEntityTooLarge:
		return "REQUEST_ENTITY_TOO_LARGE"
	case http.StatusRequestURITooLong:
		return "REQUEST_URI_TOO_LONG"
	case http.StatusUnsupportedMediaType:
		return "UNSUPPORTED_MEDIA_TYPE"
	case http.StatusRequestedRangeNotSatisfiable:
		return "REQUESTED_RANGE_NOT_SATISFIABLE"
	case http.StatusExpectationFailed:
		return "EXPECTATION_FAILED"
	case http.StatusTeapot:
		return "I_AM_A_TEAPOT"
	case http.StatusMisdirectedRequest:
		return "MISDIRECTED_REQUEST"
	case http.StatusUnprocessableEntity:
		return "UNPROCESSABLE_ENTITY"
	case http.StatusLocked:
		return "LOCKED"
	case http.StatusFailedDependency:
		return "FAILED_DEPENDENCY"
	case http.StatusTooEarly:
		return "TOO_EARLY"
	case http.StatusUpgradeRequired:
		return "UPGRADE_REQUIRED"
	case http.StatusPreconditionRequired:
		return "PRECONDITION_REQUIRED"
	case http.StatusTooManyRequests:
		return "TOO_MANY_REQUESTS"
	case http.StatusRequestHeaderFieldsTooLarge:
		return "REQUEST_HEADER_FIELDS_TOO_LARGE"
	case http.StatusUnavailableForLegalReasons:
		return "UNAVAILABLE_FOR_LEGAL_REASONS"
	case http.StatusInternalServerError:
		return "INTERNAL_SERVER_ERROR"
	case http.StatusNotImplemented:
		return "NOT_IMPLEMENTED"
	case http.StatusBadGateway:
		return "BAD_GATEWAY"
	case http.StatusServiceUnavailable:
		return "SERVICE_UNAVAILABLE"
	case http.StatusGatewayTimeout:
		return "GATEWAY_TIMEOUT"
	case http.StatusHTTPVersionNotSupported:
		return "HTTP_VERSION_NOT_SUPPORTED"
	case http.StatusVariantAlsoNegotiates:
		return "VARIANT_ALSO_NEGOTIATES"
	case http.StatusInsufficientStorage:
		return "INSUFFICIENT_STORAGE"
	case http.StatusLoopDetected:
		return "LOOP_DETECTED"
	case http.StatusNotExtended:
		return "NOT_EXTENDED"
	case http.StatusNetworkAuthenticationRequired:
		return "NETWORK_AUTHENTICATION_REQUIRED"
	default:
		return "UNKNOWN_ERROR"
	}

	}
	return "UNKNOWN_ERROR"
	}
}
