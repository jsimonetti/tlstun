package main

import (
	"os"

	"github.com/jsimonetti/tlstun/cli/cmd"

	"github.com/spf13/viper"
)

func init() {

	viper.SetEnvPrefix("TLSTUN")
	viper.AutomaticEnv()

	viper.SetConfigName("tlstun") // name of config file (without extension)
	viper.AddConfigPath(".")      // more path to look for the config files

	viper.ReadInConfig()
}

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		println(err)
		os.Exit(-1)
	}
}
