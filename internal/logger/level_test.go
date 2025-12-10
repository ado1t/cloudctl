package logger

import "testing"

func TestVerbosityToLevel(t *testing.T) {
	tests := []struct {
		verbosity      int
		expectedLevel  LogLevel
		expectedSource bool
	}{
		{0, LevelError, false},
		{1, LevelInfo, false},
		{2, LevelDebug, false},
		{3, LevelDebug, true},
		{4, LevelDebug, true},
	}

	for _, tt := range tests {
		t.Run(string(rune('0'+tt.verbosity)), func(t *testing.T) {
			level, addSource := VerbosityToLevel(tt.verbosity)
			if level != tt.expectedLevel {
				t.Errorf("VerbosityToLevel(%d) level = %v, want %v", tt.verbosity, level, tt.expectedLevel)
			}
			if addSource != tt.expectedSource {
				t.Errorf("VerbosityToLevel(%d) addSource = %v, want %v", tt.verbosity, addSource, tt.expectedSource)
			}
		})
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
	}{
		{"debug", LevelDebug},
		{"DEBUG", LevelDebug},
		{"info", LevelInfo},
		{"INFO", LevelInfo},
		{"warn", LevelWarn},
		{"WARN", LevelWarn},
		{"warning", LevelWarn},
		{"error", LevelError},
		{"ERROR", LevelError},
		{"invalid", LevelInfo},
		{"", LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseLevel(tt.input)
			if result != tt.expected {
				t.Errorf("ParseLevel(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsValidLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"debug", true},
		{"info", true},
		{"warn", true},
		{"warning", true},
		{"error", true},
		{"DEBUG", true},
		{"INFO", true},
		{"invalid", false},
		{"", false},
		{"trace", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := IsValidLevel(tt.input)
			if result != tt.expected {
				t.Errorf("IsValidLevel(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLevelToString(t *testing.T) {
	tests := []struct {
		input    LogLevel
		expected string
	}{
		{LevelDebug, "debug"},
		{LevelInfo, "info"},
		{LevelWarn, "warn"},
		{LevelError, "error"},
	}

	for _, tt := range tests {
		t.Run(string(tt.input), func(t *testing.T) {
			result := LevelToString(tt.input)
			if result != tt.expected {
				t.Errorf("LevelToString(%v) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}
