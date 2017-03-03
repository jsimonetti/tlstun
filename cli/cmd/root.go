package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// The main command describes the service and defaults to printing the
// help message.
var RootCmd = &cobra.Command{
	Use:   filepath.Base(os.Args[0]),
	Short: "TLSTun",
	Long:  `TLSTun allows tunneling traffic through smart firewalls`,
}
