package cloudflare

import (
	"errors"
	"testing"
)

func TestWrapError(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		operation string
		wantType  ErrorType
	}{
		{
			name:      "认证错误",
			err:       errors.New("authentication failed"),
			operation: "测试操作",
			wantType:  ErrorTypeAuth,
		},
		{
			name:      "资源不存在",
			err:       errors.New("zone not found"),
			operation: "测试操作",
			wantType:  ErrorTypeNotFound,
		},
		{
			name:      "资源冲突",
			err:       errors.New("record already exists"),
			operation: "测试操作",
			wantType:  ErrorTypeConflict,
		},
		{
			name:      "速率限制",
			err:       errors.New("rate limit exceeded"),
			operation: "测试操作",
			wantType:  ErrorTypeRateLimit,
		},
		{
			name:      "网络错误",
			err:       errors.New("network timeout"),
			operation: "测试操作",
			wantType:  ErrorTypeNetwork,
		},
		{
			name:      "验证错误",
			err:       errors.New("invalid domain name"),
			operation: "测试操作",
			wantType:  ErrorTypeValidation,
		},
		{
			name:      "权限错误",
			err:       errors.New("permission denied"),
			operation: "测试操作",
			wantType:  ErrorTypePermission,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapped := WrapError(tt.err, tt.operation)
			if wrapped == nil {
				t.Fatal("WrapError() 返回 nil")
			}

			var cfErr *Error
			if !errors.As(wrapped, &cfErr) {
				t.Fatal("WrapError() 未返回 *Error 类型")
			}

			if cfErr.Type != tt.wantType {
				t.Errorf("错误类型 = %v, 期望 %v", cfErr.Type, tt.wantType)
			}

			if cfErr.Operation != tt.operation {
				t.Errorf("操作 = %v, 期望 %v", cfErr.Operation, tt.operation)
			}
		})
	}
}

func TestWrapError_Nil(t *testing.T) {
	err := WrapError(nil, "测试操作")
	if err != nil {
		t.Errorf("WrapError(nil) = %v, 期望 nil", err)
	}
}

func TestWrapError_AlreadyWrapped(t *testing.T) {
	original := NewAuthError("原始操作", "认证失败")
	wrapped := WrapError(original, "新操作")

	var cfErr *Error
	if !errors.As(wrapped, &cfErr) {
		t.Fatal("WrapError() 未返回 *Error 类型")
	}

	// 应该保持原始操作名
	if cfErr.Operation != "原始操作" {
		t.Errorf("操作 = %v, 期望 '原始操作'", cfErr.Operation)
	}
}

func TestNewError(t *testing.T) {
	err := NewError(ErrorTypeAuth, "测试操作", "测试消息")
	if err == nil {
		t.Fatal("NewError() 返回 nil")
	}

	var cfErr *Error
	if !errors.As(err, &cfErr) {
		t.Fatal("NewError() 未返回 *Error 类型")
	}

	if cfErr.Type != ErrorTypeAuth {
		t.Errorf("类型 = %v, 期望 %v", cfErr.Type, ErrorTypeAuth)
	}

	if cfErr.Operation != "测试操作" {
		t.Errorf("操作 = %v, 期望 '测试操作'", cfErr.Operation)
	}

	if cfErr.Message != "测试消息" {
		t.Errorf("消息 = %v, 期望 '测试消息'", cfErr.Message)
	}
}

func TestIsAuthError(t *testing.T) {
	authErr := NewAuthError("测试", "认证失败")
	if !IsAuthError(authErr) {
		t.Error("IsAuthError() = false, 期望 true")
	}

	otherErr := NewNotFoundError("测试", "资源")
	if IsAuthError(otherErr) {
		t.Error("IsAuthError() = true, 期望 false")
	}

	if IsAuthError(nil) {
		t.Error("IsAuthError(nil) = true, 期望 false")
	}
}

func TestIsNotFoundError(t *testing.T) {
	notFoundErr := NewNotFoundError("测试", "资源")
	if !IsNotFoundError(notFoundErr) {
		t.Error("IsNotFoundError() = false, 期望 true")
	}

	otherErr := NewAuthError("测试", "认证失败")
	if IsNotFoundError(otherErr) {
		t.Error("IsNotFoundError() = true, 期望 false")
	}
}

