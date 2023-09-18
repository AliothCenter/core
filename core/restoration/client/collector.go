package restoration

import (
	"fmt"

	"studio.sunist.work/platform/alioth-center/infrastructure/utils"
	log "studio.sunist.work/platform/alioth-center/infrastructure/utils/logger"

	"studio.sunist.work/platform/alioth-center/proto/alioth"
)

var (
	serviceName = "unregistered-service"
	logger      = log.DefaultLogger()

	nilCollector Collector = nil
)

func SetDefaultServiceName(name string) {
	serviceName = name
}

// Collector 分布式日志收集器，向 alioth-restoration 服务发送日志
type Collector interface {
	// Debug 记录一个 debug 级别的日志
	//   - fields: 日志字段
	Debug(fields Fields)

	// Info 记录一个 info 级别的日志
	//   - fields: 日志字段
	Info(fields Fields)

	// Warn 记录一个 warn 级别的日志
	//   - fields: 日志字段
	Warn(fields Fields)

	// Error 记录一个 error 级别的日志
	//   - fields: 日志字段
	Error(fields Fields)
}

type collector struct {
	client      *client
	serviceName string
}

func (r *collector) logField(fields Fields) {
	exported := fields.export()
	if exported.service == "" {
		exported.service = r.serviceName
	}

	var paramsBytes []byte
	if exported.inputFields != nil {
		paramsBytes = utils.JsonMarshal(exported.inputFields)
	}

	var processingBytes []byte
	if exported.payloadFields != nil {
		processingBytes = utils.JsonMarshal(exported.payloadFields)
	}

	var extraBytes []byte
	if exported.extraFields != nil {
		extraBytes = utils.JsonMarshal(exported.extraFields)
	}

	r.client.execute(&alioth.RestorationCollectionRequest{
		CallerService:  exported.service,
		CodePath:       exported.code,
		Level:          exported.level,
		Message:        exported.message,
		CalledAt:       exported.calledAt,
		CalledFunction: exported.calledFunction,
		TraceId:        exported.traceID,
		InputFields:    paramsBytes,
		PayloadFields:  processingBytes,
		ExtraFields:    extraBytes,
	})
}

func (r *collector) Debug(fields Fields) {
	go r.logField(fields.withLevel("debug").withService(r.serviceName))
}

func (r *collector) Info(fields Fields) {
	go r.logField(fields.withLevel("info").withService(r.serviceName))
}

func (r *collector) Warn(fields Fields) {
	go r.logField(fields.withLevel("warn").withService(r.serviceName))
}

func (r *collector) Error(fields Fields) {
	go r.logField(fields.withLevel("error").withService(r.serviceName))
}

// NewCollector 创建一个新的日志收集器
//   - serviceName: 服务名称，用于标识日志来源
//   - restorationAddr: 日志服务地址，需要包含IP和端口，如 10.0.0.1:50051
//
// 不会验证 serverAddr 的有效性，失败了也不会有任何提示，哪怕是失败日志也不会有
func NewCollector(serviceName string, restorationAddr string) (c Collector, err error) {
	if rpcClient, initClientErr := newRestorationClient(restorationAddr); initClientErr != nil {
		return nilCollector, fmt.Errorf("init restoration client error: %w", initClientErr)
	} else {
		return &collector{
			client:      rpcClient,
			serviceName: serviceName,
		}, nil
	}
}

// NewCollectorWithFailedCallback 创建一个新的日志收集器，当日志收集器无法连接到日志服务时，会调用回调函数
//   - serviceName: 服务名称，用于标识日志来源
//   - restorationAddr: 日志服务地址，需要包含IP和端口，如 10.0.0.1:50051
//   - maxFailed: 最大失败次数，如果失败次数超过这个值，则会调用 callback，大于 0 时生效
//   - callback: 调用的回调函数，不为 nil 时生效
//
// 不会验证 serverAddr 的有效性，失败了也不会有任何提示，哪怕是失败日志也不会有，在超过最大失败次数后会调用 callback
func NewCollectorWithFailedCallback(serviceName string, restorationAddr string, maxFailed int, callback func(err error)) (c Collector, err error) {
	if rpcClient, initClientErr := newClientWithFailedCallback(restorationAddr, maxFailed, callback); initClientErr != nil {
		return nilCollector, fmt.Errorf("init restoration client error: %w", initClientErr)
	} else {
		return &collector{
			client:      rpcClient,
			serviceName: serviceName,
		}, nil
	}
}

type externalCollector struct {
	client      *externalClient
	serviceName string
}

func (r *externalCollector) logField(fields Fields) {
	exported := fields.export()
	if exported.service == "" {
		exported.service = r.serviceName
	}

	var paramsBytes []byte
	if exported.inputFields != nil {
		paramsBytes = utils.JsonMarshal(exported.inputFields)
	}

	var processingBytes []byte
	if exported.payloadFields != nil {
		processingBytes = utils.JsonMarshal(exported.payloadFields)
	}

	var extraBytes []byte
	if exported.extraFields != nil {
		extraBytes = utils.JsonMarshal(exported.extraFields)
	}

	logExternalErr := r.client.execute(&alioth.RestorationCollectionRequest{
		CallerService:  exported.service,
		CodePath:       exported.code,
		Level:          exported.level,
		Message:        exported.message,
		CalledAt:       exported.calledAt,
		CalledFunction: exported.calledFunction,
		TraceId:        exported.traceID,
		InputFields:    paramsBytes,
		PayloadFields:  processingBytes,
		ExtraFields:    extraBytes,
	})

	if logExternalErr != nil {
		logger.Log(log.DefaultField().WithMessage("log external error").WithLevel(log.Error).WithCaller(log.Module).
			WithExtra(map[string]any{"error": logExternalErr, "fields": fields}))
	}
}

func (r *externalCollector) Debug(fields Fields) {
	go r.logField(fields.withLevel("debug").withService(r.serviceName))
}

func (r *externalCollector) Info(fields Fields) {
	go r.logField(fields.withLevel("info").withService(r.serviceName))
}

func (r *externalCollector) Warn(fields Fields) {
	go r.logField(fields.withLevel("warn").withService(r.serviceName))
}

func (r *externalCollector) Error(fields Fields) {
	go r.logField(fields.withLevel("error").withService(r.serviceName))
}

// NewExternalCollector 创建一个新的日志收集器
//   - serviceName: 服务名称，用于标识日志来源
//   - restorationAddr: 日志服务地址，需要包含协议和端口，如 https://api.sunist.work:8080
//
// 会请求 ${endpoint}/restoration/ping 接口，如果返回 200 则认为 endpoint 有效，否则返回 error
func NewExternalCollector(serviceName string, restorationAddr string) (collector Collector, err error) {
	if httpClient, initClientErr := newExternalClient(restorationAddr); initClientErr != nil {
		return nilCollector, fmt.Errorf("init restoration client error: %w", initClientErr)
	} else {
		return &externalCollector{
			client:      httpClient,
			serviceName: serviceName,
		}, nil
	}
}
