package util

import (
	"github.com/spf13/viper"
)

type Config struct {
	Env              string `mapstructure:"ENV"`
	DBDriver         string `mapstructure:"DB_DRIVER"`
	DBSource         string `mapstructure:"DB_SOURCE"`
	ServerAddress    string `mapstructure:"SERVER_ADDRESS"`
	AWSRegion        string `mapstructure:"AWS_REGION"`
	AWSAccessKeyID   string `mapstructure:"AWS_ACCESS_KEY_ID"`
	AWSSecretKey     string `mapstructure:"AWS_SECRET_ACCESS_KEY"`
	AWSSessionToken  string `mapstructure:"AWS_SESSION_TOKEN"`
	AWSS3Bucket      string `mapstructure:"AWS_S3_BUCKET"`
	CloudfrontDomain string `mapstructure:"CLOUDFRONT_DOMAIN"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	// if err := viper.ReadInConfig(); err != nil {
	//     if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
	//         return config, err
	//     }
	// }

	err = viper.Unmarshal(&config)
	return
}
