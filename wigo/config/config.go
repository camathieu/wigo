package config

import (
	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
	"github.com/root-gg/utils"
	"os"
	"strconv"
	"strings"
)

var config *Config

// LoadConfig application configuration
func LoadDefaultConfig() (err error) {
	config = NewConfig()
	config.Initialize()
	return
}

// LoadConfig application configuration
func LoadConfig(configFile string) (err error) {
	config = NewConfig()
	if err = config.Load(configFile); err != nil {
		return
	}
	config.Initialize()
	return
}

// GetConfig return the application configuration
func GetConfig() *Config {
	return config
}

func Dump() {
	utils.Dump(config)
}

type Config struct {
	// General params
	Global *GeneralConfig

	// Http params
	Http *HttpConfig

	// PushServer params
	PushServer *PushServerConfig

	// PushClient params
	PushClient *PushClientConfig

	// Remmote wigos params
	RemoteWigos  *RemoteWigoConfig
	AdvancedList []AdvancedRemoteWigoConfig

	// Noticications
	Notifications *NotificationConfig

	// OpenTSDB params
	OpenTSDB *OpenTSDBConfig
}

type GeneralConfig struct {
	Hostname              string
	ListenPort            int
	ListenAddress         string
	ProbesDirectory       string
	ProbesConfigDirectory string
	ProbesLibDirectory    string
	UuidFile              string
	LogFile               string
	Debug                 bool
	Group                 string
	Database              string
	AliveTimeout          int
}

type HttpConfig struct {
	Enabled    bool
	Address    string
	Port       int
	SslEnabled bool
	SslCert    string
	SslKey     string
	Login      string
	Password   string
	Gzip       bool
}

type PushServerConfig struct {
	Enabled            bool
	Address            string
	Port               int
	SslEnabled         bool
	SslCert            string
	SslKey             string
	AllowedClientsFile string
	AutoAcceptClients  bool
	MaxWaitingClients  int
}

type PushClientConfig struct {
	Enabled      bool
	Address      string
	Port         int
	SslEnabled   bool
	SslCert      string
	UuidSig      string
	PushInterval int
}

type RemoteWigoConfig struct {
	CheckInterval int
	CheckTries    int

	SslEnabled bool
	Login      string
	Password   string

	List         []string
	AdvancedList []AdvancedRemoteWigoConfig
}

type NotificationConfig struct {
	// Noticications
	MinLevelToSend int

	OnHostChange  bool
	OnProbeChange bool

	HttpEnabled int
	HttpUrl     string

	EmailEnabled     int
	EmailSmtpServer  string
	EmailRecipients  []string
	EmailFromName    string
	EmailFromAddress string
}

type AdvancedRemoteWigoConfig struct {
	Hostname          string
	Port              int
	CheckRemotesDepth int
	CheckInterval     int
	CheckTries        int
	SslEnabled        bool
	Login             string
	Password          string
}

type OpenTSDBConfig struct {
	Enabled       bool
	Address       []string
	SslEnabled    bool
	MetricPrefix  string
	Deduplication int
	BufferSize    int
	Tags          map[string]string
}

