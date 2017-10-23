package client

import (
	"bufio"
	"fmt"
	"github.com/mchetelat/bazo_miner/p2p"
	"github.com/mchetelat/bazo_miner/protocol"
	"math/big"
	"os"
	"strings"
	"time"
)

var (
	err     error
	msgType uint8
	tx      protocol.Transaction
)

func Init(keyFile string) {
	myKeys, err := os.Open(keyFile)
	reader := bufio.NewReader(myKeys)

	//We only need the public key
	pub1, _ := reader.ReadString('\n')
	pub2, _ := reader.ReadString('\n')

	pub1Int, _ := new(big.Int).SetString(strings.Split(pub1, "\n")[0], 16)
	pub2Int, _ := new(big.Int).SetString(strings.Split(pub2, "\n")[0], 16)

	var myPubKey [64]byte
	copy(myPubKey[0:32], pub1Int.Bytes())
	copy(myPubKey[32:64], pub2Int.Bytes())

	fmt.Printf("My Public Key: %x\n", myPubKey)
	fmt.Printf("My Public Key(Hash): %x\n", serializeHashContent(myPubKey))

	conn := Connect(p2p.BOOTSTRAP_SERVER)

	packet := p2p.BuildPacket(p2p.BLOCK_HEADER_REQ, nil)
	n, err := conn.Write(packet)

	if n != len(packet) || err != nil {
		fmt.Printf("Transmission failed\n")
	}

	var spvHeader *protocol.SPVHeader

	reader = bufio.NewReader(conn)
	header, _ := p2p.ReadHeader(reader)
	payload := make([]byte, header.Len)
	for cnt := 0; cnt < int(header.Len); cnt++ {
		payload[cnt], err = reader.ReadByte()
	}

	spvHeader = spvHeader.SPVDecode(payload)

	//fmt.Printf("%x\n", spvHeader.Hash)

	for _, pubKey := range spvHeader.TxPubKeys {
		fmt.Printf("Public Key: %x in Block %x\n", pubKey, spvHeader.Hash)
	}

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

		//fmt.Printf("%x\n", spvHeader.Hash)

		for _, pubKey := range spvHeader.TxPubKeys {
			fmt.Printf("Public Key: %x in Block %x\n", pubKey, spvHeader.Hash)
		}

		conn.Close()
	}
}

func Process(args []string) {
	switch args[0] {
	case "accTx":
		tx, err = parseAccTx(os.Args[2:])
		msgType = p2p.ACCTX_BRDCST
	case "fundsTx":
		tx, err = parseFundsTx(os.Args[2:])
		msgType = p2p.FUNDSTX_BRDCST
	case "configTx":
		tx, err = parseConfigTx(os.Args[2:])
		msgType = p2p.CONFIGTX_BRDCST
	default:
		fmt.Printf("Usage: bazo_client [accTx|fundsTx|configTx] ...\n")
		return
	}

	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	//Transaction creation successful
	packet := p2p.BuildPacket(msgType, tx.Encode())

	//Open a connection
	conn := Connect(p2p.BOOTSTRAP_SERVER)
	n, err := conn.Write(packet)

	if n != len(packet) || err != nil {
		fmt.Printf("Transmission failed\n")
	}

	fmt.Printf("Successfully sent the following tansaction:\n%v\n", tx)

	//Wait for response
	start := time.Now()
	for {
		//Time out after 10 seconds
		if time.Since(start).Seconds() > 10 {
			fmt.Printf("Connection to %v aborted: (TimeOut)\n", p2p.BOOTSTRAP_SERVER)
			break
		}

		reader := bufio.NewReader(conn)
		header, _ := p2p.ReadHeader(reader)

		if header != nil && header.TypeID == p2p.TX_BRDCST_ACK {
			fmt.Printf("Transaction successfully processed by network\n")
			break
		}
	}

	conn.Close()
}
