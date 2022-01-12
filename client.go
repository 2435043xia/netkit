package netkit

import (
	"crypto/sha1"
	"fmt"
	"net"
	"time"

	"github.com/xtaci/kcp-go"
	"golang.org/x/crypto/pbkdf2"
	"google.golang.org/protobuf/proto"
)

type ClientOptions struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	ProtocolType ProtocolType
	ReadQueueLen int

	// KCP Config
	KCPPassword string
	KCPSalt     string
}

type Client[T any, P PacketType[T]] struct {
	Session *Session[T,P]
	config ClientOptions

	middlewares []Middleware[T, P]
	handlerMap map[int]*HandlerInfo[T, P] // register handle func

}

func (c *Client[T, P]) GetHandlerById(messageId int) *HandlerInfo[T, P] {
	return c.handlerMap[messageId]
}

func (c *Client[T, P]) GetMiddlewares() []Middleware[T, P] {
	return c.middlewares
}

func NewClient[T any, P PacketType[T]](options ClientOptions) *Client[T, P] {
	return &Client[T, P]{
		config: options,
		handlerMap: map[int]*HandlerInfo[T, P]{},
		middlewares: make([]Middleware[T, P],0),
	}
}

func (c *Client[T, P]) Connect(ip string, port int) error {
	var err error
	var conn net.Conn
	switch c.config.ProtocolType {
	case ProtocolTCP:
		conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
	case ProtocolKCP:
		var block kcp.BlockCrypt
		key := pbkdf2.Key([]byte(c.config.KCPPassword), []byte(c.config.KCPSalt), 1024, 32, sha1.New)
		block, err = kcp.NewAESBlockCrypt(key)
		if err != nil {
			return err
		}
		conn, err = kcp.DialWithOptions(fmt.Sprintf("%s:%d", ip, port), block, 10, 3)
	default:
		return fmt.Errorf("unknown protocolType")
	}
	if err != nil {
		return err
	}
	c.Session = NewSession(conn, SessionInfo[T,P]{
		ReadTimeout:  c.config.ReadTimeout,
		WriteTimeout: c.config.WriteTimeout,
		QueueLen:     c.config.ReadQueueLen,
		Handler:      c,
	})
	return nil
}

func (c *Client[T, P]) RegisterHandler(messageId int, handler PacketHandler[T, P]) {
	c.handlerMap[messageId] = &HandlerInfo[T, P]{
		MessageId: messageId,
		Handler:   handler,
	}
}

func (c *Client[T, P]) SendPacket(packet P) error {
	return c.Session.SendPacket(packet)
}

func (c *Client[T, P]) AddMiddleware(middleware Middleware[T, P]) {
	c.middlewares = append(c.middlewares, middleware)
}

func (c *Client[T, P]) SendPB(messageId int, pb proto.Message) error {
	packet := NewPacket[T, P](1)
	packet.AppendPB(pb)
	return c.SendPacket(packet)
}
