package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

type Server struct {
	Ip        string
	Port      int
	OnlineMap map[string]*User // 在线用户列表
	mapLock   sync.RWMutex
	Message   chan string   // 消息广播的channel
	TimeOut   time.Duration // 用户超时时间
}

func NewServer(ip string, port int, timeout time.Duration) *Server {
	return &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
		TimeOut:   timeout,
	}
}

func (s *Server) Handler(conn net.Conn) {
	user := NewUser(conn)
	s.Online(user)
	// 监听用户是否活跃的channel
	isAlive := make(chan bool, 3)
	// 接收客户端消息处理
	go func() {
		buffer := make([]byte, 4096)
		for {
			select {
			case <-user.ctx.Done():
				return
			default:
				n, err := conn.Read(buffer)
				isAlive <- true
				if n == 0 {
					s.Offline(user)
					return
				}
				if err != nil && err != io.EOF {
					log.Println("conn read error:", err)
				}

				msg := string(buffer[:n-1]) // 	去除收到消息到\n
				s.DoMessage(user, msg)
			}
		}
	}()
	for {
		select {
		case <-isAlive:
		// 当用户活跃时将重置下面的定时器
		case <-user.ctx.Done():
			return
		case <-time.After(time.Second * s.TimeOut):
			// 	移除超时的客户端
			_, err := user.conn.Write([]byte("由于你长时间未发言，已被移出群聊\n"))
			if err != nil {
				log.Println("send msg field, error:", err)
			}
			user.cancel()
			close(user.C)
			_ = user.conn.Close()
			delete(s.OnlineMap, user.Name)
			return
		}
	}
}
func (s *Server) Online(user *User) {
	s.mapLock.Lock()
	s.OnlineMap[user.Name] = user
	s.mapLock.Unlock()
	s.BroadCast(user, "已上线")
}

func (s *Server) Offline(user *User) {
	user.cancel()
	s.mapLock.Lock()
	delete(s.OnlineMap, user.Name)
	s.mapLock.Unlock()
	s.BroadCast(user, "已下线")
}

func (s *Server) DoMessage(user *User, msg string) {
	if msg == "who" {
		s.mapLock.RLock()
		for k, v := range s.OnlineMap {
			m := fmt.Sprint("[" + v.Addr + "]" + k + ": 在线")
			user.C <- m
		}
		s.mapLock.RUnlock()
	} else if len(msg) > 7 && msg[:7] == "rename-" {
		newName := strings.Split(msg, "-")[1]
		s.mapLock.RLock()
		_, ok := s.OnlineMap[newName]
		s.mapLock.RUnlock()

		if ok {
			user.C <- "当前用户名已存在"
		} else {
			s.mapLock.Lock()
			delete(s.OnlineMap, user.Name)
			user.Name = newName
			s.OnlineMap[newName] = user
			s.mapLock.Unlock()
			user.C <- "您已经将用户名更新为：" + user.Name
		}
	} else if len(msg) > 4 && msg[:3] == "to-" {
		tmp := strings.Split(msg, "-")
		remoteName := tmp[1]
		content := tmp[2]
		if remoteName == "" || content == "" {
			user.C <- "消息不正确"
			return
		}
		s.mapLock.RLock()
		toUser, ok := s.OnlineMap[remoteName]
		s.mapLock.RUnlock()
		if !ok {
			user.C <- "用户不存在"
		}
		toUser.C <- "[" + user.Name + "]" + " 对你说：" + content
		user.C <- "给" + toUser.Name + "发消息成功"

	} else {
		// 其它消息广播处理
		s.BroadCast(user, msg)
	}
}

// 广播消息
func (s *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	s.Message <- sendMsg
}

func (s *Server) ListenMessages() {
	for {
		msg := <-s.Message

		s.mapLock.RLock()
		for _, client := range s.OnlineMap {
			client.C <- msg
		}
		s.mapLock.RUnlock()
	}
}

// 	启动服务器的接口
func (s *Server) Start() {
	// socket listen
	addr := fmt.Sprintf("%s:%d", s.Ip, s.Port)
	listener, err := net.Listen("tcp", addr)

	log.Println("server listen:", addr)
	if err != nil {
		log.Panicln("net.Listen error:", err)
	}

	defer listener.Close()

	// 监听消息 如果服务端收到消息 发送给客户端channel
	go s.ListenMessages()

	// 接收连接请求 接收到请求后传递 Handler 处理
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("net.Accept error:", err)
			continue
		}
		// do handler
		go s.Handler(conn)
	}
}
