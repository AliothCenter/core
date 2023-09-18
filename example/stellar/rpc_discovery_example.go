package stellar

import (
	"fmt"
	stellar "studio.sunist.work/platform/alioth-center/core/stellar/client"
	"studio.sunist.work/platform/alioth-center/infrastructure/utils/version"
)

func main() {
	// 获取一个stellar客户端
	client, initStellarErr := stellar.NewClient("127.0.0.1:50051")
	if initStellarErr != nil {
		panic(initStellarErr)
	}

	// 从stellar获取一个服务的地址
	serviceName := "alioth-starward"
	rpcAddress, handlerName, discoveryErr := client.Discovery(serviceName, version.NewVersion(1, 0, 0, 0))
	if discoveryErr != nil {
		panic(discoveryErr)
	} else {
		fmt.Println("rpc address:", rpcAddress)
		fmt.Println("handler name:", handlerName)
	}

	// 调用服务
	// grpcClient, dialErr := grpc.Dial(rpcAddress, grpc.WithCredentialsBundle(insecure.NewBundle()))
	// if dialErr != nil {
	//     panic(dialErr)
	// }
	// c := proto.NewServiceClient(grpcClient)
	// c.HandleSomething(...)
	// ...
}
