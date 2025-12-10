package cmd

import (
	"testing"
)

func TestExtractZoneFromValidationName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple domain",
			input:    "_c93b9a33d8bde7910f5e9680040b841c.example.com",
			expected: "example.com",
		},
		{
			name:     "subdomain",
			input:    "_c93b9a33d8bde7910f5e9680040b841c.sub.example.com",
			expected: "sub.example.com",
		},
		{
			name:     "multiple subdomains",
			input:    "_c93b9a33d8bde7910f5e9680040b841c.api.v1.example.com",
			expected: "api.v1.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractZoneFromValidationName(tt.input)
			if result != tt.expected {
				t.Errorf("extractZoneFromValidationName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExtractSubdomainFromValidationName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		zone     string
		expected string
	}{
		{
			name:     "simple domain",
			input:    "_c93b9a33d8bde7910f5e9680040b841c.example.com",
			zone:     "example.com",
			expected: "_c93b9a33d8bde7910f5e9680040b841c",
		},
		{
			name:     "subdomain",
			input:    "_c93b9a33d8bde7910f5e9680040b841c.sub.example.com",
			zone:     "sub.example.com",
			expected: "_c93b9a33d8bde7910f5e9680040b841c",
		},
		{
			name:     "root domain",
			input:    "example.com",
			zone:     "example.com",
			expected: "@",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractSubdomainFromValidationName(tt.input, tt.zone)
			if result != tt.expected {
				t.Errorf("extractSubdomainFromValidationName(%q, %q) = %q, want %q", tt.input, tt.zone, result, tt.expected)
			}
		})
	}
}
