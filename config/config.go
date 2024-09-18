package config

import (
	"standard-library/utility"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
)

// Use when ever data is in app.conf
var (
	HttpPort         int64
	NacosPort        int64
	NacosUrl         string
	NacosNamespaceId string
	NacosDataId      string
	NacosGroupId     string
)

func init() {
	HttpPort = utility.StringToInt64(getValue("HttpPort"))
	NacosPort = utility.StringToInt64(getValue("NacosPort"))
	NacosUrl = getValue("NacosUrl")
	NacosNamespaceId = getValue("NacosNamespaceId")
	NacosDataId = getValue("NacosDataId")
	NacosGroupId = getValue("NacosGroupId")
}

func getValue(key string) string {
	value, err := web.AppConfig.String(key)
	if err != nil {
		logs.Error("[Conf][getValue]Error", err)
		return ""
	}
	return value
}
