package logger

import "strings"

// VerbosityToLevel 将 verbosity 计数转换为日志级别
// -v    -> info
// -vv   -> debug
// -vvv  -> debug (with source)
func VerbosityToLevel(verbosity int) (LogLevel, bool) {
	switch verbosity {
	case 0:
		return LevelError, false
	case 1:
		return LevelInfo, false
	case 2:
		return LevelDebug, false
	case 3:
		return LevelDebug, true // 添加源代码位置
	default:
		if verbosity > 3 {
			return LevelDebug, true
		}
		return LevelError, false
	}
}

// ParseLevel 解析日志级别字符串
func ParseLevel(level string) LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

// LevelToString 将日志级别转换为字符串
func LevelToString(level LogLevel) string {
	return string(level)
}

// IsValidLevel 检查日志级别是否有效
func IsValidLevel(level string) bool {
	switch strings.ToLower(level) {
	case "debug", "info", "warn", "warning", "error":
		return true
	default:
		return false
	}
}
