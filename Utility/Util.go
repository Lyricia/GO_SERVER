package Utility

import (
	"bytes"
	"encoding/binary"
	"log"
	"net"
)

func PacketSerialization(packet interface{}) *bytes.Buffer {
	buf := &bytes.Buffer{}
	err := binary.Write(buf, binary.LittleEndian, packet)

	if err != nil {
		log.Fatalln("error buffer")
	}

	t := buf.Bytes()
	t[0] = byte(len(t))

	return buf
}

func SendPacket(conn net.Conn, packet interface{}) {
	buf := PacketSerialization(packet)

	conn.Write(buf.Bytes())
}
