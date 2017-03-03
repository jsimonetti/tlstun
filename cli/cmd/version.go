package cmd

import (
	"fmt"
	"runtime"

	"github.com/jsimonetti/tlstun/client"
	"github.com/jsimonetti/tlstun/server"

	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of TLSTun",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Server version: %s\n", server.Version())
		fmt.Printf("Client version: %s\n", client.Version())
		fmt.Printf("Go version: %s\n", runtime.Version())
	},
}
