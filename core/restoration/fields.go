package restoration

import (
	"context"

	"studio.sunist.work/platform/alioth-center/infrastructure/global/utils"
	log "studio.sunist.work/platform/alioth-center/infrastructure/global/utils/logger"
	"studio.sunist.work/platform/alioth-center/proto/alioth"
)

type Fields struct {
	callerIP       string
	service        string
	code           string
	level          string
	caller         string
	message        string
	calledAt       string
	calledFunction string
	traceID        string
	inputFields    []byte
	payloadFields  []byte
	extraFields    []byte
}

func (f *Fields) EncodePayload() map[string]any {
	payload := map[string]any{
		"caller_service":  f.service,
		"code_path":       f.code,
		"called_at":       f.calledAt,
		"called_function": f.calledFunction,
		"caller_type":     f.caller,
	}

	if f.traceID != "" {
		payload["trace_id"] = f.traceID
	}

	if f.callerIP != "" {
		payload["caller_ip"] = f.callerIP
	}

	if f.inputFields != nil && len(f.inputFields) > 0 {
		var inputFields any
		utils.JsonUnmarshal(f.inputFields, &inputFields)
		payload["caller_arguments"] = inputFields
	}

	if f.payloadFields != nil && len(f.payloadFields) > 0 {
		var payloadFields any
		utils.JsonUnmarshal(f.payloadFields, &payloadFields)
		payload["caller_processing"] = payloadFields
	}

	if f.extraFields != nil && len(f.extraFields) > 0 {
		var extraFields any
		utils.JsonUnmarshal(f.extraFields, &extraFields)
		payload["extra_data"] = extraFields
	}

	return payload
}

func (f *Fields) Level() log.LoggerLevel {
	return log.NewLevelFromString(f.level)
}

func (f *Fields) Message() string {
	return f.message
}

func NewRestorationFieldsFromRequest(ctx context.Context, request *alioth.RestorationCollectionRequest) log.LoggerField {
	ip, getIPErr := utils.GetContextClientIP(ctx)
	if getIPErr != nil {
		ip = ""
	}
	return &Fields{
		callerIP:       ip,
		service:        request.GetCallerService(),
		code:           request.GetCodePath(),
		level:          request.GetLevel(),
		message:        request.GetMessage(),
		calledAt:       request.GetCalledAt(),
		calledFunction: request.GetCalledFunction(),
		traceID:        request.GetTraceId(),
		inputFields:    request.GetInputFields(),
		payloadFields:  request.GetPayloadFields(),
		extraFields:    request.GetExtraFields(),
		caller:         string(log.Service),
	}
}

func NewExternalRestorationFieldsFromRequest(ctx context.Context, request *alioth.RestorationCollectionRequest) log.LoggerField {
	ip, getIPErr := utils.GetContextClientIP(ctx)
	if getIPErr != nil {
		ip = ""
	}
	return &Fields{
		callerIP:       ip,
		service:        request.GetCallerService(),
		code:           request.GetCodePath(),
		level:          request.GetLevel(),
		message:        request.GetMessage(),
		calledAt:       request.GetCalledAt(),
		calledFunction: request.GetCalledFunction(),
		traceID:        request.GetTraceId(),
		inputFields:    request.GetInputFields(),
		payloadFields:  request.GetPayloadFields(),
		extraFields:    request.GetExtraFields(),
		caller:         string(log.External),
	}
}
