package response

import "strings"

// ErrorCode is a stable, reusable error classification used across
// article service and gateway handlers.
type ErrorCode string

const (
	ErrorCodeValidation       ErrorCode = "VALIDATION_ERROR"
	ErrorCodeInvalidJSON      ErrorCode = "INVALID_JSON"
	ErrorCodeInvalidArgument  ErrorCode = "INVALID_ARGUMENT"
	ErrorCodeNotFound         ErrorCode = "NOT_FOUND"
	ErrorCodeInternal         ErrorCode = "INTERNAL_ERROR"
	ErrorCodeTimeout          ErrorCode = "TIMEOUT"
	ErrorCodeUnavailable      ErrorCode = "SERVICE_UNAVAILABLE"
	ErrorCodeForbidden        ErrorCode = "FORBIDDEN"
	ErrorCodeUnauthorized     ErrorCode = "UNAUTHORIZED"
	ErrorCodeRequestTimeout   ErrorCode = "REQUEST_TIMEOUT"
)

type ErrorPayload struct {
	Code    string           `json:"code"`
	Message string           `json:"message"`
	Errors  map[string]string `json:"errors,omitempty"`
}

type APIErrorResponse struct {
	Error ErrorPayload `json:"error"`
}

type APIResponse map[string]any

// ErrorResponse builds a reusable payload with a stable `error` envelope.
func ErrorResponse(code ErrorCode, message string, fieldErrors map[string]string) APIResponse {
	payload := map[string]any{
		"error": map[string]any{
			"code":    string(code),
			"message": normalizeMessage(message),
		},
	}

	if len(fieldErrors) > 0 {
		payload["error"].(map[string]any)["errors"] = fieldErrors
	}

	return APIResponse(payload)
}

func normalizeMessage(message string) string {
	msg := strings.TrimSpace(message)
	if msg == "" {
		return "request failed"
	}
	return msg
}
