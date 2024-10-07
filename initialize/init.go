package initilize

import (
	"fmt"
	"log"
	"net"
	"standard-library/config"
	"standard-library/db"
	"standard-library/grpc"
	"standard-library/mail"
	"standard-library/nacos"
	"standard-library/redis"
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

func RunGRPC(srv *grpc.Server) {
	listener, err := net.Listen("tcp", fmt.Sprint(":", config.HttpPort))
	if err != nil {
		log.Panicf("GRPC service listen failed %s\n", err.Error())
	}
	logs.Warn("server Running on http://:%d", config.HttpPort)
	err = srv.Srv.Serve(listener)
	if err != nil {
		log.Panicf("GRPC service start failed %s\n", err.Error())
	}
}

// InitGRPC 初始化GRPC连接池
// srvName添加旧版服务发现兼容配置，全部转换后删除-(02-13)
func InitGRPC() {
	for serviceName, address := range nacos.Service {
		go func(serviceName, address string) {
			if err := grpc.Register(serviceName, address, nacos.GRPC.Copy()); err != nil {
				logs.Error("[config.Service]InitGRPC Service <%s> Address <%s> failed register,Error:<%s>", serviceName, address, err.Error())
				return
			} else {
				logs.Info("[config.Service]InitGRPC Service <%s> Address <%s> success register", serviceName, address)
			}
		}(serviceName, address)
	}
}
