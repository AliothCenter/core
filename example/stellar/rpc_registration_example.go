package stellar

import (
	"fmt"
	stellar "studio.sunist.work/platform/alioth-center/core/stellar/client"
	"studio.sunist.work/platform/alioth-center/infrastructure/utils/exit"
	"studio.sunist.work/platform/alioth-center/infrastructure/utils/version"
)

func main() {
	client, initClientErr := stellar.NewClient("127.0.0.1:50051")
	if initClientErr != nil {
		panic(initClientErr)
	}

	_, handlerName, registerErr := client.Register("alioth-example", version.AlphaVersion, 50052)
	if registerErr != nil {
		panic(registerErr)
	}

	fmt.Println("成功注册服务 alioth-example: ", handlerName)

	// 根据情况，退出时卸载服务
	// removeInstance("alioth-example", handlerName, client)
}

// 退出时卸载服务
func removeInstance(serviceName, handlerName string, client stellar.Client) {
	exit.AddExitFunctions(func() error {
		err := client.Unmount(serviceName, handlerName)
		if err != nil {
			return err
		}
		return nil
	})
}
