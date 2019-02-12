package monitor

import (
	"encoding/json"
)

type Version struct {
	GitTag    string `json:"git_tag"`
	GitHash   string `json:"git_hash"`
	BuildTime string `json:"build_time"`
}

var (
	globalVersion    Version
	versionJsonCache []byte
)

func initVersion(v Version) {
	globalVersion = v
	bs, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	versionJsonCache = bs
}
