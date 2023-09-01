package config

type HttpConfig struct {
	ListenIP       string `json:"listen_ip" yaml:"listen_ip"`
	ListenPort     int    `json:"listen_port" yaml:"listen_port"`
	TimeoutSeconds int    `json:"timeout" yaml:"timeout_seconds"`
}
