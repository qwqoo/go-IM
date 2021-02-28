package client

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"
)

type Client struct {
	name string
	addr string
	conn net.Conn
	send chan string
	recv chan string
}

func NewClient(addr string) Client {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Panicln("connect error:", err)
	}
	return Client{
		name: addr,
		addr: addr,
		conn: conn,
		send: make(chan string),
		recv: make(chan string),
	}
}

func (c *Client) recvMessage() {
	msg := make([]byte, 4096)
	for {
		_, err := c.conn.Read(msg)
		if err == io.EOF {
			os.Exit(-1)
		} else if err != nil {
			log.Println("recv msg error:", err)
		}
		fmt.Println(string(msg))
	}

}
func (c *Client) sendMessage(msg string) error {
	_, err := c.conn.Write([]byte(msg))
	if err != nil {
		log.Println("send message error:", err)
		return err
	}
	return nil
}
func Run(addr string) {
	c := NewClient(addr)
	var user string
	for i := 0; i < 3; i++ {
		fmt.Print("请输入用户名：")
		_, err := fmt.Scan(&user)
		if err != nil || len(user) < 2 {
			fmt.Println("输入错误 请重新输入！！(用户名长度需大于2)", err)
			continue
		}
		fmt.Println(user)
		err = c.sendMessage("rename-" + user)
		if err != nil {
			log.Panicln("rename error:", err)
		}
		c.name = user
		go c.recvMessage()
		time.Sleep(time.Millisecond * 10)
		break
	}
	var msg string
	for {
		fmt.Print("请输入聊天内容：")
		_, err := fmt.Scan(&msg)
		if err != nil {
			fmt.Println("输入错误，请重新输入！！")
		}
		if msg == "q" || msg == "quit" || msg == "exit" {
			os.Exit(0)
		}
		err = c.sendMessage(msg)
		if err != nil {
			continue
		}
		time.Sleep(time.Millisecond * 500)
	}
}
