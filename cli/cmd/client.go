package cmd

import (
	"strconv"

	"github.com/jsimonetti/tlstun/client"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	RootCmd.AddCommand(clientCmd)
	clientCmd.AddCommand(registerCmd)
	clientCmd.AddCommand(regstatusCmd)

	clientCmd.PersistentFlags().Bool("verbose", false, "Turn on verbose logging.")
	clientCmd.PersistentFlags().String("bind", "127.0.0.1", "Ip of the client")
	clientCmd.PersistentFlags().Int("port", 1080, "Port of the client")
	clientCmd.PersistentFlags().String("server", "127.0.0.1:8443", "Address of the server")
	clientCmd.PersistentFlags().Bool("insecure", false, "Don't check server certificate")
	clientCmd.PersistentFlags().Bool("nopoison", false, "Don't poison firewall cache")
	clientCmd.PersistentFlags().String("ca", "", "CA certificate filename")
	clientCmd.PersistentFlags().String("certfile", "", "Client certificate filename")
	clientCmd.PersistentFlags().String("keyfile", "", "Client key filename")

	viper.BindPFlag("client_verbose", clientCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("client_bind", clientCmd.PersistentFlags().Lookup("bind"))
	viper.BindPFlag("client_port", clientCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("client_serveraddress", clientCmd.PersistentFlags().Lookup("server"))
	viper.BindPFlag("client_insecure", clientCmd.PersistentFlags().Lookup("insecure"))
	viper.BindPFlag("client_nopoison", clientCmd.PersistentFlags().Lookup("nopoison"))
	viper.BindPFlag("client_ca", clientCmd.PersistentFlags().Lookup("ca"))
	viper.BindPFlag("client_certfile", clientCmd.PersistentFlags().Lookup("certfile"))
	viper.BindPFlag("client_keyfile", clientCmd.PersistentFlags().Lookup("keyfile"))
}

func clientConfig() client.Config {
	return client.Config{
		Port:          strconv.Itoa(viper.GetInt("client_port")),
		Address:       viper.GetString("client_bind"),
		ServerAddress: viper.GetString("client_serveraddress"),
		Verbose:       viper.GetBool("client_verbose"),
		CA:            viper.GetString("client_ca"),
		Certificate:   viper.GetString("client_certfile"),
		Key:           viper.GetString("client_keyfile"),
		Insecure:      viper.GetBool("client_insecure"),
		NoPoison:      viper.GetBool("client_nopoison"),
	}
}

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Start TLSTun client",
	Run:   startClient,
}

func startClient(cmd *cobra.Command, args []string) {
	c := client.NewClient(clientConfig())
	c.Start()
}

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register TLSTun client",
	Run:   registerClient,
}

func registerClient(cmd *cobra.Command, args []string) {
	c := client.NewClient(clientConfig())
	c.Register()
}

var regstatusCmd = &cobra.Command{
	Use:   "regstatus",
	Short: "Check registation status with server",
	Run:   regStatus,
}

func regStatus(cmd *cobra.Command, args []string) {
	c := client.NewClient(clientConfig())
	c.RegisterStatus()
}
