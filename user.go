package main

import (
	"log"
	"net"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn
}

func NewUser(conn net.Conn) *User {
	user := &User{
		Name: conn.RemoteAddr().String(),
		Addr: conn.RemoteAddr().String(),
		C:    make(chan string),
		conn: conn,
	}

	// 新建用户后创建用户goroutine监听channel的消息 发送给client
	go user.ListenMessage()
	return user
}

// 监听当前user的channel，接收到消息 发送给客户端
func (u *User) ListenMessage() {
	for {
		msg := <-u.C
		_, err := u.conn.Write([]byte(msg + "\n"))
		if err != nil {
			log.Println("send msg field, error:", err)
		}
	}
}
