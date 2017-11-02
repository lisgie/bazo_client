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
	err          error
	msgType      uint8
	tx           protocol.Transaction
	myPubKey     [64]byte
	myPubKeyHash [32]byte
)

const (
	USAGE_MSG = "Usage: bazo_client [pubKey|accTx|fundsTx|configTx] ...\n"
)

func Init(keyFile string) {
	myPubKey, myPubKeyHash, err = getKeys(keyFile)
	if err != nil {
		fmt.Printf("%v\n%v", err, USAGE_MSG)
	} else {
		fmt.Printf("My Public Key: %x\n", myPubKey)
		fmt.Printf("My Public Key(Hash): %x\n", myPubKeyHash)

		for _, block := range requestRelevantBlocks() {

			fmt.Printf("%v\n", block.String())

			for _, txHash := range block.AccTxData {
				tx := requestTx(p2p.ACCTX_REQ, txHash)
				fmt.Printf("%v\n", tx)
			}

			for _, txHash := range block.FundsTxData {
				tx := requestTx(p2p.FUNDSTX_REQ, txHash)
				fmt.Printf("%v\n", tx)
			}
		}
	}
}

func requestTx(txType uint8, txHash [32]byte) (tx protocol.Transaction) {
	conn := Connect(p2p.BOOTSTRAP_SERVER)

	packet := p2p.BuildPacket(txType, txHash[:])
	n, err := conn.Write(packet)

	if n != len(packet) || err != nil {
		fmt.Printf("Transmission failed\n")
	}

	reader := bufio.NewReader(conn)
	header, _ := p2p.ReadHeader(reader)
	payload := make([]byte, header.Len)
	for cnt := 0; cnt < int(header.Len); cnt++ {
		payload[cnt], err = reader.ReadByte()
	}

	switch txType {
	case p2p.ACCTX_REQ:
		var accTx *protocol.AccTx
		accTx = accTx.Decode(payload)
		tx = accTx
	case p2p.FUNDSTX_REQ:
		var fundsTx *protocol.FundsTx
		fundsTx = fundsTx.Decode(payload)
		tx = fundsTx
	}

	conn.Close()

	return tx
}

func requestRelevantBlocks() (relevantBlocks []*protocol.Block) {
	for _, blockHash := range getRelevantBlockHashes() {
		var block *protocol.Block
		conn := Connect(p2p.BOOTSTRAP_SERVER)

		packet := p2p.BuildPacket(p2p.BLOCK_REQ, blockHash[:])
		n, err := conn.Write(packet)

		if n != len(packet) || err != nil {
			fmt.Printf("Transmission failed\n")
		}

		reader := bufio.NewReader(conn)
		header, _ := p2p.ReadHeader(reader)
		payload := make([]byte, header.Len)
		for cnt := 0; cnt < int(header.Len); cnt++ {
			payload[cnt], err = reader.ReadByte()
		}

		block = block.Decode(payload)

		relevantBlocks = append(relevantBlocks, block)

		conn.Close()

	}

	return relevantBlocks
}

func getRelevantBlockHashes() (relevantBlockHashes [][32]byte) {
	spvHeader := requestSPVHeader(nil)

	for _, pubKeyHash := range spvHeader.TxPubKeys {
		if pubKeyHash == myPubKeyHash {
			relevantBlockHashes = append(relevantBlockHashes, spvHeader.Hash)
		}
	}

	prevHash := spvHeader.PrevHash

	for spvHeader.Hash != [32]byte{} {
		spvHeader = requestSPVHeader(prevHash[:])
		for _, pubKeyHash := range spvHeader.TxPubKeys {
			if pubKeyHash == myPubKeyHash {
				relevantBlockHashes = append(relevantBlockHashes, spvHeader.Hash)
			}
		}

		prevHash = spvHeader.PrevHash
	}

	return relevantBlockHashes
}

func requestSPVHeader(blockHash []byte) (spvHeader *protocol.SPVHeader) {
	conn := Connect(p2p.BOOTSTRAP_SERVER)

	packet := p2p.BuildPacket(p2p.BLOCK_HEADER_REQ, blockHash)
	n, err := conn.Write(packet)

	if n != len(packet) || err != nil {
		fmt.Printf("Transmission failed\n")
	}

	reader := bufio.NewReader(conn)
	header, _ := p2p.ReadHeader(reader)
	payload := make([]byte, header.Len)
	for cnt := 0; cnt < int(header.Len); cnt++ {
		payload[cnt], err = reader.ReadByte()
	}

	spvHeader = spvHeader.SPVDecode(payload)

	conn.Close()

	return spvHeader
}

func getKeys(keyFile string) (myPubKey [64]byte, myPubKeyHash [32]byte, err error) {
	myKeys, err := os.Open(keyFile)
	if err != nil {
		return myPubKey, myPubKeyHash, err
	}

	reader := bufio.NewReader(myKeys)

	//We only need the public key
	pub1, _ := reader.ReadString('\n')
	pub2, _ := reader.ReadString('\n')

	pub1Int, _ := new(big.Int).SetString(strings.Split(pub1, "\n")[0], 16)
	pub2Int, _ := new(big.Int).SetString(strings.Split(pub2, "\n")[0], 16)

	copy(myPubKey[0:32], pub1Int.Bytes())
	copy(myPubKey[32:64], pub2Int.Bytes())

	myPubKeyHash = serializeHashContent(myPubKey)

	return myPubKey, myPubKeyHash, err
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
