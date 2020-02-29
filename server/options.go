package server

type Option struct {
	ClientHttpAddress  string
	ClientProxyAddress string

	ServerProxyAddress string

	Cipher Cipher
}

type Options func(*Option)

func WithClientHttpAddress(address string) Options {
	return func(options *Option) {
		options.ClientHttpAddress = address
	}
}

func WithClientProxyAddress(address string) Options {
	return func(options *Option) {
		options.ClientProxyAddress = address
	}
}

func WithServerProxyAddress(address string) Options {
	return func(options *Option) {
		options.ServerProxyAddress = address
	}
}

func WithCipher(password string) Options {
	return func(option *Option) {
		var err error
		option.Cipher, err = NewDictCipher([]byte(password))
		if err != nil {
			panic(err)
		}
	}
}
