package conf

import (
	"detect/log"
	"github.com/spf13/viper"
	"os"
	"time"
)

type AgentConfig struct {
	MasterAddr string `mapstructure:"master_addr"`
	Interval   time.Duration
	Hostname   string
	Log        *log.LogConfig
}

func LoadConfig(path string, name string) (*AgentConfig, error) {
	_, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	config := &AgentConfig{}
	viper.SetConfigName(name)
	viper.SetConfigType("yaml") // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(path)
	err = viper.ReadInConfig() // Find and read the config file
	if err != nil {            // Handle errors reading the config file
		return nil, err
	}
	err = viper.Unmarshal(config)
	if err != nil {
		return nil, err
	}
	if config.Hostname == "" {
		config.Hostname = GetHostName()
	}
	return config, nil
}

func GetHostName() string {
	if name, err := os.Hostname(); err != nil {
		return name
	}
	return ""
}
