package conf

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"rocketmqtt/logger"
	"strings"

	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

var RunConfig *Config

type Config struct {
	Broker        Broker         `yaml:"broker"`
	Listen        Listen         `yaml:"listen"`
	TlsInfo       tlsInfo        `yaml:"tlsInfo"`
	DeliversRules []DeliversRule `yaml:"deliversRules"`
	Plugins       struct {
		Rocketmq []Rocketmq `yaml:"rocketmq"`
		Kafka    []Kafka    `yaml:"kafka"`
	} `yaml:"plugins"`
	Auth map[string]string `yaml:"auth"`
}

type Broker struct {
	ID           string `default:"{{hostname}}" yaml:"id"`
	TcpKeepalive int    `default:"125" yaml:"tcpKeepalive"`
	WorkerNum    int    `default:"1024" yaml:"workerNum"`
	LogLevel     string `default:"debug" yaml:"logLevel"`
}

type Listen struct {
	Host          string `default:"0.0.0.0" yaml:"host"`
	Port          string `default:"1883" yaml:"port"`
	TLSPort       string `default:"8883" yaml:"tlsPort"`
	ManagePort    string `default:"7070" yaml:"managePort"`
	PprofPort     string `default:"6060" yaml:"pprofPort"`
	MetricsPort   string `default:"5050" yaml:"metricsPort"`
	WebsocketPort string `default:"80" yaml:"websocketPort"`
	WebsocketPath string `default:"/ws" yaml:"websocketPath"`
	WebsocketTls  bool   `default:"false" yaml:"websocketTls"`
}

type tlsInfo struct {
	Enabled  bool   `default:"false" yaml:"enabled"`
	Verify   bool   `yaml:"verify"`
	CaFile   string `yaml:"caFile"`
	CertFile string `yaml:"certFile"`
	KeyFile  string `yaml:"keyFile"`
}

type DeliversRule struct {
	Pattern   string    `yaml:"pattern"`
	Plugin    string    `yaml:"plugin"`
	Target    string    `yaml:"target"`
	Topic     string    `yaml:"topic"`
	Tag       string    `yaml:"tag"`
	NameSplit *[]string `yaml:"-"`
}

type Rocketmq struct {
	Name            string `yaml:"name"`
	Enable          bool   `yaml:"enable"`
	EnableSubscribe bool   `yaml:"enableSubscribe"`
	SubscribeTopic  string `yaml:"subscribeTopic"`
	SubscribeModel  string `yaml:"subscribeModel"`
	SubscribeTag    string `yaml:"subscribeTag"`
	NameSrv         string `yaml:"nameSrv"`
	GroupName       string `yaml:"groupName"`
}

type Kafka struct {
	Name      string   `yaml:"name"`
	Enable    bool     `yaml:"enable"`
	Addr      []string `yaml:"addr"`
	GroupName string   `yaml:"groupName"`
}

type Auth struct {
	Username string
	Password string
}

func NewTLSConfig() (*tls.Config, error) {

	tlsInfo := RunConfig.TlsInfo

	cert, err := tls.LoadX509KeyPair(tlsInfo.CertFile, tlsInfo.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("error parsing X509 certificate/key pair: %v", zap.Error(err))
	}
	cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return nil, fmt.Errorf("error parsing certificate: %v", zap.Error(err))
	}

	// Create TLSConfig
	// We will determine the cipher suites that we prefer.
	config := tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	// Require client certificates as needed
	if tlsInfo.Verify {
		config.ClientAuth = tls.RequireAndVerifyClientCert
	}
	// Add in CAs if applicable.
	if tlsInfo.CaFile != "" {
		rootPEM, err := ioutil.ReadFile(tlsInfo.CaFile)
		if err != nil || rootPEM == nil {
			return nil, err
		}
		pool := x509.NewCertPool()
		ok := pool.AppendCertsFromPEM([]byte(rootPEM))
		if !ok {
			return nil, fmt.Errorf("failed to parse root ca certificate")
		}
		config.ClientCAs = pool
	}

	return &config, nil
}

func init() {
	content, err := ioutil.ReadFile("conf/liumqtt.yaml")
	if err != nil {
		panic(err)
	}
	var c Config
	err = yaml.Unmarshal(content, &c)
	if err != nil {
		panic(err)
	}
	// 设置日志邓丽
	switch c.Broker.LogLevel {
	case "debug":
		logger.Instance, _ = logger.NewDevLogger()

	case "info":
		logger.Instance, _ = logger.NewProdLogger(zap.InfoLevel)

	case "warn":
		logger.Instance, _ = logger.NewProdLogger(zap.WarnLevel)

	case "error":
		logger.Instance, _ = logger.NewProdLogger(zap.ErrorLevel)

	default:
		logger.Instance, _ = logger.NewProdLogger(zap.InfoLevel)

	}

	for i, deliver := range c.DeliversRules {
		s := strings.Split(deliver.Pattern, "/")
		deliver.NameSplit = &s
		c.DeliversRules[i] = deliver
	}
	// 如果是hostname则获取系统hostname
	if c.Broker.ID == "{{HOSTNAME}}" {
		hostname, err := os.Hostname()
		if err != nil {
			c.Broker.ID = "unkown-hostname"
		} else {
			c.Broker.ID = hostname
		}
	}

	RunConfig = &c
}
