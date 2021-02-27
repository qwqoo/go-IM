package main

import "github.com/qwqoo/go-IM/server"

// Todo 解析配置文件
// Todo 初始化log
func init() {
	// fmt.Println("。。。。。")
}

func main() {
	s := server.NewServer("127.0.0.1", 8888, 30)
	s.Start()
}
