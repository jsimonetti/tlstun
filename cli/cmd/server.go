package cmd

import (
	"strconv"

	"github.com/jsimonetti/tlstun/server"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	RootCmd.AddCommand(serverCmd)

	serverCmd.PersistentFlags().Bool("verbose", false, "Turn on verbose logging.")
	serverCmd.PersistentFlags().String("bind", "", "Ip of the server")
	serverCmd.PersistentFlags().Int("port", 8443, "Port of the server")
	serverCmd.PersistentFlags().String("registerpassword", "", "Register password")
	serverCmd.PersistentFlags().String("ca", "", "CA certificate filename")
	serverCmd.PersistentFlags().String("certfile", "", "Server certificate filename")
	serverCmd.PersistentFlags().String("keyfile", "", "Server key filename")

	viper.BindPFlag("server_verbose", serverCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("server_bind", serverCmd.PersistentFlags().Lookup("bind"))
	viper.BindPFlag("server_port", serverCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("server_registerpassword", serverCmd.PersistentFlags().Lookup("registerpassword"))
	viper.BindPFlag("server_ca", serverCmd.PersistentFlags().Lookup("ca"))
	viper.BindPFlag("server_certfile", serverCmd.PersistentFlags().Lookup("certfile"))
	viper.BindPFlag("server_keyfile", serverCmd.PersistentFlags().Lookup("keyfile"))
}

func serverConfig() server.Config {
	return server.Config{
		Port:         strconv.Itoa(viper.GetInt("server_port")),
		Address:      viper.GetString("server_bind"),
		Verbose:      viper.GetBool("server_verbose"),
		RegisterPass: viper.GetString("server_registerpassword"),
		CA:           viper.GetString("server_ca"),
		Certificate:  viper.GetString("server_certfile"),
		Key:          viper.GetString("server_keyfile"),
	}
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start TLSTun server",
	Run:   startServer,
}

func startServer(cmd *cobra.Command, args []string) {
	s := server.NewServer(serverConfig())
	s.Start()
}
