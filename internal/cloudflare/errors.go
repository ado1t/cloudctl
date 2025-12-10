package cloudflare

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorType 错误类型
type ErrorType int

const (
	// ErrorTypeUnknown 未知错误
	ErrorTypeUnknown ErrorType = iota
	// ErrorTypeAuth 认证错误
	ErrorTypeAuth
	// ErrorTypeNotFound 资源不存在
	ErrorTypeNotFound
	// ErrorTypeConflict 资源冲突
	ErrorTypeConflict
	// ErrorTypeRateLimit 速率限制
	ErrorTypeRateLimit
	// ErrorTypeNetwork 网络错误
	ErrorTypeNetwork
	// ErrorTypeValidation 验证错误
	ErrorTypeValidation
	// ErrorTypePermission 权限错误
	ErrorTypePermission
)

// Error Cloudflare 错误封装
type Error struct {
	Type      ErrorType
	Operation string
	Message   string
	Err       error
	HTTPCode  int
}

// Error 实现 error 接口
func (e *Error) Error() string {
	if e.Operation != "" {
		return fmt.Sprintf("%s: %s", e.Operation, e.Message)
	}
	return e.Message
}

// Unwrap 实现 errors.Unwrap
func (e *Error) Unwrap() error {
	return e.Err
}

// Is 实现 errors.Is
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	return e.Type == t.Type
}

// WrapError 封装错误
func WrapError(err error, operation string) error {
	if err == nil {
		return nil
	}

	// 如果已经是我们的错误类型，直接返回
	var cfErr *Error
	if errors.As(err, &cfErr) {
		if cfErr.Operation == "" {
			cfErr.Operation = operation
		}
		return cfErr
	}

	// 创建新的错误
	cfErr = &Error{
		Type:      ErrorTypeUnknown,
		Operation: operation,
		Message:   err.Error(),
		Err:       err,
	}

	// 根据错误内容判断类型
	cfErr.Type = classifyError(err)

	return cfErr
}

// classifyError 根据错误内容分类错误
func classifyError(err error) ErrorType {
	if err == nil {
		return ErrorTypeUnknown
	}

	errMsg := err.Error()

	// 检查常见的错误模式
	switch {
	case contains(errMsg, "authentication", "unauthorized", "invalid token", "invalid api token"):
		return ErrorTypeAuth
	case contains(errMsg, "not found", "does not exist"):
		return ErrorTypeNotFound
	case contains(errMsg, "already exists", "conflict", "duplicate"):
		return ErrorTypeConflict
	case contains(errMsg, "rate limit", "too many requests"):
		return ErrorTypeRateLimit
	case contains(errMsg, "network", "connection", "timeout", "dial"):
		return ErrorTypeNetwork
	case contains(errMsg, "invalid", "validation", "bad request"):
		return ErrorTypeValidation
	case contains(errMsg, "forbidden", "permission denied", "access denied"):
		return ErrorTypePermission
	default:
		return ErrorTypeUnknown
	}
}

// contains 检查字符串是否包含任意一个子串（不区分大小写）
func contains(s string, substrs ...string) bool {
	s = toLower(s)
	for _, substr := range substrs {
		if containsSubstr(s, toLower(substr)) {
			return true
		}
	}
	return false
}

// toLower 转换为小写
func toLower(s string) string {
	// 简单实现，避免引入 strings 包
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c = c + ('a' - 'A')
		}
		result[i] = c
	}
	return string(result)
}

