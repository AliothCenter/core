package log

import (
	"context"
	"fmt"
	"runtime"

	"studio.sunist.work/platform/alioth-center/infrastructure/global/utils"

	"github.com/sirupsen/logrus"
)

type LoggerField interface {
	EncodePayload() map[string]any
	Level() LoggerLevel
	Message() string
}

type AliothLoggerField struct {
	level    LoggerLevel
	caller   LoggerCaller
	funcName string
	fileName string
	line     int
	message  string
	extra    any
	trace    string
	fields   map[string]any
}

func DefaultField() *AliothLoggerField {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "unknown"
		line = 0
	}

	if f := runtime.FuncForPC(pc); f == nil {
		return &AliothLoggerField{
			funcName: "unknown",
			fileName: file,
			line:     line,
		}
	} else {
		return &AliothLoggerField{
			funcName: f.Name(),
			fileName: file,
			line:     line,
		}
	}
}

func (f *AliothLoggerField) WithLevel(level LoggerLevel) *AliothLoggerField {
	f.level = level
	return f
}

func (f *AliothLoggerField) WithCaller(caller LoggerCaller) *AliothLoggerField {
	f.caller = caller
	return f
}

func (f *AliothLoggerField) WithMessage(message string) *AliothLoggerField {
	f.message = message
	return f
}

func (f *AliothLoggerField) WithExtraField(key string, value any) *AliothLoggerField {
	if f.fields == nil {
		f.fields = map[string]any{}
	}
	f.fields[key] = value
	return f
}

func (f *AliothLoggerField) WithExtra(extra ...any) *AliothLoggerField {
	if len(extra) == 1 {
		f.extra = extra[0]
	} else if len(extra) != 0 {
		f.extra = extra
	}
	return f
}

func (f *AliothLoggerField) WithContext(ctx context.Context) *AliothLoggerField {
	if traceID, err := utils.GetTraceID(ctx); err == nil {
		f.trace = traceID
	}
	return f
}

func (f *AliothLoggerField) WithFields(level LoggerLevel, caller LoggerCaller, message string, ctx context.Context, extra ...any) *AliothLoggerField {
	return f.WithLevel(level).WithCaller(caller).WithMessage(message).WithContext(ctx).WithExtra(extra...)
}

func (f *AliothLoggerField) EncodePayload() map[string]any {
	fields := logrus.Fields{
		"caller_type": f.caller,
		"function":    f.funcName,
		"filepath":    fmt.Sprintf("%s:%d", f.fileName, f.line),
	}

	if f.extra != nil {
		fields["extra"] = f.extra
	}

	if f.trace != "" {
		fields["trace_id"] = f.trace
	}

	return fields
}

func (f *AliothLoggerField) Level() LoggerLevel {
	return f.level
}

func (f *AliothLoggerField) Message() string {
	return f.message
}
