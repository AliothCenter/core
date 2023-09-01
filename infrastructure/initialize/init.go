package initialize

import (
	"flag"
	"io"
	"os"

	"studio.sunist.work/platform/alioth-center/infrastructure/global/errors"
	"studio.sunist.work/platform/alioth-center/infrastructure/global/utils/logger"
	"studio.sunist.work/platform/alioth-center/infrastructure/initialize/config"

	"gopkg.in/yaml.v3"
)

var (
	configFile = "example/config.yaml"
	logger     = log.DefaultLogger()
	globalConf = config.GlobalConfig{}
)

func GlobalConfig() config.GlobalConfig {
	return globalConf
}

func init() {
	flag.StringVar(&configFile, "config", "example/config.yaml", "directory of config files")
	flag.Parsed()

	// 读取配置文件
	if file, openFileErr := os.OpenFile(configFile, os.O_RDONLY, 0o666); openFileErr != nil {
		panicErr := errors.NewOpenConfigFileInitializeError(configFile, openFileErr)
		logger.Log(log.DefaultField().WithCaller(log.Internal).WithLevel(log.Panic).
			WithMessage("failed to open config file").WithExtra(panicErr))
	} else if bytesOfConfig, readConfigErr := io.ReadAll(file); readConfigErr != nil {
		panicErr := errors.NewReadConfigFileInitializeError(configFile, readConfigErr)
		logger.Log(log.DefaultField().WithCaller(log.Internal).WithLevel(log.Panic).
			WithMessage("failed to read config file").WithExtra(panicErr))
	} else if unmarshalConfigErr := yaml.Unmarshal(bytesOfConfig, &globalConf); unmarshalConfigErr != nil {
		panicErr := errors.NewReadConfigFileInitializeError(configFile, unmarshalConfigErr)
		logger.Log(log.DefaultField().WithCaller(log.Internal).WithLevel(log.Panic).
			WithMessage("failed to unmarshal config file").WithExtra(panicErr))
	}
}
