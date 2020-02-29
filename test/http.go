package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

func main() {
	req, err := http.NewRequest(http.MethodGet, "http://www.baidu.com", nil)
	if err != nil {
		log.Print(err)
	}
	cli := &http.Client{}
	u, _ := url.Parse("http://127.0.0.1:9011")
	cli.Transport = &http.Transport{
		Proxy: http.ProxyURL(u),
	}
	cli.Timeout = time.Second * 99999
	resp, err := cli.Do(req)
	if err != nil {
		log.Println(err)
		return
	}

	io.Copy(os.Stdout, resp.Body)
}
