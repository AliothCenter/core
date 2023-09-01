package config

type GlobalConfig struct {
	Database DatabaseConfig `json:"database" yaml:"database"`
	Grpc     GrpcConfig     `json:"grpc" yaml:"grpc"`
	Http     HttpConfig     `json:"http" yaml:"http"`
}
