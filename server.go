package main

import (
	"fmt"
	"log"
	"net"
	"sync"
)

type Server struct{
	Ip string
	Port int
	OnlineMap map[string]*User // 在线用户列表
	mapLock sync.RWMutex
	Message chan string  // 消息广播的channel
}

func NewServer(ip string, port int) *Server {
	return &Server{
		Ip : ip,
		Port: port,
		OnlineMap: make(map[string]*User),
		Message: make(chan string),
	}
}

func (s *Server) Handler(conn net.Conn){
	// 用户上线 加入onlineMap
	user := NewUser(conn)
	s.mapLock.Lock()
	s.OnlineMap[user.Name]= user
	s.mapLock.Unlock()

	// 广播用户上线消息
	s.BroadCast(user,"已上线")
}

func (s *Server)BroadCast(user *User,msg string)  {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	s.Message <- sendMsg
}

func (s * Server) ListenMessages(){
	for {
		msg := <- s.Message

		s.mapLock.RLock()
		for _,client := range s.OnlineMap{
			client.C <- msg
		}
		s.mapLock.RUnlock()
	}
}
// 	启动服务器的接口
func (s *Server) Start() {
	// socket listen
	addr := fmt.Sprintf("%s:%d",s.Ip,s.Port)
	listener,err := net.Listen("tcp",addr)

	log.Println("server listen:", addr)
	if err != nil{
		log.Panicln("net.Listen error:",err)
	}

	defer listener.Close()

	// 监听消息 如果服务端收到消息 发送给客户端channel
	go s.ListenMessages()
	// accept
	for {
		conn,err := listener.Accept()
		if err != nil {
			log.Println("net.Accept error:",err)
			continue
		}
		// do handler
		go s.Handler(conn)
	}
}