package conf

import (
	"detect/log"
	"github.com/spf13/viper"
	"os"
)

type MasterConfig struct {
	Addr     string
	HttpPort int `mapstructure:"http_port"`
	RpcPort  int `mapstructure:"rpc_port"`
	Mysql    string
	Log      *log.LogConfig
	Storage  Storage
	Endpoint string
}

type Storage struct {
	Falcon  StorageConf
	Db      StorageConf
	Metrics StorageConf
}

type StorageConf struct {
	Enable  bool
	Address string
}

func NewMasterConfig() *MasterConfig {
	return &MasterConfig{}
}

func LoadConfig(path string, name string) (*MasterConfig, error) {
	_, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	masterConfig := &MasterConfig{}
	viper.SetConfigName(name)
	viper.SetConfigType("yaml") // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(path)
	err = viper.ReadInConfig() // Find and read the config file
	if err != nil {            // Handle errors reading the config file
		return nil, err
	}
	err = viper.Unmarshal(masterConfig)
	if err != nil {
		return nil, err
	}
	return masterConfig, nil
}
