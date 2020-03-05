package client

import (
	"fmt"
	"github.com/clearcodecn/flowers/server"
	"github.com/clearcodecn/flowers/sig"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
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
	argDebug             bool
)

func init() {
	Cmd.Flags().StringVarP(&argClientAddress, "addr", "a", "0.0.0.0:9011", "client listen address,if you want to listen all address please use 0.0.0.0:port")
	Cmd.Flags().StringVarP(&argServerAddress, "saddr", "s", "", "server listen grpc address")
	Cmd.Flags().StringVarP(&argClientHTTPAddress, "haddr", "", "0.0.0.0:8011", "client http address")
	Cmd.Flags().StringVarP(&argServerAddress, "password", "p", "helloworld", "server password")
	Cmd.Flags().BoolVar(&argDebug, "debug", false, "is debug mode")
}

func run(cmd *cobra.Command, args []string) error {
	client, err := server.NewClientProxyServer(
		server.WithClientProxyAddress(argClientAddress),
		server.WithServerProxyAddress(argServerAddress),
		server.WithCodec(),
		server.Debug(argDebug),
	)
	if err != nil {
		return err
	}
	sig.RegisterClose(func() {
		if err := client.Stop(); err != nil {
			logrus.Errorf("client stop: %s", err)
		}
	})
	var msg = ""
	var socks = ""
	if strings.HasPrefix(argClientAddress, ":") {
		msg = fmt.Sprintf("http://127.0.0.1%s", argClientAddress)
		socks = fmt.Sprintf("socks://127.0.0.1%s", argClientAddress)
	} else {
		msg = fmt.Sprintf("http://%s", argClientAddress)
		socks = fmt.Sprintf("socks://%s", argClientAddress)
	}

	fmt.Printf("       ===    代理信息： \n")
	fmt.Printf("		http代理 : %s\n", msg)
	fmt.Printf("		socks代理: %s\n", socks)
	fmt.Printf("\n")
	fmt.Printf("\n")

	return client.Run()
}
