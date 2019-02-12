package oss

import (
	"fmt"

	"online_judge/talcity/scaffold/util"
)

// GenerateObjectName
// serviceName, eg. forum, dataset...
// forum/uuid/x.png
func GenerateObjectName(fileName, serviceName string) string {
	return fmt.Sprintf("%s/%s/%s", serviceName, util.UUID(), fileName)
}
