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
	req, err := http.NewRequest(http.MethodGet, "http://dj1.baidu.com/v.gif?logactid=1234567890&showTab=10000&opType=showpv&mod=superman%3Alib&submod=index&superver=supernewplus&glogid=2360765836&type=2011&pid=315&isLogin=0&version=PCHome&terminal=PC&qid=2360766021&sid=1421_21087_30823_30717&super_frm=&from_login=&from_reg=&query=&curcard=2&curcardtab=&_r=0.2964419084551888", nil)
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
