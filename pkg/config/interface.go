package config

import (
	"github.com/spf13/viper"
)

// Create mock using:
// mockgen -source=pkg/config/interface.go -destination=pkg/config/mock/mock_config.go
type Interface interface {
	Init() error
	Set(key string, value interface{})
	SetDefault(key string, value interface{})
	AllSettings() map[string]interface{}
	IsSet(key string) bool
	Get(key string) interface{}
	GetBool(key string) bool
	GetInt(key string) int
	GetString(key string) string
	GetStringSlice(key string) []string
	UnmarshalKey(key string, rawVal interface{}, decoder ...viper.DecoderConfigOption) error
	GetBase64Decoded(key string) (string, error)
	ReadConfig(configFilePath string) error
}

const PACKAGR_PACKAGE_TYPE = "package_type"
const PACKAGR_SCM = "scm"
const PACKAGR_VERSION_METADATA_PATH = "version_metadata_path"
const PACKAGR_VERSION_BUMP_MESSAGE = "engine_version_bump_msg"
const PACKAGR_GENERIC_VERSION_TEMPLATE = "generic_version_template"
const PACKAGR_GIT_AUTHOR_NAME = "engine_git_author_name"
const PACKAGR_GIT_AUTHOR_EMAIL = "engine_git_author_email"
const PACKAGR_ENGINE_REPO_CONFIG_PATH = "engine_repo_config_path"
const PACKAGR_MGR_KEEP_LOCK_FILE = "mgr_keep_lock_file"
