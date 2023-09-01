package errors

import "fmt"

type GetRPCClientIPFailedError struct {
	basicAliothError
}

func (e *GetRPCClientIPFailedError) Error() string {
	return "get rpc client ip failed"
}

func NewGetRPCClientIPFailedError() AliothError {
	return &GetRPCClientIPFailedError{}
}

type UnsupportedNetworkError struct {
	basicAliothError
	network string
}

func (e *UnsupportedNetworkError) Error() string {
	return fmt.Sprintf("unsupported network: %s", e.network)
}

func NewUnsupportedNetworkError(network string) AliothError {
	return &UnsupportedNetworkError{
		network: network,
	}
}

type InvalidIPAddressError struct {
	basicAliothError
	ip string
}

func (e *InvalidIPAddressError) Error() string {
	return fmt.Sprintf("invalid ip address: %s", e.ip)
}

func NewInvalidIPAddressError(ip string) AliothError {
	return &InvalidIPAddressError{
		ip: ip,
	}
}

type InvalidTraceIDError struct {
	basicAliothError
}

func (e *InvalidTraceIDError) Error() string {
	return "invalid trace id"
}

func NewInvalidTraceIDError() AliothError {
	return &InvalidTraceIDError{}
}

type InvalidVersionError struct {
	basicAliothError
	version string
	err     error
}

func (e *InvalidVersionError) Error() string {
	return fmt.Errorf("invalid version [%s]: %w", e.version, e.err).Error()
}

func NewInvalidVersionError(version string, err error) AliothError {
	return &InvalidVersionError{
		version: version,
		err:     err,
	}
}
