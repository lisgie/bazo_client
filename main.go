package main

import (
	"crypto/ecdsa"
	"github.com/lisgie/bazo_miner/protocol"
	"github.com/lisgie/bazo_miner/p2p"
	"fmt"
	"net"
	"time"
	"math/rand"
	"math/big"
	"crypto/elliptic"
	"bytes"
	"encoding/binary"
	"golang.org/x/crypto/sha3"
)

const(
	pubA1 = "c2be9abbeaec39a066c2a09cee23bb9ab2a0b88f2880b1e785b4d317adf0dc7c"
	pubA2 = "8ce020fde838d9c443f6c93345dafe7fd74f091c4d2f30b37e2453679a257ed5"
	privA = "ba127fa8f802b008b9cdb58f4e44809d48f1b000cff750dda9cd6b312395c1c5"
	pubB1 = "5d7eefd58e3d2f309471928ab4bbd104e52973372c159fa652b8ca6b57ff68b8"
	pubB2 = "ab301a6a77b201c416ddc13a2d33fdf200a5302f6f687e0ea09085debaf8a1d9"
	privB = "7a0a9babcc97ea7991ed67ed7f800f70c5e04e99718960ad8efab2ca052f00c7"
)

const (
	VERSION_ID = 1
)

var accA, accB protocol.Account
var PrivKeyA ecdsa.PrivateKey
var RootPrivKey ecdsa.PrivateKey

func main() {
	prepAccs()

	var conn net.Conn

	conn,err := net.Dial("tcp", "127.0.0.1:8000")

	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	packet := p2p.BuildPacket(p2p.MINER_PING,nil)
	conn.Write(packet)

	buf := make([]byte, 4096)

	_,err = conn.Read(buf)
	if err != nil {
		fmt.Printf("%v\n", err)
		conn.Close()
		return
	}

	//sending a fundstx broadcast
	var txCnt uint32
	rand := rand.New(rand.NewSource(time.Now().Unix()))
	tx, _ := protocol.ConstrFundsTx(0x02,rand.Uint64()%100+1, rand.Uint64()%50+1, txCnt, accA.Hash(),accB.Hash(), &PrivKeyA)


	time.Sleep(3*time.Second)
	fmt.Printf("%v\n", tx)
	packet = p2p.BuildPacket(p2p.FUNDSTX_BRDCST,tx.Encode())
	conn.Write(packet)


	conn.Close()




	/*sendAccReq()
	sendFundsTx()
	sendAccTx()
	sendAccTx()
	sendAccTx()
	sendAccTx()
	time.Sleep(5*time.Second)
	sendAccReq()*/
}

func prepAccs() {

	puba1,_ := new(big.Int).SetString(pubA1,16)
	puba2,_ := new(big.Int).SetString(pubA2,16)
	priva,_ := new(big.Int).SetString(privA,16)
	PubKeyA := ecdsa.PublicKey{
		elliptic.P256(),
		puba1,
		puba2,
	}
	PrivKeyA = ecdsa.PrivateKey{
		PubKeyA,
		priva,
	}

	pubb1,_ := new(big.Int).SetString(pubB1,16)
	pubb2,_ := new(big.Int).SetString(pubB2,16)
	privb,_ := new(big.Int).SetString(privB,16)
	PubKeyB := ecdsa.PublicKey{
		elliptic.P256(),
		pubb1,
		pubb2,
	}
	PrivKeyB := ecdsa.PrivateKey{
		PubKeyB,
		privb,
	}

	accA = protocol.Account{Balance: 15000}
	copy(accA.Address[0:32], PrivKeyA.PublicKey.X.Bytes())
	copy(accA.Address[32:64], PrivKeyA.PublicKey.Y.Bytes())
	//accA.Hash = sha3.Sum256(accA.Address[:])


	//This one is just for testing purposes
	accB = protocol.Account{Balance: 702}
	copy(accB.Address[0:32], PrivKeyB.PublicKey.X.Bytes())
	copy(accB.Address[32:64], PrivKeyB.PublicKey.Y.Bytes())
	//accB.Hash = sha3.Sum256(accB.Address[:])


	var pubKey [64]byte

	pub1,_ := new(big.Int).SetString(protocol.RootPub1,16)
	pub2,_ := new(big.Int).SetString(protocol.RootPub2,16)
	priv,_ := new(big.Int).SetString(protocol.RootPriv,16)
	PubKeyA = ecdsa.PublicKey{
		elliptic.P256(),
		pub1,
		pub2,
	}
	RootPrivKey = ecdsa.PrivateKey{
		PubKeyA,
		priv,
	}

	copy(pubKey[32-len(pub1.Bytes()):32],pub1.Bytes())
	copy(pubKey[64-len(pub2.Bytes()):],pub2.Bytes())


	var buf bytes.Buffer

	binary.Write(&buf,binary.LittleEndian, pubKey[:])

	rootHash :=  sha3.Sum256(buf.Bytes())

	var shortRootHash [8]byte
	copy(shortRootHash[:], rootHash[0:8])
	//rootAcc := bc.Account{Hash:rootHash, Address:pubKey}
}

