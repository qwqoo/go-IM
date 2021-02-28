package main

import (
	"flag"
	"github.com/qwqoo/go-IM/client"
)

var addr string

func init() {
	flag.StringVar(&addr, "ip", "127.0.0.1:8888", "设置服务器IP和端口 默认是127.0.0.1:8888")
}

func main() {
	flag.Parse()
	client.Run(addr)

}