// containsSubstr 检查字符串是否包含子串
func containsSubstr(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// NewError 创建新的错误
func NewError(errType ErrorType, operation, message string) error {
	return &Error{
		Type:      errType,
		Operation: operation,
		Message:   message,
	}
}

// NewAuthError 创建认证错误
func NewAuthError(operation, message string) error {
	return NewError(ErrorTypeAuth, operation, message)
}

// NewNotFoundError 创建资源不存在错误
func NewNotFoundError(operation, resource string) error {
	return NewError(ErrorTypeNotFound, operation, fmt.Sprintf("资源不存在: %s", resource))
}

// NewConflictError 创建资源冲突错误
func NewConflictError(operation, resource string) error {
	return NewError(ErrorTypeConflict, operation, fmt.Sprintf("资源已存在: %s", resource))
}

// NewValidationError 创建验证错误
func NewValidationError(operation, message string) error {
	return NewError(ErrorTypeValidation, operation, message)
}

// IsAuthError 判断是否是认证错误
func IsAuthError(err error) bool {
	var cfErr *Error
	if errors.As(err, &cfErr) {
		return cfErr.Type == ErrorTypeAuth
	}
	return false
}

// IsNotFoundError 判断是否是资源不存在错误
func IsNotFoundError(err error) bool {
	var cfErr *Error
	if errors.As(err, &cfErr) {
		return cfErr.Type == ErrorTypeNotFound
	}
	return false
}

// IsConflictError 判断是否是资源冲突错误
func IsConflictError(err error) bool {
	var cfErr *Error
	if errors.As(err, &cfErr) {
		return cfErr.Type == ErrorTypeConflict
	}
	return false
}

// IsRateLimitError 判断是否是速率限制错误
func IsRateLimitError(err error) bool {
	var cfErr *Error
	if errors.As(err, &cfErr) {
		return cfErr.Type == ErrorTypeRateLimit
	}
	return false
}

// IsNetworkError 判断是否是网络错误
func IsNetworkError(err error) bool {
	var cfErr *Error
	if errors.As(err, &cfErr) {
		return cfErr.Type == ErrorTypeNetwork
	}
	return false
}

// IsValidationError 判断是否是验证错误
func IsValidationError(err error) bool {
	var cfErr *Error
	if errors.As(err, &cfErr) {
		return cfErr.Type == ErrorTypeValidation
	}
	return false
}

// IsPermissionError 判断是否是权限错误
func IsPermissionError(err error) bool {
	var cfErr *Error
	if errors.As(err, &cfErr) {
		return cfErr.Type == ErrorTypePermission
	}
	return false
}

// GetHTTPStatusCode 从错误中提取 HTTP 状态码
func GetHTTPStatusCode(err error) int {
	var cfErr *Error
	if errors.As(err, &cfErr) {
		return cfErr.HTTPCode
	}
	return 0
}

// FormatError 格式化错误信息，提供用户友好的错误消息
func FormatError(err error) string {
	if err == nil {
		return ""
	}

	var cfErr *Error
	if !errors.As(err, &cfErr) {
		return err.Error()
	}

	// 根据错误类型提供友好的错误消息
	switch cfErr.Type {
	case ErrorTypeAuth:
		return fmt.Sprintf("认证失败: %s\n提示: 请检查 API Token 是否正确配置", cfErr.Message)
	case ErrorTypeNotFound:
		return fmt.Sprintf("资源不存在: %s\n提示: 请使用 list 命令查看可用资源", cfErr.Message)
	case ErrorTypeConflict:
		return fmt.Sprintf("资源冲突: %s\n提示: 资源可能已存在，请检查后重试", cfErr.Message)
	case ErrorTypeRateLimit:
		return fmt.Sprintf("请求频率超限: %s\n提示: 请稍后重试或减少请求频率", cfErr.Message)
	case ErrorTypeNetwork:
		return fmt.Sprintf("网络错误: %s\n提示: 请检查网络连接", cfErr.Message)
	case ErrorTypeValidation:
		return fmt.Sprintf("参数验证失败: %s\n提示: 请检查命令参数是否正确", cfErr.Message)
	case ErrorTypePermission:
		return fmt.Sprintf("权限不足: %s\n提示: 请检查 API Token 权限配置", cfErr.Message)
	default:
		return cfErr.Error()
	}
}

// GetExitCode 根据错误类型返回退出码
func GetExitCode(err error) int {
	if err == nil {
		return 0
	}

	var cfErr *Error
	if !errors.As(err, &cfErr) {
		return 1 // 一般错误
	}

	switch cfErr.Type {
	case ErrorTypeAuth:
		return 3 // 认证错误
	case ErrorTypeValidation:
		return 5 // 参数错误
	case ErrorTypePermission:
		return 3 // 权限错误（同认证错误）
	case ErrorTypeNetwork:
		return 4 // API 错误
	case ErrorTypeRateLimit:
		return 4 // API 错误
	default:
		return 1 // 一般错误
	}
}

// HTTPStatusToErrorType 将 HTTP 状态码转换为错误类型
func HTTPStatusToErrorType(statusCode int) ErrorType {
	switch statusCode {
	case http.StatusUnauthorized:
		return ErrorTypeAuth
	case http.StatusForbidden:
		return ErrorTypePermission
	case http.StatusNotFound:
		return ErrorTypeNotFound
	case http.StatusConflict:
		return ErrorTypeConflict
	case http.StatusTooManyRequests:
		return ErrorTypeRateLimit
	case http.StatusBadRequest:
		return ErrorTypeValidation
	default:
		if statusCode >= 500 {
			return ErrorTypeNetwork
		}
		return ErrorTypeUnknown
	}
}
