package main

import (
	"fmt"
	"github.com/myself659/xiciproxy"
	"time"
)

func main() {
	xcproxy := xiciproxy.NewProxyPool()
	url := "https://www.baidu.com"
	resp, err := xcproxy.ProxyGet(url)
	fmt.Println(resp)
	fmt.Println(err)
	<-time.After(60 * time.Second)
}
