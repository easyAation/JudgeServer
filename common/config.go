package common

import (
	"github.com/BurntSushi/toml"
	"github.com/easyAation/scaffold/db"
	"time"
)

var Config Configs

type Configs struct {
	Listen    int
	AllowCORS bool
	MySQL     db.MySQLConfig
	Redis     db.RedisConfig
	Compile   CompileConfig
	SandBox   SandBoxConfig
	Token     TokenConfig
}

type CompileConfig struct {
	CodeDir string
	ExeDir  string
}

type SandBoxConfig struct {
	Exe        string
	ProblemDir string
	OutPutDir  string
}

type TokenConfig struct {
	Expiration Duration
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

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalText(text []byte) (err error) {
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

func (d *Duration) D() time.Duration {
	return d.Duration
}
