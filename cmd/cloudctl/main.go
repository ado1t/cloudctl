package main

import (
	"fmt"
	"os"

	"github.com/ado1t/cloudctl/internal/cmd"
)

var (
	// Version 版本号，通过 ldflags 注入
	Version = "dev"
	// BuildTime 构建时间，通过 ldflags 注入
	BuildTime = "unknown"
)

func main() {
	// 设置版本信息
	cmd.SetVersionInfo(Version, BuildTime)

	// 执行根命令
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
