package initilize

import (
	"fmt"
	"log"
	"net"
	"standard-library/config"
	"standard-library/db"
	es "standard-library/elasticsearch"
	"standard-library/grpc"
	"standard-library/mail"
	"standard-library/nacos"
	"standard-library/redis"
	"strconv"
	"strings"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/i18n"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

func InitLogs() {
	logs.SetLogger(logs.AdapterMultiFile, `{"filename":"./logs/system.log","separate":["emergency", "alert", "critical", "error", "warning", "notice", "info", "debug"]}`)
	logs.SetLogger(logs.AdapterConsole, `{"level":7,"color":true}`) // Set level to Trace for maximum verbosity
	logs.EnableFuncCallDepth(true)                                  // Enable func call depth to display file and line numbers
	logs.SetLogFuncCallDepth(3)

	logs.Info("[InitLogs] Init Logs Success")
}

func InitES() {
	// Register the custom adapter with Beego logs
	logs.Register("elasticsearch", es.NewElasticsearchLogger)

	// Set Beego logs to use the custom Elasticsearch adapter
	logs.SetLogger(logs.AdapterConsole) // Optional: Also log to console
	logs.SetLogger("elasticsearch")     // Use the custom Elasticsearch adapter

	// Example log message
	logs.Info("This is a test log message that will be sent to Elasticsearch 8.6.0")
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
	// Initialize Nacos Naming Client
	if err := nacos.InitNacosClient(); err != nil {
		logs.Error("[InitNacosConfig] Failed to initialize Nacos Naming Client:", err)
		return
	}

	// Create a Config Client for fetching Nacos configurations
	clientConfig := constant.ClientConfig{
		NamespaceId:         config.NacosNamespaceId,
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
		LogLevel:            "debug",
	}

	serverConfigs := []constant.ServerConfig{
		{
			IpAddr:      config.NacosUrl,
			Port:        uint64(config.NacosPort),
			ContextPath: "/nacos",
			Scheme:      "http",
		},
	}

	configClient, err := clients.CreateConfigClient(map[string]interface{}{
		"clientConfig":  clientConfig,
		"serverConfigs": serverConfigs,
	})
	if err != nil {
		logs.Error("[InitNacosConfig] Failed to create Nacos Config Client:", err)
		return
	}

	err = nacos.SyncConf(configClient, config.NacosDataId, config.NacosGroupId)
	if err != nil {
		logs.Error("[InitNacosConfig] Failed to sync Nacos config:", err)
	}

	logs.Info("[InitNacosConfig] Successfully initialized Nacos")
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
				logs.Error("[config.Service] InitGRPC Service <%s> Address <%s> failed to register, Error: <%s>", serviceName, address, err.Error())
				return
			}
			logs.Info("[config.Service] InitGRPC Service <%s> Address <%s> successfully registered", serviceName, address)

			// Register GRPC service with Nacos
			registerServiceWithNacos(serviceName, address)
		}(serviceName, address)
	}
}

func registerServiceWithNacos(serviceName, address string) {
	if nacos.NacosNamingClient == nil {
		logs.Error("[registerServiceWithNacos] NacosNamingClient is not initialized")
		return
	}

	parts := strings.Split(address, ":")
	if len(parts) != 2 {
		logs.Error("[registerServiceWithNacos] Invalid service address format: %s", address)
		return
	}
	ip := parts[0]
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		logs.Error("[registerServiceWithNacos] Invalid port in service address: %s", address)
		return
	}

	// Register service with Nacos
	success, err := nacos.NacosNamingClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          ip,
		Port:        uint64(port),
		ServiceName: serviceName,
		Weight:      1.0,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
	})

	if err != nil || !success {
		logs.Error("[registerServiceWithNacos] Failed to register service <%s> with Nacos: %v", serviceName, err)
	} else {
		logs.Info("[registerServiceWithNacos] Successfully registered service <%s> with Nacos", serviceName)
	}
}
