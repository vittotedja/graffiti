package util

import (
	"github.com/spf13/viper"
)

type Config struct {
	Env           string `mapstructure:"ENV"`
	DBDriver      string `mapstructure:"DB_DRIVER"`
	DBSource      string `mapstructure:"DB_SOURCE"`
	ServerAddress string `mapstructure:"SERVER_ADDRESS"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	// err = viper.ReadInConfig()
	// if err != nil {
	// 	return
	// }

	if err := viper.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
            return config, err
        }
    }

	err = viper.Unmarshal(&config)
	return
}
