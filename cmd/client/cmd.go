package client

import (
	"github.com/clearcodecn/flowers/server"
	"github.com/clearcodecn/flowers/sig"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	Cmd = &cobra.Command{
		Use:        "client",
		Aliases:    []string{"c"},
		SuggestFor: nil,
		Short:      "flowers client",
		Long:       "start proxy at client side",
		RunE:       run,
	}

	argClientAddress string
	argServerAddress string

	argClientHTTPAddress string
	argPassword          string
)

func init() {
	Cmd.Flags().StringVarP(&argClientAddress, "addr", "a", "127.0.0.1:9011", "client listen address,if you want to listen all address please use 0.0.0.0:port")
	Cmd.Flags().StringVarP(&argServerAddress, "saddr", "s", "", "server listen grpc address")
	Cmd.Flags().StringVarP(&argClientHTTPAddress, "haddr", "", "127.0.0.1:8011", "client http address")
	Cmd.Flags().StringVarP(&argServerAddress, "password", "p", "helloworld", "server password")
}

func run(cmd *cobra.Command, args []string) error {
	client, err := server.NewClientProxyServer(
		server.WithClientProxyAddress(argClientAddress),
		server.WithServerProxyAddress(argServerAddress),
		server.WithCipher(argPassword),
	)
	if err != nil {
		return err
	}
	sig.RegisterClose(func() {
		if err := client.Stop(); err != nil {
			logrus.Errorf("client stop: %s", err)
		}
	})
	return client.Run()
}
