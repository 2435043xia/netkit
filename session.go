package netkit

import (
	"fmt"
	"log"
	"net"
	"time"
)

type Session[T any, P PacketType[T]] struct {
	SessionInfo[T,P]
	conn   net.Conn
	readTimeout  time.Duration
	writeTimeout time.Duration

	// chanRead  chan *Packet //默认将包传到服务器携程
	chanWrite chan P
}

type SessionInfo[T any, P PacketType[T]] struct {
	ReadTimeout, WriteTimeout time.Duration
	QueueLen int
	Handler Handler[T,P]
}

func NewSession[T any, P PacketType[T]](conn net.Conn, info SessionInfo[T,P]) *Session[T, P] {
	return &Session[T, P]{
		conn:   conn,
		// chanRead:  server.ChanPacket,
		SessionInfo: info,
		chanWrite: make(chan P, info.QueueLen),
	}
}

func (s *Session[T, P]) HandleReadLoop() {
	for {
		if s.readTimeout != 0 {
			s.conn.SetReadDeadline(time.Now().Add(s.readTimeout))
		}
		packet := NewPacket[T, P](0)
		err := packet.ReadFromConn(s.conn)
		if err != nil {
			log.Println("read data err:", err)
			return
		}

		for _, m := range s.Handler.GetMiddlewares() {
			err = m(s, packet)
			if err != nil {
				break
			}
		}
		if err != nil {
			log.Println(err)
			continue
		}
		// s.chanRead <- packet
		handler := s.Handler.GetHandlerById(packet.GetMsgId())
		if handler == nil {
			log.Println("handler not found", packet.GetMsgId())
			continue
		}
		//需要操作其他协程的直接加锁
		handler.Handler(s, packet)
	}
}

// 读取到的包传到哪里
// func (s *Session[T, P]) SetReadChan(readChan chan *Packet) {
// 	s.chanRead = readChan
// }

func (s *Session[T, P]) HandleWriteLoop() {
	for {
		if s.WriteTimeout != 0 {
			s.conn.SetWriteDeadline(time.Now().Add(s.WriteTimeout))
		}
		packet := <-s.chanWrite
		//if packet == nil {
		//	continue
		//}
		_, err := s.conn.Write(packet.GetSendData())
		if err != nil {
			log.Panicln(err)
			return
		}
	}
}

func (s *Session[T, P]) SendPacket(packet P) error {
	var timeout <-chan time.Time
	if s.WriteTimeout != 0 {
		timeout = time.After(s.WriteTimeout)
	}
	select {
	case s.chanWrite <- packet:
	case <-timeout:
		return fmt.Errorf("send packet timeout")
	}
	return nil
}
