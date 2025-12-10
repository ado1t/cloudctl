package cmd

import (
	"testing"
)

func TestSetVersionInfo(t *testing.T) {
	ver := "1.0.0"
	build := "2024-01-01_12:00:00"

	SetVersionInfo(ver, build)

	if version != ver {
		t.Errorf("Expected version %s, got %s", ver, version)
	}

	if buildTime != build {
		t.Errorf("Expected buildTime %s, got %s", build, buildTime)
	}

	expectedVersion := "1.0.0 (built at 2024-01-01_12:00:00)"
	if rootCmd.Version != expectedVersion {
		t.Errorf("Expected rootCmd.Version %s, got %s", expectedVersion, rootCmd.Version)
	}
}

func TestRootCommandExists(t *testing.T) {
	if rootCmd == nil {
		t.Error("rootCmd should not be nil")
	}

	if rootCmd.Use != "cloudctl" {
		t.Errorf("Expected Use to be 'cloudctl', got '%s'", rootCmd.Use)
	}
}

func TestSubCommandsExist(t *testing.T) {
	commands := rootCmd.Commands()

	// 检查必需的子命令是否存在
	requiredCommands := []string{"config", "cf", "aws"}
	foundCommands := make(map[string]bool)

	for _, cmd := range commands {
		foundCommands[cmd.Name()] = true
	}

	for _, required := range requiredCommands {
		if !foundCommands[required] {
			t.Errorf("Missing required command: %s", required)
		}
	}
}
