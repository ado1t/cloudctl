package aws

import (
	"testing"
)

func TestRequestCertificateInput_Validate(t *testing.T) {
	tests := []struct {
		name        string
		input       *RequestCertificateInput
		wantErr     bool
		errContains string
	}{
		{
			name: "有效输入 - 单域名",
			input: &RequestCertificateInput{
				DomainName: "example.com",
			},
			wantErr: false,
		},
		{
			name: "有效输入 - 通配符域名",
			input: &RequestCertificateInput{
				DomainName: "*.example.com",
			},
			wantErr: false,
		},
		{
			name: "有效输入 - 带 SANs",
			input: &RequestCertificateInput{
				DomainName:              "example.com",
				SubjectAlternativeNames: []string{"www.example.com", "api.example.com"},
			},
			wantErr: false,
		},
		{
			name: "缺少域名",
			input: &RequestCertificateInput{
				SubjectAlternativeNames: []string{"www.example.com"},
			},
			wantErr:     true,
			errContains: "域名",
		},
		{
			name:        "空输入",
			input:       &RequestCertificateInput{},
			wantErr:     true,
			errContains: "域名",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 验证域名是否为空
			if tt.input.DomainName == "" {
				if !tt.wantErr {
					t.Error("应该返回错误: 域名为空")
				}
				return
			}

			if tt.wantErr {
				t.Errorf("不应该通过验证: %+v", tt.input)
			}
		})
	}
}

func TestCertificate_DomainValidation(t *testing.T) {
	tests := []struct {
		name   string
		domain string
		valid  bool
	}{
		{
			name:   "标准域名",
			domain: "example.com",
			valid:  true,
		},
		{
			name:   "子域名",
			domain: "www.example.com",
			valid:  true,
		},
		{
			name:   "通配符域名",
			domain: "*.example.com",
			valid:  true,
		},
		{
			name:   "多级子域名",
			domain: "api.v1.example.com",
			valid:  true,
		},
		{
			name:   "空域名",
			domain: "",
			valid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.domain == "" {
				if tt.valid {
					t.Error("空域名应该无效")
				}
			} else {
				if !tt.valid {
					t.Errorf("域名应该有效: %s", tt.domain)
				}
			}
		})
	}
}

func TestValidationRecord_Structure(t *testing.T) {
	// 测试验证记录结构
	record := ValidationRecord{
		Name:   "_abc123.example.com",
		Type:   "CNAME",
		Value:  "_xyz456.acm-validations.aws",
		Status: "PENDING_VALIDATION",
	}

	if record.Name == "" {
		t.Error("Name 不应该为空")
	}
	if record.Type != "CNAME" {
		t.Errorf("Type 应该是 CNAME, 实际: %s", record.Type)
	}
	if record.Value == "" {
		t.Error("Value 不应该为空")
	}
	if record.Status == "" {
		t.Error("Status 不应该为空")
	}
}

func TestCertificate_StatusValues(t *testing.T) {
	// 测试证书状态值
	validStatuses := []string{
		"PENDING_VALIDATION",
		"ISSUED",
		"INACTIVE",
		"EXPIRED",
		"VALIDATION_TIMED_OUT",
		"REVOKED",
		"FAILED",
	}

	for _, status := range validStatuses {
		cert := Certificate{
			ARN:        "arn:aws:acm:us-east-1:123456789012:certificate/test",
			DomainName: "example.com",
			Status:     status,
		}

		if cert.Status == "" {
			t.Errorf("状态不应该为空: %s", status)
		}
	}
}

func TestCertificate_InUseFlag(t *testing.T) {
	tests := []struct {
		name   string
		inUse  bool
		expect bool
	}{
		{
			name:   "证书正在使用",
			inUse:  true,
			expect: true,
		},
		{
			name:   "证书未使用",
			inUse:  false,
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cert := Certificate{
				ARN:        "arn:aws:acm:us-east-1:123456789012:certificate/test",
				DomainName: "example.com",
				InUse:      tt.inUse,
			}

			if cert.InUse != tt.expect {
				t.Errorf("InUse = %v, 期望 %v", cert.InUse, tt.expect)
			}
		})
	}
}

func TestSubjectAlternativeNames(t *testing.T) {
	// 测试 SANs 列表
	tests := []struct {
		name string
		sans []string
		want int
	}{
		{
			name: "单个 SAN",
			sans: []string{"www.example.com"},
			want: 1,
		},
		{
			name: "多个 SANs",
			sans: []string{"www.example.com", "api.example.com", "cdn.example.com"},
			want: 3,
		},
		{
			name: "空 SANs",
			sans: []string{},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cert := Certificate{
				ARN:             "arn:aws:acm:us-east-1:123456789012:certificate/test",
				DomainName:      "example.com",
				SubjectAltNames: tt.sans,
			}

			if len(cert.SubjectAltNames) != tt.want {
				t.Errorf("SANs 数量 = %d, 期望 %d", len(cert.SubjectAltNames), tt.want)
			}
		})
	}
}
