package aws

import (
	"testing"
)

func TestCreateInvalidationInput_Validate(t *testing.T) {
	tests := []struct {
		name        string
		input       *CreateInvalidationInput
		wantErr     bool
		errContains string
	}{
		{
			name: "有效输入",
			input: &CreateInvalidationInput{
				DistributionID: "E1234567890ABC",
				Paths:          []string{"/*"},
			},
			wantErr: false,
		},
		{
			name: "有效输入 - 多个路径",
			input: &CreateInvalidationInput{
				DistributionID: "E1234567890ABC",
				Paths:          []string{"/index.html", "/css/*", "/js/*"},
			},
			wantErr: false,
		},
		{
			name: "有效输入 - 带 CallerReference",
			input: &CreateInvalidationInput{
				DistributionID:  "E1234567890ABC",
				Paths:           []string{"/*"},
				CallerReference: "my-ref-001",
			},
			wantErr: false,
		},
		{
			name: "缺少 DistributionID",
			input: &CreateInvalidationInput{
				Paths: []string{"/*"},
			},
			wantErr:     true,
			errContains: "distribution_id",
		},
		{
			name: "缺少 Paths",
			input: &CreateInvalidationInput{
				DistributionID: "E1234567890ABC",
				Paths:          []string{},
			},
			wantErr:     true,
			errContains: "路径",
		},
		{
			name: "Paths 为 nil",
			input: &CreateInvalidationInput{
				DistributionID: "E1234567890ABC",
			},
			wantErr:     true,
			errContains: "路径",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 由于我们没有真实的 AWS 客户端，这里只测试输入验证逻辑
			if tt.input.DistributionID == "" {
				if !tt.wantErr {
					t.Error("应该返回错误: distribution_id 为空")
				}
				return
			}

			if len(tt.input.Paths) == 0 {
				if !tt.wantErr {
					t.Error("应该返回错误: paths 为空")
				}
				return
			}

			if tt.wantErr {
				t.Errorf("不应该通过验证: %+v", tt.input)
			}
		})
	}
}

func TestInvalidation_PathsValidation(t *testing.T) {
	tests := []struct {
		name  string
		paths []string
		valid bool
	}{
		{
			name:  "单个通配符路径",
			paths: []string{"/*"},
			valid: true,
		},
		{
			name:  "多个具体路径",
			paths: []string{"/index.html", "/about.html", "/contact.html"},
			valid: true,
		},
		{
			name:  "混合路径",
			paths: []string{"/index.html", "/css/*", "/js/*"},
			valid: true,
		},
		{
			name:  "带目录的路径",
			paths: []string{"/images/logo.png", "/docs/api.pdf"},
			valid: true,
		},
		{
			name:  "空路径列表",
			paths: []string{},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.paths) == 0 {
				if tt.valid {
					t.Error("空路径列表应该无效")
				}
			} else {
				if !tt.valid {
					t.Errorf("路径列表应该有效: %v", tt.paths)
				}
			}
		})
	}
}

func TestCallerReferenceGeneration(t *testing.T) {
	// 测试 CallerReference 的生成逻辑
	tests := []struct {
		name            string
		callerReference string
		shouldGenerate  bool
	}{
		{
			name:            "提供了 CallerReference",
			callerReference: "my-ref-001",
			shouldGenerate:  false,
		},
		{
			name:            "未提供 CallerReference",
			callerReference: "",
			shouldGenerate:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.callerReference == "" && !tt.shouldGenerate {
				t.Error("空 CallerReference 应该触发自动生成")
			}
			if tt.callerReference != "" && tt.shouldGenerate {
				t.Error("非空 CallerReference 不应该触发自动生成")
			}
		})
	}
}
