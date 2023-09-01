package utils

import (
	"context"
	"net"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/grpc/peer"

	"studio.sunist.work/platform/alioth-center/infrastructure/global/errors"
)

type keyType string

const traceIDKey keyType = "trace_id"

// GetContextClientIP 获取上下文中的客户端IP
func GetContextClientIP(ctx context.Context) (ip string, err error) {
	if pr, getPrSuccess := peer.FromContext(ctx); !getPrSuccess {
		// 如果获取失败，则返回错误
		return "", errors.NewGetRPCClientIPFailedError()
	} else if pr.Addr == net.Addr(nil) {
		// 如果获取到的是空接口，则返回错误
		return "", errors.NewGetRPCClientIPFailedError()
	} else if pr.Addr.Network() != "tcp" {
		// 如果获取到的不是tcp协议，则返回错误
		return "", errors.NewUnsupportedNetworkError(pr.Addr.Network())
	} else if ipSlice := strings.Split(pr.Addr.String(), ":"); len(ipSlice) != 2 {
		// 如果获取到的不是ip:port(ipv4)格式，则返回错误
		return "", errors.NewInvalidIPAddressError(pr.Addr.String())
	} else {
		// 如果获取到的是ip:port(ipv4)格式，则返回ip
		return ipSlice[0], nil
	}
}

// GetTraceID 获取上下文中的trace_id
func GetTraceID(ctx context.Context) (traceID string, err error) {
	if ctx == nil {
		return "", errors.NewInvalidTraceIDError()
	} else if id, convertTraceIDSuccess := ctx.Value("trace_id").(string); !convertTraceIDSuccess {
		return "", errors.NewInvalidTraceIDError()
	} else {
		return id, nil
	}
}

// AddTraceID 添加trace_id到上下文中
func AddTraceID(ctx context.Context) (newCtx context.Context) {
	if _, err := GetTraceID(ctx); err != nil {
		// 如果不存在trace_id，则生成一个
		return context.WithValue(ctx, traceIDKey, uuid.NewString())
	} else {
		// 如果存在trace_id，则直接返回
		return ctx
	}
}
