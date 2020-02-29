# flowers 

> simple and powerful proxy over http2/grpc 
> support http/https/socks5


# Installation 
```shell script
go get github.com/clearcodecn/flowers/...
```

# Run Server

```shell script
flowers server -p password
```

# Run Client 
```shell script
flowers client -a 0.0.0.0:9011 -s 127.0.0.1:9012 -p password
```