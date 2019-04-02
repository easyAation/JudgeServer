package common

import (
	"github.com/BurntSushi/toml"
)

var Config Configs

type Configs struct {
	Listen    int
	AllowCORS bool
	Compile   CompileConfig
}

type CompileConfig struct {
	CodeDir string
	ExeDir  string
}

func InitConfig(fpath string) {
	config, err := loadConfig(fpath)
	if err != nil {
		panic(err)
	}
	// TODO 校验配置文件合法性
	Config = *config
}

func loadConfig(fpath string) (*Configs, error) {
	var config = new(Configs)
	if _, err := toml.DecodeFile(fpath, config); err != nil {
		return nil, err
	}
	return config, nil
}
