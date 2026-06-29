package config

import (
	"strconv"

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
	HttpPort = stringToInt64(getValue("HttpPort"))
	NacosPort = stringToInt64(getValue("NacosPort"))
	NacosUrl = getValue("NacosUrl")
	NacosNamespaceId = getValue("NacosNamespaceId")
	NacosDataId = getValue("NacosDataId")
	NacosGroupId = getValue("NacosGroupId")
}

func stringToInt64(s string) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		logs.Error("[Conf][stringToInt64]Error", err)
		panic(err)
	}
	return i
}

func getValue(key string) string {
	value, err := web.AppConfig.String(key)
	if err != nil {
		logs.Error("[Conf][getValue]Error", err)
		return ""
	}
	return value
}