/*func sendAccReq() {
	var conn net.Conn
	conn, _ = net.Dial("tcp", "127.0.0.1:8081")
	header := p2p.ConstructHeader(len(accA.Hash), protocol.ACC_REQ)
	toSend := make([]byte, len(header)+len(accA.Hash))
	copy(toSend[:p2p.HEADER_LEN],header[:])
	copy(toSend[p2p.HEADER_LEN:],accA.Hash[:])

	conn.Write(toSend)

	reader := bufio.NewReader(conn)
	recvHeader := protocol.ExtractHeader(reader)

	recvBuf := make([]byte, recvHeader.Len)
	for i := 0; i < int(recvHeader.Len); i++ {
		in,_ := reader.ReadByte()
		recvBuf[i] = in
	}

	acc := protocol.DecodeAcc(recvBuf)
	fmt.Printf("%v\n", acc)
	conn.Close()
}

func sendAccTx() {
	var conn net.Conn
	rand := rand.New(rand.NewSource(time.Now().Unix()))
	conn, _ = net.Dial("tcp", "127.0.0.1:8081")
	tx,_ := protocol.ConstrAccTx(rand.Uint64()%100+1,&RootPrivKey)
	accData := protocol.EncodeAccTx(tx)
	header := protocol.ConstructHeader(len(accData), protocol.ACCTX)

	toSend := make([]byte, len(accData)+len(header))
	copy(toSend[0:protocol.HEADER_LEN],header[:])
	copy(toSend[protocol.HEADER_LEN:],accData[:])
	conn.Write(toSend)

	conn.Close()
}

func sendFundsTx() {
	var conn net.Conn
	var txCnt uint32
	rand := rand.New(rand.NewSource(time.Now().Unix()))
	conn, _ = net.Dial("tcp", "127.0.0.1:8081")
	tx, _ := protocol.ConstrFundsTx(0x02,rand.Uint64()%100+1, rand.Uint64()%50+1, txCnt, accA.Hash,accB.Hash, &PrivKeyA)
	fundsData := protocol.EncodeFundsTx(tx)
	header := protocol.ConstructHeader(len(fundsData), protocol.FUNDSTX)

	toSend := make([]byte, len(fundsData)+len(header))
	copy(toSend[0:protocol.HEADER_LEN],header[:])
	copy(toSend[protocol.HEADER_LEN:],fundsData[:])

	conn.Write(toSend)
	conn.Close()
}

func sendTimeReq() {
	var conn net.Conn

	conn, _ = net.Dial("tcp", "127.0.0.1:8081")

	header := protocol.ConstructHeader(0, protocol.TIME_REQ)
	toSend := make([]byte, protocol.HEADER_LEN)
	copy(toSend[:], header[:])
	conn.Write(toSend)

	reader := bufio.NewReader(conn)
	recvHeader := protocol.ExtractHeader(reader)

	recvBuf := make([]byte, recvHeader.Len)
	for i := 0; i < int(recvHeader.Len); i++ {
		in,_ := reader.ReadByte()
		recvBuf[i] = in
	}

	conn.Close()
}


*/
