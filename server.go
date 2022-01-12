package netkit

import (
	"crypto/sha1"
	"fmt"
	"net"
	"time"

	"github.com/xtaci/kcp-go"
	"golang.org/x/crypto/pbkdf2"
)

type ProtocolType int

const (
	ProtocolTCP ProtocolType = iota
	ProtocolKCP
)

type ServerConfig struct {
	Id             int
	Name           string
	IP             string
	Port           int
	Type           ProtocolType
	SessionChanLen int
	WriteTimeout   time.Duration
	ReadTimeout    time.Duration

	// KCP Config
	KCPPassword string
	KCPSalt     string
}

type Server[T any, P PacketType[T]] struct {
	config   ServerConfig
	listener net.Listener

	// ChanConnect chan *Session //收到连接信息时处理
	ChanPacket  chan P //服务器协程的包
	OnConnected func(sess *Session[T, P])
	handlerMap  map[int]*HandlerInfo[T, P] //注册在Server上的处理函数
	middlewares []Middleware[T, P]         //对包的预处理和判断
}

func NewServer[T any, P PacketType[T]](config ServerConfig) *Server[T, P] {
	return &Server[T, P]{
		config:   config,
		listener: nil,
		// ChanConnect: make(chan *Session),
		handlerMap:  map[int]*HandlerInfo[T, P]{},
		OnConnected: func(sess *Session[T, P]) {},
	}
}

func (s *Server[T, P]) RegisterHandler(messageId int, handler PacketHandler[T, P]) {
	s.handlerMap[messageId] = &HandlerInfo[T, P]{
		MessageId: messageId,
		Handler:   handler,
	}
}

func (s *Server[T, P]) AddMiddleware(middleware Middleware[T, P]) {
	s.middlewares = append(s.middlewares, middleware)
}

func (s *Server[T, P]) GetHandlerById(messageId int) *HandlerInfo[T, P] {
	return s.handlerMap[messageId]
}

func (s *Server[T, P]) GetMiddlewares() []Middleware[T,P] {
	return s.middlewares
}

func (s *Server[T, P]) Init() error {
	switch s.config.Type {
	case ProtocolTCP:
		if listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.config.IP, s.config.Port)); err == nil {
			s.listener = listener
			return nil
		} else {
			return err
		}
	case ProtocolKCP:
		key := pbkdf2.Key([]byte(s.config.KCPPassword), []byte(s.config.KCPSalt), 1024, 32, sha1.New)
		block, err := kcp.NewAESBlockCrypt(key)
		if err != nil {
			return err
		}
		if listener, err := kcp.ListenWithOptions(fmt.Sprintf("%s:%d", s.config.IP, s.config.Port), block, 10, 3); err == nil {
			s.listener = listener
			return nil
		} else {
			return err
		}
	default:
		return fmt.Errorf("unknown protocol type, type: %d", s.config.Type)
	}
}

func (s *Server[T, P]) Start() error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return err
		}
		session := NewSession(conn, SessionInfo[T,P]{
			ReadTimeout:  s.config.ReadTimeout,
			WriteTimeout: s.config.WriteTimeout,
			QueueLen:     s.config.SessionChanLen,
			Handler:      s,
		})
		s.OnConnected(session)
		// select {
		// case s.ChanConnect <- session:
		// case <-time.After(time.Second * 2):
		// 	log.Println("connect timeout")
		// }
	}
}
