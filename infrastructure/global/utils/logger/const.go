package log

import "strings"

type LoggerLevel string

const (
	Debug LoggerLevel = "debug"
	Info  LoggerLevel = "info"
	Warn  LoggerLevel = "warn"
	Error LoggerLevel = "error"
	Panic LoggerLevel = "panic"
)

type LoggerCaller string

const (
	Internal LoggerCaller = "internal"
	Module   LoggerCaller = "module"
	Service  LoggerCaller = "service"
	External LoggerCaller = "external"
)

func NewLevelFromString(level string) LoggerLevel {
	switch strings.ToLower(level) {
	case "debug":
		return Debug
	case "info":
		return Info
	case "warn":
		return Warn
	case "error":
		return Error
	default:
		return Info
	}
}
