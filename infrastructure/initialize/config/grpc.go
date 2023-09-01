package config

type GrpcConfig struct {
	ListenIP       string `json:"listen_ip" yaml:"listen_ip"`
	ListenPort     int    `json:"listen_port" yaml:"listen_port"`
	TimeoutSeconds int    `json:"timeout_seconds" yaml:"timeout_seconds"`
}
