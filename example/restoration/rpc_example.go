package main

import (
	"context"
	"fmt"
	restoration "studio.sunist.work/platform/alioth-center/core/restoration/client"
	stellar "studio.sunist.work/platform/alioth-center/core/stellar/client"
	"studio.sunist.work/platform/alioth-center/infrastructure/utils"
	"studio.sunist.work/platform/alioth-center/infrastructure/utils/version"
	"time"
)

type RpcExampleStruct struct {
	IntArray     []int             `json:"int_array"`
	BoolVar      bool              `json:"bool_var"`
	MapVar       map[string]string `json:"map_var"`
	SubStructure struct {
		Hello string `json:"hello"`
	} `json:"sub_structure"`
}

func main() {
	// 初始化一个 stellar 客户端
	client, initStellarErr := stellar.NewClient("127.0.0.1:50051")
	if initStellarErr != nil {
		panic(initStellarErr)
	}

	// 从stellar获取一个 restoration 服务的地址
	address, _, discoveryErr := client.Discovery("alioth-restoration", version.NewVersion(1, 0, 0, 0))
	if discoveryErr != nil {
		panic(discoveryErr)
	}

	// 添加一个自定义的收集器，将所有日志记录的错误打印
	collector, initCollectorErr := restoration.NewCollectorWithFailedCallback("rpc-example", address,
		1, func(err error) {
			fmt.Println(err)
		})
	if initCollectorErr != nil {
		panic(initCollectorErr)
	}

	// 需要发送的结构体
	exampleStructure := RpcExampleStruct{
		IntArray: []int{1, 2, 3},
		BoolVar:  true,
		MapVar: map[string]string{
			"hello": "world",
			"foo":   "bar",
		},
		SubStructure: struct {
			Hello string `json:"hello"`
		}{
			Hello: "world",
		},
	}

	// 打印日志
	ctx := utils.AddTraceID(context.Background())
	collector.Debug(restoration.NewCollection(ctx, "hello, world"))
	collector.Info(restoration.NewCollection(ctx, "hello, world").WithParams(exampleStructure))
	collector.Warn(restoration.NewCollection(ctx, "hello, world").WithProcessing(exampleStructure))
	collector.Error(restoration.NewCollection(ctx, "hello, world").WithExtra(exampleStructure))

	time.Sleep(time.Second)
}

// 打印结果：
// restoration/stdout:
// {"called_at":"20xx.xx.xx-xx:36:59.926+08:00","called_function":"main.main","caller_ip":"127.0.0.1","caller_service":"rpc-example","caller_type":"service","code_path":"/path/to/restoration/rpc_example.go:45","level":"debug","msg":"hello, world","time":"2023-09-01T12:36:59+08:00","trace_id":"a103dea9-4744-4fa4-8836-1bf0758cb3e4"}
// {"called_at":"20xx.xx.xx-xx:36:59.927+08:00","called_function":"main.main","caller_ip":"127.0.0.1","caller_service":"rpc-example","caller_type":"service","code_path":"/path/to/restoration/rpc_example.go:46","extra_data":{"bool_var":true,"int_array":[1,2,3],"map_var":{"foo":"bar","hello":"world"},"sub_structure":{"hello":"world"}},"level":"info","msg":"hello, world","time":"20xx-xx-xxTxx:36:59+08:00","trace_id":"a103dea9-4744-4fa4-8836-1bf0758cb3e4"}
// {"called_at":"20xx.xx.xx-xx:36:59.928+08:00","called_function":"main.main","caller_ip":"127.0.0.1","caller_processing":{"bool_var":true,"int_array":[1,2,3],"map_var":{"foo":"bar","hello":"world"},"sub_structure":{"hello":"world"}},"caller_service":"rpc-example","caller_type":"service","code_path":"/path/to/restoration/rpc_example.go:47","level":"warning","msg":"hello, world","time":"20xx-xx-xxTxx:36:59+08:00","trace_id":"a103dea9-4744-4fa4-8836-1bf0758cb3e4"}
// {"called_at":"20xx.xx.xx-xx:36:59.928+08:00","called_function":"main.main","caller_ip":"127.0.0.1","caller_service":"rpc-example","caller_type":"service","code_path":"/path/to/restoration/rpc_example.go:48","extra_data":{"bool_var":true,"int_array":[1,2,3],"map_var":{"foo":"bar","hello":"world"},"sub_structure":{"hello":"world"}},"level":"error","msg":"hello, world","time":"20xx-xx-xxTxx:36:59+08:00","trace_id":"a103dea9-4744-4fa4-8836-1bf0758cb3e4"}
// restoration/stderr:
// {"called_at":"20xx.xx.xx-xx:36:59.928+08:00","called_function":"main.main","caller_ip":"127.0.0.1","caller_service":"rpc-example","caller_type":"service","code_path":"/path/to/restoration/rpc_example.go:48","extra_data":{"bool_var":true,"int_array":[1,2,3],"map_var":{"foo":"bar","hello":"world"},"sub_structure":{"hello":"world"}},"level":"error","msg":"hello, world","time":"20xx-xx-xxTxx:36:59+08:00","trace_id":"a103dea9-4744-4fa4-8836-1bf0758cb3e4"}
