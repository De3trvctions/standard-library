package nacos

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"standard-library/config"
	"standard-library/grpc"
	"standard-library/mail"
	"strconv"
	"strings"
	"sync"

	"github.com/beego/beego/v2/core/logs"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

var settingsMap sync.Map

// Global Nacos Naming Client
var NacosNamingClient naming_client.INamingClient

var (
	Mail                  []*mail.Option
	Service               map[string]string //GRPC服务注册map[serviceName]address
	GRPC                  *grpc.Option      //GRPC配置
	Lang                  string
	TokenSalt             string
	TokenExpMinute        int64
	TokenRefreshExpMinute int64
	TokenMaxExpSecond     int64
	ValidCodeExpMinute    int64
	DBDriver              string
	DBUser                string
	DBPassword            string
	DBHost                string
	DBPort                string
	DBName                string
	RedisAddr             string
	RedisPort             string
)

// Initializes and syncs Nacos config
func SyncConf(conf config_client.IConfigClient, dataId, groupId string) (err error) {
	content, err := conf.GetConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  groupId,
	})
	if err != nil {
		logs.Error("[SyncConf] Error fetching Nacos config:", err)
		return
	}

	formatted := parseProperties(content)
	settingsMap.Store(dataId, formatted)
	setValues(dataId)

	return
}

// Initialize Nacos Naming Client
func InitNacosClient() error {
	// Create client config
	clientConfig := constant.ClientConfig{
		NamespaceId:         config.NacosNamespaceId, // Namespace ID
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
		LogLevel:            "debug",
	}

	// Create server config
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr:      config.NacosUrl, // Nacos server IP
			Port:        uint64(config.NacosPort),
			ContextPath: "/nacos",
			Scheme:      "http",
		},
	}

	// Create a Naming Client for Service Registration
	namingClient, err := clients.CreateNamingClient(map[string]interface{}{
		"clientConfig":  clientConfig,
		"serverConfigs": serverConfigs,
	})
	if err != nil {
		logs.Error("[InitNacosClient] Error creating Nacos Naming Client:", err)
		return err
	}

	NacosNamingClient = namingClient
	logs.Info("[InitNacosClient] Successfully initialized Nacos Naming Client")
	return nil
}

type Setting struct {
	m sync.Map //map[k]v  v must string
}

// 格式化Properties
func parseProperties(cfg string) (m *Setting) {
	m = &Setting{m: sync.Map{}}
	scanner := bufio.NewScanner(bytes.NewReader([]byte(cfg)))
	scanner.Buffer(make([]byte, 8192), 8192)
	for scanner.Scan() {
		b := scanner.Bytes()
		if err := scanner.Err(); err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		s := strings.TrimSpace(string(b))
		index := strings.Index(s, "=")
		if index < 0 {
			continue
		}
		key := strings.TrimSpace(s[:index])
		if len(key) == 0 {
			continue
		}
		value := strings.TrimSpace(s[index+1:])
		if len(value) == 0 {
			continue
		}
		m.m.Store(key, value)
	}
	return m
}

// GetStore 获取Store
func GetStore(dataId string) (*Setting, bool) {
	v, ok := settingsMap.Load(dataId)
	if ok {
		return v.(*Setting), ok
	}
	return nil, ok
}

func setValues(dataId string) {
	setting, ok := GetStore(dataId)
	if !ok {
		return
	}
	setting.Json("Mail", &Mail, nil)
	setting.Json("Service", &Service, `{"service-login":"localhost:55000", "service-account":"localhost:55001"}`)
	setting.Json("GRPC", &GRPC, nil)

	Lang = setting.String("Lang", "zh-CN")
	TokenSalt = setting.String("TokenSalt", "")
	TokenExpMinute = setting.Int("TokenExpMinute", 0)
	TokenRefreshExpMinute = setting.Int("TokenRefreshExpMinute", 0)
	if TokenRefreshExpMinute == 0 {
		TokenRefreshExpMinute = TokenExpMinute
	}
	TokenMaxExpSecond = setting.Int("TokenMaxExpSecond", 0)
	ValidCodeExpMinute = setting.Int("ValidCodeExpMinute", 0)
	DBDriver = setting.String("DBDriver", "mysql")
	DBUser = setting.String("DBUser", "")
	DBPassword = setting.String("DBPassword", "")
	DBHost = setting.String("DBHost", "")
	DBPort = setting.String("DBPort", "")
	DBName = setting.String("DBName", "")
	RedisAddr = setting.String("RedisAddr", "127.0.0.1")
	RedisPort = setting.String("RedisPort", "6379")
}

func (s *Setting) Strings(k string, def ...string) []string {
	v, ok := s.m.Load(k)
	if ok {
		str := v.(string)
		if len(str) == 0 {
			return def
		}
		return strings.Split(v.(string), ",")
	}
	return def
}

// String 获取值
func (s *Setting) String(k, def string) string {
	v, ok := s.m.Load(k)
	if ok {
		return v.(string)
	}
	return def
}

func (s *Setting) Json(k string, m interface{}, def interface{}) {
	v, ok := s.m.Load(k)
	if ok {
		str := v.(string)
		if len(str) == 0 {
			logs.Error("config Setting Json key:%s val:%s len = 0", k, str)
			m = def
			return
		}
		err := json.Unmarshal([]byte(str), m)
		if err != nil {
			logs.Error("config Setting Json key:%s val:%s err:%s", k, str, err)
			m = def
		}
		return
	}
	return
}

func (s *Setting) Int(k string, def int64) int64 {
	v, ok := s.m.Load(k)
	if ok {
		str := v.(string)
		if str == "" {
			return def
		}
		i, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return def
		}
		return i
	}
	return def
}
