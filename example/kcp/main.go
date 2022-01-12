package main

import (
	"fmt"
	"github.com/2435043xia/netkit"
	"log"
	"time"
)

const pwd, salt = "netkit", "netkit"

func main() {
	server := netkit.NewServer[netkit.Packet](netkit.ServerConfig{
		IP:             "0.0.0.0",
		Port:           8955,
		Type:           netkit.ProtocolKCP,
		SessionChanLen: 10,
		KCPPassword:    pwd,
		KCPSalt:        salt,
	})
	err := server.Init()
	if err != nil {
		log.Panicln(err)
	}
	server.AddMiddleware(func(sess *netkit.Session[netkit.Packet, *netkit.Packet], packet *netkit.Packet) error{
		fmt.Println("server middleware 1")
		return nil
	})
	server.OnConnected = func(sess *netkit.Session[netkit.Packet, *netkit.Packet]) {
		go sess.HandleReadLoop()
		go sess.HandleWriteLoop()
	}
	// register event onmessage
	server.RegisterHandler(1, func(session *netkit.Session[netkit.Packet, *netkit.Packet], packet *netkit.Packet) {
		fmt.Println("server recv:", packet)
		session.SendPacket(netkit.NewPacket[netkit.Packet](2))
	})
	go server.Start()

	client := netkit.NewClient[netkit.Packet](netkit.ClientOptions{
		ProtocolType: netkit.ProtocolKCP,
		KCPPassword:  pwd,
		KCPSalt:      salt,
	})
	client.RegisterHandler(2, func(session *netkit.Session[netkit.Packet, *netkit.Packet], packet *netkit.Packet) {
		fmt.Println("client recv:", packet)
	})
	client.AddMiddleware(func(sess *netkit.Session[netkit.Packet, *netkit.Packet], packet *netkit.Packet) error{
		fmt.Println("client middleware 1")
		return nil
	})
	err = client.Connect("127.0.0.1", 8955)
	if err != nil {
		log.Panicln(err)
	}
	go client.Session.HandleWriteLoop()
	go client.Session.HandleReadLoop()
	for {
		client.SendPB(1, nil)
		time.Sleep(time.Second)
	}
}
