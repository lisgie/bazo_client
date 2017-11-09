package client

import (
	"fmt"
	"github.com/mchetelat/bazo_miner/p2p"
	"github.com/mchetelat/bazo_miner/protocol"
	"log"
	"os"
	"time"
)

var (
	acc         protocol.Account
	acc_created = false
	err         error
	msgType     uint8
	pubKey      [64]byte
	pubKeyHash  [32]byte
	tx          protocol.Transaction
	logger      *log.Logger
)

const (
	USAGE_MSG = "Usage: bazo_client [pubKey|accTx|fundsTx|configTx] ...\n"
)

func Init() {
	logger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func State(keyFile string) {
	pubKey, pubKeyHash, err = getKeys(keyFile)
	if err != nil {
		fmt.Printf("%v\n%v", err, USAGE_MSG)
	} else {
		fmt.Printf("My Public Key: %x\n", pubKey)
		fmt.Printf("My Public Key(Hash): %x\n", pubKeyHash)

		acc.Address = pubKey

		for {
			acc.Balance = 0

			err := getAccState()
			if err != nil {
				println(err)
				break
			}

			if acc_created {
				logger.Println(acc.String())
			} else {
				logger.Println("Account does not exist.")
			}

			time.Sleep(20 * time.Second)
		}
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
	conn.Write(packet)

	header, _, err := rcvData(conn)
	if err != nil {
		fmt.Printf("Disconnected: %v\n", err)
		return
	}

	if header != nil && header.TypeID == p2p.TX_BRDCST_ACK {
		fmt.Printf("Successfully sent the following tansaction:%v", tx)
	}

	conn.Close()
}
