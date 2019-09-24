package main

import (
	"Server/Protocol"
	"Server/Utility"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
)

type Client struct {
	Conn net.Conn
	Id   uint32
	X    uint32
	Y    uint32
}

func main() {
	fmt.Println("Launching server...")

	port := strconv.Itoa(Protocol.SERVER_PORT)

	ln, err := net.Listen("tcp", ":"+port)
	if nil != err {
		log.Fatalf("fail to bind Port err: %v", err)
	}

	var idcounter uint32 = 0

	for {
		conn, err := ln.Accept()

		if nil != err {
			log.Fatalln("Connection error")
			continue
		}

		c := Client{
			Conn: conn,
			Id:   idcounter,
			X:    4,
			Y:    4,
		}

		packet := Protocol.Packet_SC_Login_OK{
			Packet_type: Protocol.SC_LOGIN_OK,
			Id:          c.Id,
		}

		Utility.SendPacket(c.Conn, packet)

		go ConnectionProcess(&c)
		idcounter++
	}
}

func ConnectionProcess(Client *Client) {
	recvBuf := make([]byte, 4096)

	packet := Protocol.Packet_SC_Put_Player{
		Packet_type: Protocol.SC_PUT_PLAYER,
		X:           uint16(Client.X),
		Y:           uint16(Client.Y),
		Id:          Client.Id,
	}
	Utility.SendPacket(Client.Conn, packet)

	for {
		n, err := Client.Conn.Read(recvBuf)
		if nil != err {
			if io.EOF == err {
				log.Printf("connection is closed from client; %v", Client.Conn.RemoteAddr().String())
				return
			}
			log.Printf("fail to receive data; err: %v", err)
			return
		}

		// Packet Processing
		if n > 0 {
			packettype := recvBuf[1]
			recvPacket := Protocol.CS_Packet_Move{}
			err = binary.Read(bytes.NewBuffer(recvBuf[:n]), binary.LittleEndian, &recvPacket)
			if err != nil {
				log.Fatalln("Recv data Parsing Error")
			}

			switch packettype {
			case Protocol.CS_DOWN:
				Client.Y++
				break
			case Protocol.CS_UP:
				Client.Y--
				break
			case Protocol.CS_RIGHT:
				Client.X++
				break
			case Protocol.CS_LEFT:
				Client.X--
				break
			}

			packet := Protocol.Packet_SC_POS{
				Packet_type: Protocol.SC_POS,
				Id:          Client.Id,
				X:           uint8(Client.X),
				Y:           uint8(Client.Y),
			}
			Utility.SendPacket(Client.Conn, packet)
		}
	}
}
