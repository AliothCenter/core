package restoration

import (
	"context"
	"runtime"
	"strconv"
	"time"

	"studio.sunist.work/platform/alioth-center/infrastructure/global"
	"studio.sunist.work/platform/alioth-center/infrastructure/utils"
)

type Fields interface {
	withBasic(ctx context.Context, message string) Fields
	withLevel(level string) Fields
	withService(name string) Fields
	withCaller(caller string) Fields
	WithParams(params any) Fields
	WithProcessing(processing any) Fields
	WithExtra(extra any) Fields
	export() *fields
}

type fields struct {
	service        string
	code           string
	level          string
	message        string
	calledAt       string
	calledFunction string
	traceID        string
	callerType     string
	inputFields    any
	payloadFields  any
	extraFields    any
}

func (f *fields) withBasic(ctx context.Context, message string) Fields {
	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "unknown"
		line = 0
	}
	f.code = file + ":" + strconv.Itoa(line)

	if fc := runtime.FuncForPC(pc); fc == nil {
		f.calledFunction = "unknown"
	} else {
		f.calledFunction = fc.Name()
	}

	traceID, getTraceErr := utils.GetTraceID(ctx)
	if getTraceErr != nil {
		f.traceID = ""
	} else {
		f.traceID = traceID
	}

	f.service = serviceName
	f.message = message
	f.calledAt = time.Now().Format(global.AliothTimeFormat)

	return f
}

func (f *fields) withLevel(level string) Fields {
	f.level = level
	return f
}

func (f *fields) withService(name string) Fields {
	f.service = name
	return f
}

func (f *fields) withCaller(caller string) Fields {
	f.callerType = caller
	return f
}

func (f *fields) WithParams(params any) Fields {
	f.extraFields = params
	return f
}

func (f *fields) WithProcessing(processing any) Fields {
	f.payloadFields = processing
	return f
}

func (f *fields) WithExtra(extra any) Fields {
	f.extraFields = extra
	return f
}

func (f *fields) export() *fields {
	return f
}

func NewCollection(ctx context.Context, message string) Fields {
	return (&fields{}).withBasic(ctx, message)
}
