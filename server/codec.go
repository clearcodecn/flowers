package server

import (
	"compress/gzip"
)

var (
	defaultLevel = gzip.DefaultCompression
)

type Codec interface {
	Encode([]byte) []byte
	Decode([]byte) []byte
}

type CompressCodec struct {
	Level int
}

func (c *CompressCodec) Encode(p []byte) []byte {
	return p
}

func (c *CompressCodec) Decode(p []byte) []byte {
	return p
}
