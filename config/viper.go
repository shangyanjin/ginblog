package config

import (
"fmt"
"os"

"github.com/spf13/viper"
)

type Conf struct {
	Http  Http
	Mysql Mysql
	Redis Redis
	Wxapp Wxapp
	Oss   Oss
}

type Http struct {
	Addr string
	Port string
}

type Mysql struct {
	Addr     string
	Username string
	Password string
	Database string
}

type Redis struct {
	Addr   string
	Passwd string
	Db     int
}

type Wxapp struct {
	Appid     string	//微信小程序 ID
	Appsecret string	//微信小程序 密钥
	MchID     string 	//微信支付商户 ID
	ApiKey 	  string  	//微信支付商户 KEY
	ApiSecret string  	//微信支付商户 密钥
	NotifyURL string  	//微信支付商户 URL
}

type Oss struct {
	AccessKeyId     string
	AccessKeySecret string
	BucketName      string
	EndPoint        string
}

func LoadConf() *Conf {
	env := getMode()
	viper.SetConfigName("config")
	viper.AddConfigPath("conf/")
	viper.AddConfigPath("./")

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	C := &Conf{}
	err = viper.Unmarshal(C)
	if err != nil {
		panic(env)
	}
	return C
}

func getMode() string {
	env := os.Getenv("RUN_MODE")
	if env == "" {
		env = "dev"
	}
	return env
}
