package server

import (
	"flowers/server"
	"flowers/sig"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	Cmd = &cobra.Command{
		Use:        "server",
		Aliases:    []string{"s"},
		SuggestFor: nil,
		Short:      "flowers client",
		Long:       "start proxy at client side",
		RunE:       run,
	}

	argServerAddress string
)

func init() {
	Cmd.Flags().StringVarP(&argServerAddress, "saddr", "s", ":9012", "server listen grpc address")
}

func run(cmd *cobra.Command, args []string) error {
	s := server.NewProxyServer(
		server.WithServerProxyAddress(argServerAddress),
		server.WithCipher("123456"),
	)
	sig.RegisterClose(func() {
		if err := s.Stop(); err != nil {
			logrus.Errorf("server stop: %s", err)
		}
	})
	return s.Run()
}
