package config

type DatabaseConfig struct {
	Host              string `json:"host" yaml:"host"`
	Port              int    `json:"port" yaml:"port"`
	Username          string `json:"username" yaml:"username"`
	Password          string `json:"password" yaml:"password"`
	Database          string `json:"database" yaml:"database"`
	MaxIdle           int    `json:"max_idle" yaml:"max_idle"`
	MaxOpen           int    `json:"max_open" yaml:"max_open"`
	SqlLogger         string `json:"sql_logger" yaml:"sql_logger"`
	DisablePrepareSql bool   `json:"disable_prepare_sql" yaml:"disable_prepare_sql"`
	DebugMode         bool   `json:"debug_mode" yaml:"debug_mode"`
	AllowGlobalUpdate bool   `json:"allow_global_update" yaml:"allow_global_update"`
	TranslateError    bool   `json:"translate_error" yaml:"translate_error"`
	SyncDaoModels     bool   `json:"sync_dao_models" yaml:"sync_dao_models"`
}
