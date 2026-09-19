package config

import (
	"strings"

	"github.com/spf13/viper"
)

func NewViper() *viper.Viper {
	v := viper.New()
	v.SetEnvPrefix("NETRADOCK")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	v.SetDefault("port", 8080)
	v.SetDefault("session.hours", 12)
	v.SetDefault("secure_cookie", false)
	v.SetDefault("log.level", "info")
	return v
}
