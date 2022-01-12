package netkit

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"

	"google.golang.org/protobuf/proto"
)

const (
	HeadSize    = 8
	OffsetLen   = 0
	OffsetMsgId = OffsetLen + 4
	OffsetData  = OffsetMsgId + 4
)

type PacketType[T any] interface {
	*T
	ReadFromConn(conn net.Conn) error
	ParsePBData(pb proto.Message) error
	ParseHead() error
	AppendPB(pb proto.Message) error
	GetSendData() []byte
	GetMsgId() int
	SetMsgId(msgId int)
}

type JSONPacket struct {
	len   int
	msgId int
	data  []byte
}

func (J *JSONPacket) GetMsgId() int {
	//TODO implement me
	panic("implement me")
	return 0
}

func (J *JSONPacket) SetMsgId(int) {
	//TODO implement me
	panic("implement me")
}

func (J *JSONPacket) ReadFromConn(conn net.Conn) error {
	//TODO implement me
	panic("implement me")
}

func (J *JSONPacket) ParsePBData(pb proto.Message) error {
	//TODO implement me
	panic("implement me")
}

func (J *JSONPacket) ParseHead() error {
	//TODO implement me
	panic("implement me")
}

func (J *JSONPacket) AppendPB(pb proto.Message) error {
	//TODO implement me
	panic("implement me")
}

func (J *JSONPacket) GetSendData() []byte {
	//TODO implement me
	panic("implement me")
}

type Packet struct {
	len   int
	msgId int
	data  []byte
}

func (p *Packet) GetMsgId() int {
	return p.msgId
}

func (p *Packet) SetMsgId(id int) {
	p.msgId = id
}

func NewPacket[T any, P PacketType[T]](messageId int) P {
	var str T
	p := P(&str)
	p.SetMsgId(messageId)
	return p
}

func (p *Packet) ReadFromConn(conn net.Conn) error {
	p.data = make([]byte, HeadSize)
	_, err := io.ReadFull(conn, p.data)
	if err != nil {
		return err
	}
	err = p.ParseHead()
	if err != nil {
		return err
	}
	p.data = make([]byte, p.len)
	_, err = io.ReadFull(conn, p.data)
	if err != nil {
		return err
	}

	return nil
}

func (p *Packet) ParsePBData(pb proto.Message) error {
	return proto.Unmarshal(p.data, pb)
}

func (p *Packet) ParseHead() error {
	if len(p.data) < HeadSize {
		return fmt.Errorf("data size too short")
	}
	p.len = int(binary.LittleEndian.Uint32(p.data))
	p.msgId = int(binary.LittleEndian.Uint32(p.data[OffsetMsgId:]))
	return nil
}

func (p *Packet) AppendPB(pb proto.Message) error {
	b, err := proto.Marshal(pb)
	if err != nil {
		return err
	}
	if len(p.data) == 0 {
		p.data = b
	} else {
		p.data = append(p.data, b...)
	}
	p.len += len(b)
	return nil
}

func (p *Packet) GetSendData() []byte {
	b := make([]byte, HeadSize+p.len)
	binary.LittleEndian.PutUint32(b[OffsetLen:], uint32(p.len))
	binary.LittleEndian.PutUint32(b[OffsetMsgId:], uint32(p.msgId))
	copy(b[OffsetData:], p.data)
	return b
}
