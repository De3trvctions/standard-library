package initilize

import (
	"api-login/config"
	"api-login/crontask"
	"api-login/db"
	"api-login/mail"
	"api-login/nacos"
	"api-login/redis"
	"strings"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/i18n"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
)

func InitLogs() {
	logs.SetLogger("console", `{"level":7,"color":true}`) // Set level to Trace for maximum verbosity
	logs.EnableFuncCallDepth(true)                        // Enable func call depth to display file and line numbers
	logs.SetLogFuncCallDepth(3)
	logs.Info("[InitLogs] Init Logs Success")
}

func InitRedis() {
	redis.InitRedis(nacos.RedisAddr, nacos.RedisPort)
}

func InitDB() {
	syncDB := true
	db.InitDB(syncDB)
}

func InitLanguage() {
	langs := nacos.Lang
	langTypes := strings.Split(langs, "|")
	for _, lang := range langTypes {
		if lang != "" {
			logs.Info("[InitLanguage] Initialize language: ", lang)
			if err := i18n.SetMessage(lang, "conf/locale_"+lang+".ini"); err != nil {
				logs.Error("[InitLanguage] Fail to set message file:", err)
			}
		}
	}
	logs.Info("[InitLanguage] Init Language Success")
}

func InitMail(option ...*mail.Option) {
	if len(option) >= 1 {
		logs.Error("1")
		mail.New(option...)
	} else {
		mail.New(nacos.Mail...)
	}
}

func InitNacosConfig() {
	// Create a client config
	clientConfig := constant.ClientConfig{
		NamespaceId:         config.NacosNamespaceId, // namespaceId
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
		LogLevel:            "debug",
	}

	// Create a server config
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr:      config.NacosUrl, // Nacos server IP
			ContextPath: "/nacos",
			Port:        uint64(config.NacosPort),
			Scheme:      "http",
		},
	}

	// Create a config client
	configClient, err := clients.CreateConfigClient(map[string]interface{}{
		"clientConfig":  clientConfig,
		"serverConfigs": serverConfigs,
	})
	if err != nil {
		logs.Error("[InitNacosConfig] Init Nacos error 1", err)
	}

	err = nacos.SyncConf(configClient, config.NacosDataId, config.NacosGroupId)
	if err != nil {
		logs.Error("[InitNacosConfig] Init Nacos error 2", err)
	}
	logs.Info("[InitLanguage] Init Nacos Success")
}

func InitCron() {
	crontask.InitCronTask()
}