func TestIsConflictError(t *testing.T) {
	conflictErr := NewConflictError("测试", "资源")
	if !IsConflictError(conflictErr) {
		t.Error("IsConflictError() = false, 期望 true")
	}

	otherErr := NewAuthError("测试", "认证失败")
	if IsConflictError(otherErr) {
		t.Error("IsConflictError() = true, 期望 false")
	}
}

func TestIsValidationError(t *testing.T) {
	validationErr := NewValidationError("测试", "验证失败")
	if !IsValidationError(validationErr) {
		t.Error("IsValidationError() = false, 期望 true")
	}

	otherErr := NewAuthError("测试", "认证失败")
	if IsValidationError(otherErr) {
		t.Error("IsValidationError() = true, 期望 false")
	}
}

func TestFormatError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "nil 错误",
			err:  nil,
			want: "",
		},
		{
			name: "认证错误",
			err:  NewAuthError("测试", "无效的 token"),
			want: "认证失败: 无效的 token\n提示: 请检查 API Token 是否正确配置",
		},
		{
			name: "资源不存在",
			err:  NewNotFoundError("测试", "example.com"),
			want: "资源不存在: 资源不存在: example.com\n提示: 请使用 list 命令查看可用资源",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatError(tt.err)
			if got != tt.want {
				t.Errorf("FormatError() = %v, 期望 %v", got, tt.want)
			}
		})
	}
}

func TestGetExitCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{
			name: "nil 错误",
			err:  nil,
			want: 0,
		},
		{
			name: "认证错误",
			err:  NewAuthError("测试", "认证失败"),
			want: 3,
		},
		{
			name: "验证错误",
			err:  NewValidationError("测试", "验证失败"),
			want: 5,
		},
		{
			name: "权限错误",
			err:  NewError(ErrorTypePermission, "测试", "权限不足"),
			want: 3,
		},
		{
			name: "网络错误",
			err:  NewError(ErrorTypeNetwork, "测试", "网络超时"),
			want: 4,
		},
		{
			name: "未知错误",
			err:  NewError(ErrorTypeUnknown, "测试", "未知错误"),
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetExitCode(tt.err)
			if got != tt.want {
				t.Errorf("GetExitCode() = %v, 期望 %v", got, tt.want)
			}
		})
	}
}

func TestHTTPStatusToErrorType(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       ErrorType
	}{
		{"401 Unauthorized", 401, ErrorTypeAuth},
		{"403 Forbidden", 403, ErrorTypePermission},
		{"404 Not Found", 404, ErrorTypeNotFound},
		{"409 Conflict", 409, ErrorTypeConflict},
		{"429 Too Many Requests", 429, ErrorTypeRateLimit},
		{"400 Bad Request", 400, ErrorTypeValidation},
		{"500 Internal Server Error", 500, ErrorTypeNetwork},
		{"200 OK", 200, ErrorTypeUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HTTPStatusToErrorType(tt.statusCode)
			if got != tt.want {
				t.Errorf("HTTPStatusToErrorType() = %v, 期望 %v", got, tt.want)
			}
		})
	}
}

func TestError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *Error
		want string
	}{
		{
			name: "带操作名",
			err: &Error{
				Type:      ErrorTypeAuth,
				Operation: "测试操作",
				Message:   "认证失败",
			},
			want: "测试操作: 认证失败",
		},
		{
			name: "不带操作名",
			err: &Error{
				Type:    ErrorTypeAuth,
				Message: "认证失败",
			},
			want: "认证失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("Error() = %v, 期望 %v", got, tt.want)
			}
		})
	}
}

func TestError_Unwrap(t *testing.T) {
	innerErr := errors.New("内部错误")
	err := &Error{
		Type:    ErrorTypeUnknown,
		Message: "外部错误",
		Err:     innerErr,
	}

	unwrapped := errors.Unwrap(err)
	if unwrapped != innerErr {
		t.Errorf("Unwrap() = %v, 期望 %v", unwrapped, innerErr)
	}
}

func TestError_Is(t *testing.T) {
	err1 := &Error{Type: ErrorTypeAuth}
	err2 := &Error{Type: ErrorTypeAuth}
	err3 := &Error{Type: ErrorTypeNotFound}

	if !errors.Is(err1, err2) {
		t.Error("相同类型的错误应该匹配")
	}

	if errors.Is(err1, err3) {
		t.Error("不同类型的错误不应该匹配")
	}
}
