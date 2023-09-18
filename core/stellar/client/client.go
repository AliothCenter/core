package stellar

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"studio.sunist.work/platform/alioth-center/infrastructure/initialize"
	"studio.sunist.work/platform/alioth-center/infrastructure/utils/version"
	"studio.sunist.work/platform/alioth-center/proto/alioth"
)

// Client stellar 客户端接口
type Client interface {
	// Register 注册服务
	//   - service: 服务名称
	//   - version: 服务版本
	//   - port: 服务端口
	Register(service string, version version.Version, port int) (address string, handler string, err error)

	// Discovery 发现服务
	//   - service: 服务名称
	//   - minVersion: 最小版本
	Discovery(service string, minVersion version.Version) (address string, handler string, err error)

	// Unmount 卸载服务
	//   - service: 服务名称
	//   - handler: 服务处理器名称
	Unmount(service string, handler string) (err error)
}

// client stellar 客户端，使用 rpc 协议
type client struct {
	conn alioth.AliothStellarClient
}

func (c *client) Register(service string, version version.Version, port int) (address string, handler string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(initialize.GlobalConfig().Grpc.TimeoutSeconds)*time.Second)
	response, executeErr := c.conn.ServiceRegistration(ctx, &alioth.ServiceRegistrationRequest{
		Service: service,
		Port:    int32(port),
		Version: version.Export(),
	})
	cancel()

	if executeErr != nil {
		return "", "", fmt.Errorf("failed to register service: %w", executeErr)
	} else {
		return response.GetAddress(), response.GetName(), nil
	}
}

func (c *client) Discovery(service string, minVersion version.Version) (address string, handler string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(initialize.GlobalConfig().Grpc.TimeoutSeconds)*time.Second)
	response, executeErr := c.conn.ServiceDiscovery(ctx, &alioth.ServiceDiscoveryRequest{
		Service:    service,
		MinVersion: minVersion.Export(),
	})
	cancel()

	if executeErr != nil {
		return "", "", fmt.Errorf("failed to discovery service: %w", executeErr)
	} else {
		return response.GetAddress(), response.GetName(), nil
	}
}

func (c *client) Unmount(service string, handler string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(initialize.GlobalConfig().Grpc.TimeoutSeconds)*time.Second)
	_, executeErr := c.conn.ServiceUnmount(ctx, &alioth.ServiceUnmountRequest{
		Service: service,
		Name:    handler,
	})
	cancel()

	if executeErr != nil {
		return fmt.Errorf("failed to unmount service: %w", executeErr)
	} else {
		return nil
	}
}

// NewClient 创建一个使用 rpc 协议的 stellar 客户端
//   - serverAddr: stellar 服务的地址，需要包含IP和端口，如
//
// 不会验证 serverAddr 的有效性，失败了就返回错误
func NewClient(serverAddr string) (c Client, err error) {
	if conn, dialErr := grpc.Dial(serverAddr, grpc.WithCredentialsBundle(insecure.NewBundle())); dialErr != nil {
		return nil, fmt.Errorf("failed to dial grpc client: %w", dialErr)
	} else {
		clt := alioth.NewAliothStellarClient(conn)
		return &client{conn: clt}, nil
	}
}
