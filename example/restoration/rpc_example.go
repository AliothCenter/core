package main

import (
	"context"
	"fmt"
	sdk "studio.sunist.work/platform/alioth-center/core/restoration/client"
	"studio.sunist.work/platform/alioth-center/infrastructure/global/utils"
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
	// 添加一个自定义的收集器，将所有日志记录的错误打印
	collector, err := sdk.NewCollectorWithFailedCallback("rpc-example", "127.0.0.1:50051",
		1, func(err error) {
			fmt.Println(err)
		})
	if err != nil {
		panic(err)
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
	collector.Debug(sdk.NewCollection(ctx, "hello, world"))
	collector.Info(sdk.NewCollection(ctx, "hello, world").WithParams(exampleStructure))
	collector.Warn(sdk.NewCollection(ctx, "hello, world").WithProcessing(exampleStructure))
	collector.Error(sdk.NewCollection(ctx, "hello, world").WithExtra(exampleStructure))

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
