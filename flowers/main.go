package main

import (
	"fmt"
	"github.com/clearcodecn/flowers/flowers/client"
	"github.com/clearcodecn/flowers/flowers/server"
	"github.com/clearcodecn/flowers/flowers/speed"
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
		speed.Cmd,
	)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
