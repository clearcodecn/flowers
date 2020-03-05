package server

type Option struct {
	ClientHttpAddress  string
	ClientProxyAddress string

	ServerProxyAddress string

	Debug bool

	Codec Codec
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

func Debug(d bool) Options {
	return func(o *Option) {
		o.Debug = d
	}
}

func WithCodec() Options {
	return func(option *Option) {
		option.Codec = new(CompressCodec)
	}
}
