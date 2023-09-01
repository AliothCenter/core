package restoration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"studio.sunist.work/platform/alioth-center/infrastructure/initialize"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"studio.sunist.work/platform/alioth-center/infrastructure/global/errors"
	"studio.sunist.work/platform/alioth-center/proto/alioth"
)

// client restoration 客户端，使用 grpc 协议
type client struct {
	failed     int
	maxFailed  int
	failedCall func(err error)
	conn       alioth.AliothRestorationClient
}

func (c *client) execute(request *alioth.RestorationCollectionRequest) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(initialize.GlobalConfig().Grpc.TimeoutSeconds)*time.Second)
	_, e := c.conn.RestorationCollection(ctx, request)
	cancel()

	if e != nil {
		c.failed++
		if c.failed >= c.maxFailed && c.failedCall != nil {
			c.failedCall(e)
			c.failed = 0
		}
	}
}

// newRestorationClient 创建一个使用 grpc 协议的 restoration 客户端
//   - serverAddr: restoration 服务的地址，需要包含IP和端口，如 10.0.0.1:50051
//
// 不会验证 serverAddr 的有效性，失败了也不会有任何提示，哪怕是失败日志也不会有
func newRestorationClient(serverAddr string) (c *client, err error) {
	if conn, dialErr := grpc.Dial(serverAddr, grpc.WithCredentialsBundle(insecure.NewBundle())); dialErr != nil {
		return nil, fmt.Errorf("failed to dial grpc client: %v", dialErr)
	} else {
		clt := alioth.NewAliothRestorationClient(conn)
		return &client{conn: clt, maxFailed: 1 << 31, failedCall: func(error) {}}, nil
	}
}

// newClientWithFailedCallback 创建一个使用 grpc 协议的 restoration 客户端
//   - serverAddr: restoration 服务的地址，需要包含IP和端口，如 10.0.0.1:50051
//   - maxFailed: 最大失败次数，如果失败次数超过这个值，则会调用 callback，大于 0 时生效
//   - callback: 调用的回调函数，不为 nil 时生效
//
// 不会验证 serverAddr 的有效性，失败了也不会有任何提示，哪怕是失败日志也不会有，在超过最大失败次数后会调用 callback
func newClientWithFailedCallback(serviceAddr string, maxFailed int, callback func(err error)) (client *client, err error) {
	restorationClient, initClientErr := newRestorationClient(serviceAddr)
	if initClientErr != nil {
		return nil, initClientErr
	} else {
		if maxFailed > 0 {
			restorationClient.maxFailed = maxFailed
		}
		if callback != nil {
			restorationClient.failedCall = callback
		}

		return restorationClient, nil
	}
}

// externalClient restoration 客户端，使用 http 协议
type externalClient struct {
	endpoint string
}

func (c externalClient) execute(request *alioth.RestorationCollectionRequest) (err error) {
	// 序列化请求
	payload, marshalErr := json.Marshal(request)
	if marshalErr != nil {
		return fmt.Errorf("failed to marshal request: %w", marshalErr)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(initialize.GlobalConfig().Http.TimeoutSeconds)*time.Second)
	defer cancel()

	// 构建http请求
	httpRequest, buildHttpRequestErr := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewBuffer(payload))
	if buildHttpRequestErr != nil {
		return fmt.Errorf("failed to build http request: %w", buildHttpRequestErr)
	}

	// 发送http请求
	httpResponse, sendHttpRequestErr := http.DefaultClient.Do(httpRequest)
	if sendHttpRequestErr != nil {
		return fmt.Errorf("failed to send http request: %w", sendHttpRequestErr)
	}

	// 检查http响应
	if httpResponse.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send http request: %w", errors.NewRestorationExternalResponseError(httpResponse.StatusCode))
	}

	return nil
}

// newExternalClient 创建一个使用 http 协议的 restoration 客户端，会验证 endpoint 的有效性
//   - endpoint: restoration 服务的地址，需要包含协议和端口，如 https://api.sunist.work:8080
//
// 会请求 ${endpoint}/restoration/ping 接口，如果返回 200 则认为 endpoint 有效
func newExternalClient(endpoint string) (client *externalClient, err error) {
	// 验证endpoint
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(initialize.GlobalConfig().Http.TimeoutSeconds)*time.Second)
	defer cancel()
	path, joinPathErr := url.JoinPath(endpoint, "/restoration/ping")
	collection, collectionPathErr := url.JoinPath(endpoint, "/restoration/collection")
	if joinPathErr != nil {
		return nil, fmt.Errorf("failed to find endpoint: %w", joinPathErr)
	}
	if collectionPathErr != nil {
		return nil, fmt.Errorf("failed to find endpoint: %w", collectionPathErr)
	}

	httpRequest, buildHttpRequestErr := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if buildHttpRequestErr != nil {
		return nil, fmt.Errorf("failed to build http request: %w", buildHttpRequestErr)
	}

	httpResponse, sendHttpRequestErr := http.DefaultClient.Do(httpRequest)
	if sendHttpRequestErr != nil {
		return nil, fmt.Errorf("failed to send http request: %w", sendHttpRequestErr)
	}

	if httpResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to send http request: %w", errors.NewRestorationExternalResponseError(httpResponse.StatusCode))
	}

	return &externalClient{endpoint: collection}, nil
}
