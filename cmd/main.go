package main

import (
	"github.com/clearcodecn/flowers/cmd/client"
	"github.com/clearcodecn/flowers/cmd/server"
	"fmt"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:     "flowers",
		Aliases: []string{"f"},
	}
)

func init() {
	rootCmd.AddCommand(
		client.Cmd,
		server.Cmd,
	)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
