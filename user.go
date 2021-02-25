package main

import (
	"context"
	"log"
	"net"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	ctx    context.Context
	cancel context.CancelFunc
}

func NewUser(conn net.Conn) *User {
	ctx, cancel := context.WithCancel(context.Background())
	user := &User{
		Name:   conn.RemoteAddr().String(),
		Addr:   conn.RemoteAddr().String(),
		C:      make(chan string),
		conn:   conn,
		ctx:    ctx,
		cancel: cancel,
	}

	// 新建用户后创建用户goroutine监听channel的消息 发送给client
	go user.ListenMessage()
	return user
}

// 监听当前user的channel，接收到消息 发送给客户端
func (u *User) ListenMessage() {
	for {
		select {
		case <-u.ctx.Done():
			return
		case msg := <-u.C:
			_, err := u.conn.Write([]byte(msg + "\n"))
			if err != nil {
				log.Println("send msg field, error:", err)
			}
		}
	}
}
