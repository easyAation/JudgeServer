package loadconfig

import (
	"github.com/koding/multiconfig"
)

type config interface {
}

// LoadConfig toml path + config struct
func Load(path string, conf interface{}) {
	m := multiconfig.NewWithPath(path)
	m.MustLoad(conf)
}
