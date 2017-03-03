package main

import (
	"github.com/jsimonetti/tlstun/cli/cmd"
	"github.com/spf13/cobra/doc"
)

func main() {
	doc.GenMarkdownTree(cmd.RootCmd, "./")
}