func NewConfig() (this *Config) {
	// General params
	this = new(Config)
	this.Global = new(GeneralConfig)
	this.Http = new(HttpConfig)
	this.PushServer = new(PushServerConfig)
	this.PushClient = new(PushClientConfig)
	this.RemoteWigos = new(RemoteWigoConfig)
	this.Notifications = new(NotificationConfig)
	this.OpenTSDB = new(OpenTSDBConfig)

	this.Global.Hostname = ""
	this.Global.Group = "none"
	this.Global.ProbesDirectory = "/usr/local/wigo/probes"
	this.Global.ProbesConfigDirectory = "/etc/wigo/conf.d"
	this.Global.ProbesLibDirectory = "/var/lib/wigo/lib"
	this.Global.LogFile = "/var/log/wigo.log"
	this.Global.UuidFile = "/var/lib/wigo/uuid"
	this.Global.Database = "/var/lib/wigo/wigo.db"
	this.Global.AliveTimeout = 60
	this.Global.Debug = false

	// Http server
	this.Http.Enabled = true
	this.Http.Address = "0.0.0.0"
	this.Http.Port = 4000
	this.Http.SslEnabled = false
	this.Http.SslCert = "/etc/wigo/ssl/wigo.crt"
	this.Http.SslKey = "/etc/wigo/ssl/wigo.key"
	this.Http.Login = ""
	this.Http.Password = ""
	this.Http.Gzip = true

	// Push server
	this.PushServer.Enabled = false
	this.PushServer.Address = "0.0.0.0"
	this.PushServer.Port = 4001
	this.PushServer.SslEnabled = true
	this.PushServer.SslCert = "/etc/wigo/ssl/wigo.crt"
	this.PushServer.SslKey = "/etc/wigo/ssl/wigo.key"
	this.PushServer.AllowedClientsFile = "/var/lib/wigo/allowed"
	this.PushServer.MaxWaitingClients = 100
	this.PushServer.AutoAcceptClients = false

	// Push client
	this.PushClient.Enabled = false
	this.PushClient.Address = "127.0.0.1"
	this.PushClient.Port = 4001
	this.PushClient.SslEnabled = true
	this.PushClient.SslCert = "/etc/wigo/ssl/wigo.crt"
	this.PushClient.UuidSig = "/etc/wigo/ssl/uuid.sig"
	this.PushClient.PushInterval = 15

	// Remote Wigos
	this.RemoteWigos.List = nil
	this.RemoteWigos.CheckInterval = 10
	this.RemoteWigos.CheckTries = 3
	this.AdvancedList = nil

	// Notifications
	this.Notifications.MinLevelToSend = 101

	this.Notifications.OnHostChange = false
	this.Notifications.OnProbeChange = false

	this.Notifications.HttpEnabled = 0
	this.Notifications.HttpUrl = ""

	this.Notifications.EmailEnabled = 0
	this.Notifications.EmailSmtpServer = ""
	this.Notifications.EmailFromAddress = ""
	this.Notifications.EmailFromName = ""
	this.Notifications.EmailRecipients = nil

	// OpenTSDB
	this.OpenTSDB.Enabled = false
	this.OpenTSDB.Address = nil
	this.OpenTSDB.SslEnabled = false
	this.OpenTSDB.MetricPrefix = "wigo"
	this.OpenTSDB.Deduplication = 600
	this.OpenTSDB.BufferSize = 10000
	this.OpenTSDB.Tags = make(map[string]string)

	return
}

// Load config from file
func (c *Config) Load(configFile string) (err error) {
	if _, err = toml.DecodeFile(configFile, &c); err != nil {
		log.Errorf("Failed to load configuration from file %s : %s", configFile, err)
	}
	return
}

func (c *Config) Initialize() {
	// Compatiblity with old RemoteWigos lists
	if c.RemoteWigos.List != nil {
		for _, remoteWigo := range c.RemoteWigos.List {

			// Split data into hostname/port
			splits := strings.Split(remoteWigo, ":")

			hostname := splits[0]
			port := 0
			if len(splits) > 1 {
				port, _ = strconv.Atoi(splits[1])
			}

			if port == 0 {
				port = c.Global.ListenPort
			}

			// Create new RemoteWigoConfig
			AdvancedRemoteWigo := new(AdvancedRemoteWigoConfig)
			AdvancedRemoteWigo.Hostname = hostname
			AdvancedRemoteWigo.Port = port

			AdvancedRemoteWigo.SslEnabled = false
			AdvancedRemoteWigo.Login = ""
			AdvancedRemoteWigo.Password = ""

			// Push new AdvancedRemoteWigo to remoteWigosList
			c.AdvancedList = append(c.AdvancedList, *AdvancedRemoteWigo)
		}
	}

	c.RemoteWigos.AdvancedList = c.AdvancedList
	c.AdvancedList = nil

	os.Setenv("WIGO_PROBE_CONFIG_ROOT", c.Global.ProbesConfigDirectory)
	os.Setenv("WIGO_PROBE_LIB_ROOT", c.Global.ProbesLibDirectory)
}
