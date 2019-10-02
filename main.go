package main

import (
	"Server/Protocol"
	"Server/Utility"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"strconv"
	"sync"
)

type Client struct {
	Conn           net.Conn
	Id             uint32
	X              int32
	Y              int32
	PacketSize     int32
	PrevPacketSize int32
	PacketBuf      []byte
}

func (this *Client) Init() {
	this.PacketSize = 0
	this.PrevPacketSize = 0
	this.PacketBuf = make([]byte, 512)
}

var ClientList map[uint32]*Client
var C_List_lock *sync.Mutex
var RunningFlag bool = true

//var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

func main() {
	//defer profile.Start(profile.CPUProfile).Stop()

	//flag.Parse()

	fmt.Println("Launching server...")

	ClientList = make(map[uint32]*Client)
	C_List_lock = new(sync.Mutex)

	port := strconv.Itoa(Protocol.SERVER_PORT)

	ln, err := net.Listen("tcp", ":"+port)
	if nil != err {
		log.Fatalf("fail to bind Port err: %v", err)
	}

	var idcounter uint32 = 0
	for RunningFlag {
		conn, err := ln.Accept()

		if nil != err {
			log.Fatalln("Connection error")
			continue
		}

		c := Client{
			Conn: conn,
			Id:   idcounter,
		}
		c.X = int32(rand.Intn(50))
		c.Y = int32(rand.Intn(50))
		c.Init()

		packet := Protocol.Packet_SC_Login_OK{
			Packet_type: Protocol.SC_LOGIN_OK,
			Id:          c.Id,
		}

		Utility.SendPacket(c.Conn, packet)
		idcounter++

		// to all
		putpacket := Protocol.Packet_SC_Put_Player{
			Packet_type: Protocol.SC_PUT_PLAYER,
			X:           uint16(c.X),
			Y:           uint16(c.Y),
			Id:          c.Id,
		}

		C_List_lock.Lock()
		ClientList[c.Id] = &c
		C_List_lock.Unlock()

		var idx uint32 = 0
		for idx < idcounter {
			cl := ClientList[idx]
			if cl != nil {
				Utility.SendPacket(cl.Conn, putpacket)
			}
			idx++
		}
		idx = 0
		// to me
		for idx < idcounter {
			cl := ClientList[idx]

			if cl != nil && cl.Id != c.Id {
				packet := Protocol.Packet_SC_Put_Player{
					Packet_type: Protocol.SC_PUT_PLAYER,
					X:           uint16(cl.X),
					Y:           uint16(cl.Y),
					Id:          cl.Id,
				}

				Utility.SendPacket(c.Conn, packet)
			}
			idx++
		}

		go ConnectionProcess(&c)

	}

	println("Server End")
}

func ConnectionProcess(Client *Client) {
	recvBuf := make([]byte, 4096)

	defer func(buf []byte) {
		recover()

		C_List_lock.Lock()
		delete(ClientList, Client.Id)
		C_List_lock.Unlock()

		Client.PacketBuf = nil
		Client = nil
		recvBuf = nil
	}(recvBuf)

	for {
		n, err := Client.Conn.Read(recvBuf)
		if nil != err {
			if io.EOF == err {
				log.Printf("connection is closed from client :: %v", Client.Conn.RemoteAddr().String())
				return
			}

			//log.Println(err)
			return
		}

		if n == 0 {
			fmt.Printf("Packet size - 0 /// Disconnected???")
		} else {
			var recv_size int32 = int32(n)
			var saved_size int32 = 0

			// Packet assemble
			for 0 < recv_size {
				if Client.PacketSize == 0 {
					Client.PacketSize = int32(recvBuf[0])
				}

				remainsize := Client.PacketSize - Client.PrevPacketSize

				if remainsize <= recv_size {
					// Packet Process
					copy(Client.PacketBuf[Client.PrevPacketSize:], recvBuf[saved_size:saved_size+remainsize])

					ProcessPacket(Client, Client.PacketBuf)

					recv_size -= remainsize
					saved_size += remainsize
					Client.PacketSize = 0
					Client.PrevPacketSize = 0

				} else {
					// copy rest Packet part
					copy(Client.PacketBuf[Client.PrevPacketSize:], recvBuf[:recv_size])
					Client.PrevPacketSize += recv_size
				}
			}
		}
	}
}

func ProcessPacket(Client *Client, Buf []byte) {
	PacketSize := Buf[0]
	PacketType := Buf[1]

	// Temporary Process
	recvPacket := Protocol.CS_Packet_Move{}
	err := binary.Read(bytes.NewBuffer(Buf[:PacketSize]), binary.LittleEndian, &recvPacket)
	if err != nil {
		log.Fatalln("Recv data Parsing Error")
	}

	switch PacketType {
	case Protocol.CS_DOWN:
		Client.Y++
	case Protocol.CS_UP:
		Client.Y--
	case Protocol.CS_RIGHT:
		Client.X++
	case Protocol.CS_LEFT:
		Client.X--
	}

	packet := Protocol.Packet_SC_POS{
		Packet_type: Protocol.SC_POS,
		Id:          Client.Id,
		X:           uint8(Client.X),
		Y:           uint8(Client.Y),
	}

	C_List_lock.Lock()
	for _, c := range ClientList {
		Utility.SendPacket(c.Conn, packet)
	}
	C_List_lock.Unlock()
}
