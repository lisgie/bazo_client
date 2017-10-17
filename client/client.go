package client

import (
	"bufio"
	"fmt"
	"github.com/mchetelat/bazo_miner/p2p"
	"github.com/mchetelat/bazo_miner/protocol"
	"net"
)

func Init() {
	conn := Connect(p2p.BOOTSTRAP_SERVER)

	packet := p2p.BuildPacket(p2p.BLOCK_HEADER_REQ, nil)
	n, err := conn.Write(packet)

	if n != len(packet) || err != nil {
		fmt.Printf("Transmission failed\n")
	}

	var spvHeader *protocol.SPVHeader

	reader := bufio.NewReader(conn)
	header, _ := p2p.ReadHeader(reader)
	payload := make([]byte, header.Len)
	for cnt := 0; cnt < int(header.Len); cnt++ {
		payload[cnt], err = reader.ReadByte()
	}

	spvHeader = spvHeader.SPVDecode(payload)

	fmt.Printf("%x\n", spvHeader.Hash)

	conn.Close()

	for spvHeader.Hash != [32]byte{} {
		conn = Connect(p2p.BOOTSTRAP_SERVER)

		packet = p2p.BuildPacket(p2p.BLOCK_HEADER_REQ, spvHeader.PrevHash[:])
		n, err = conn.Write(packet)

		if n != len(packet) || err != nil {
			fmt.Printf("Transmission failed\n")
		}

		reader = bufio.NewReader(conn)
		header, _ = p2p.ReadHeader(reader)
		payload = make([]byte, header.Len)
		for cnt := 0; cnt < int(header.Len); cnt++ {
			payload[cnt], err = reader.ReadByte()
		}

		spvHeader = spvHeader.SPVDecode(payload)

		fmt.Printf("%x\n", spvHeader.Hash)

		conn.Close()
	}
}

func Connect(connectionString string) (conn net.Conn) {
	conn, err := net.Dial("tcp", connectionString)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	return conn
}
