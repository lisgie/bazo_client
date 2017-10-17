package main

import (
	"bufio"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/mchetelat/bazo_client/client"
	"github.com/mchetelat/bazo_miner/p2p"
	"github.com/mchetelat/bazo_miner/protocol"
	"golang.org/x/crypto/sha3"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {

	var (
		err     error
		msgType uint8
		tx      protocol.Transaction
	)

	client.Init()

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "accTx":
			_, err := os.Stat(os.Args[4])
			if err != nil {
				err = generateKeyPair(os.Args[4])
			}
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
		conn := client.Connect(p2p.BOOTSTRAP_SERVER)
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
}

func parseAccTx(args []string) (protocol.Transaction, error) {

	if len(args) != 4 {
		return nil, errors.New("Wrong number of arguments.\nUsage: bazo_client accTx <header> <fee> <privKey> <keyOutput>")
	}

	header, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v\nUsage: bazo_client accTx <header> <fee> <privKey> <keyOutput>", err))
	}

	fee, err := strconv.Atoi(args[1])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v\nUsage: bazo_client accTx <header> <fee> <privKey> <keyOutput>", err))
	}

	_, privKey, err := extractKeyFromFile(args[2])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v\nUsage: bazo_client accTx <header> <fee> <privKey> <keyOutput>", err))
	}

	tx, newKey, err := protocol.ConstrAccTx(byte(header), uint64(fee), &privKey)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v\nUsage: bazo_client accTx <header> <fee> <privKey> <keyOutput>", err))
	}

	if tx == nil {
		return nil, errors.New("Transaction encoding failed.\nUsage: bazo_client accTx <header> <fee> <privKey> <keyOutput>")
	}

	//Write the public key to the given textfile
	if _, err = os.Stat(args[3]); !os.IsNotExist(err) {
		return nil, errors.New("Output file exists.\nUsage: bazo_client accTx <header> <fee> <privKey> <keyOutput>")
	}

	file, err := os.Create(args[3])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v\nUsage: bazo_client accTx <header> <fee> <privKey> <keyOutput>", err))
	}

	_, err = file.WriteString(string(newKey.X.Text(16)) + "\n")
	_, err2 := file.WriteString(string(newKey.Y.Text(16)) + "\n")
	_, err3 := file.WriteString(string(newKey.D.Text(16)) + "\n")

	if err != nil || err2 != nil || err3 != nil {
		return nil, errors.New("Failed to write key to file\nUsage: bazo_client accTx <header> <fee> <privKey> <keyOutput>")
	}

	return tx, nil
}

func parseFundsTx(args []string) (protocol.Transaction, error) {

	var (
		fromPubKey, toPubKey [64]byte
	)

	if len(args) != 7 {
		return nil, errors.New("Wrong number of arguments.\nUsage: bazo_client fundsTx <header> <amount> <fee> <txCnt> <fromHash> <toHash> <privKey>")
	}

	header, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v\nUsage: bazo_client fundsTx <header> <amount> <fee> <txCnt> <fromHash> <toHash> <privKey>", err))
	}

	amount, err := strconv.Atoi(args[1])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v\nUsage: bazo_client fundsTx <header> <amount> <fee> <txCnt> <fromHash> <toHash> <privKey>", err))
	}

	fee, err := strconv.Atoi(args[2])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v\nUsage: bazo_client fundsTx <header> <amount> <fee> <txCnt> <fromHash> <toHash> <privKey>", err))
	}

	txCnt, err := strconv.Atoi(args[3])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v\nUsage: bazo_client fundsTx <header> <amount> <fee> <txCnt> <fromHash> <toHash> <privKey>", err))
	}

	hashFromFile, err := os.Open(args[4])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v\nUsage: bazo_client fundsTx <header> <amount> <fee> <txCnt> <fromHash> <toHash> <privKey>", err))
	}

	reader := bufio.NewReader(hashFromFile)
	//We only need the public key
	pub1, err := reader.ReadString('\n')
	pub2, err2 := reader.ReadString('\n')
	if err != nil || err2 != nil {
		return nil, errors.New(fmt.Sprintf("%v\nUsage: bazo_client fundsTx <header> <amount> <fee> <txCnt> <fromHash> <toHash> <privKey>", err))
	}

	pub1Int, _ := new(big.Int).SetString(strings.Split(pub1, "\n")[0], 16)
	pub2Int, _ := new(big.Int).SetString(strings.Split(pub2, "\n")[0], 16)
	copy(fromPubKey[0:32], pub1Int.Bytes())
	copy(fromPubKey[32:64], pub2Int.Bytes())

	hashToFile, err := os.Open(args[5])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v\nUsage: bazo_client fundsTx <header> <amount> <fee> <txCnt> <fromHash> <toHash> <privKey>", err))
	}

	reader.Reset(hashToFile)
	//We only need the public key
	pub1, err = reader.ReadString('\n')
	pub2, err2 = reader.ReadString('\n')
	if err != nil || err2 != nil {
		return nil, errors.New(fmt.Sprintf("%v\nUsage: bazo_client fundsTx <header> <amount> <fee> <txCnt> <fromHash> <toHash> <privKey>", err))
	}

	pub1Int, _ = new(big.Int).SetString(strings.Split(pub1, "\n")[0], 16)
	pub2Int, _ = new(big.Int).SetString(strings.Split(pub2, "\n")[0], 16)
	copy(toPubKey[0:32], pub1Int.Bytes())
	copy(toPubKey[32:64], pub2Int.Bytes())

	_, privKey, err := extractKeyFromFile(args[6])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v\nUsage: bazo_client fundsTx <header> <amount> <fee> <txCnt> <fromHash> <toHash> <privKey>", err))
	}

	tx, err := protocol.ConstrFundsTx(
		byte(header),
		uint64(amount),
		uint64(fee),
		uint32(txCnt),
		serializeHashContent(fromPubKey[:]),
		serializeHashContent(toPubKey[:]),
		&privKey,
	)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v\nUsage: bazo_client fundsTx <header> <amount> <fee> <txCnt> <fromHash> <toHash> <privKey>", err))
	}

	if tx == nil {
		return nil, errors.New("Transaction encoding failed.\nUsage: bazo_client fundsTx <header> <amount> <fee> <txCnt> <fromHash> <toHash> <privKey>")
	}

	return tx, nil
}

func parseConfigTx(args []string) (protocol.Transaction, error) {

	if len(args) != 6 {
		return nil, errors.New("Wrong number of arguments.\nUsage: bazo_client configTx <header> <id> <payload> <fee> <txCnt> <privKey>")
	}

	header, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v\nUsage: bazo_client configTx <header> <id> <payload> <fee> <txCnt> <privKey>", err))
	}

	id, err := strconv.Atoi(args[1])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v\nUsage: bazo_client configTx <header> <id> <payload> <fee> <txCnt> <privKey>", err))
	}

	payload, err := strconv.Atoi(args[2])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v\nUsage: bazo_client configTx <header> <id> <payload> <fee> <txCnt> <privKey>", err))
	}

	fee, err := strconv.Atoi(args[3])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v\nUsage: bazo_client configTx <header> <id> <payload> <fee> <txCnt> <privKey>", err))
	}

	txCnt, err := strconv.Atoi(args[4])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v\nUsage: bazo_client configTx <header> <id> <payload> <fee> <txCnt> <privKey>", err))
	}

	_, privKey, err := extractKeyFromFile(args[5])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v\nUsage: bazo_client configTx <header> <id> <payload> <fee> <txCnt> <privKey>", err))
	}

	tx, err := protocol.ConstrConfigTx(
		byte(header),
		uint8(id),
		uint64(payload),
		uint64(fee),
		uint8(txCnt),
		&privKey,
	)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v\nUsage: bazo_client configTx <header> <id> <payload> <fee> <txCnt> <privKey>", err))
	}

	if tx == nil {
		return nil, errors.New("Transaction encoding failed.\nUsage: bazo_client configTx <header> <id> <payload> <fee> <txCnt> <privKey>")
	}

	return tx, nil
}

func extractKeyFromFile(filename string) (pubKey ecdsa.PublicKey, privKey ecdsa.PrivateKey, err error) {

	filehandle, err := os.Open(filename)
	if err != nil {
		return pubKey, privKey, errors.New(fmt.Sprintf("%v", err))
	}

	reader := bufio.NewReader(filehandle)

	//Public Key
	pub1, err := reader.ReadString('\n')
	pub2, err2 := reader.ReadString('\n')
	//Private Key
	priv, err3 := reader.ReadString('\n')
	if err != nil || err2 != nil {
		return pubKey, privKey, errors.New(fmt.Sprintf("Could not read key from file: %v", err))
	}

	pub1Int, b := new(big.Int).SetString(strings.Split(pub1, "\n")[0], 16)
	pub2Int, b2 := new(big.Int).SetString(strings.Split(pub2, "\n")[0], 16)

	pubKey = ecdsa.PublicKey{
		elliptic.P256(),
		pub1Int,
		pub2Int,
	}

	//File consists of public & private key
	if err3 == nil {
		privInt, b3 := new(big.Int).SetString(strings.Split(priv, "\n")[0], 16)
		if !b || !b2 || !b3 {
			return pubKey, privKey, errors.New("Failed to convert the key strings to big.Int.")
		}

		privKey = ecdsa.PrivateKey{
			pubKey,
			privInt,
		}
	}

	return pubKey, privKey, nil
}

func serializeHashContent(data interface{}) (hash [32]byte) {

	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, data)
	return sha3.Sum256(buf.Bytes())
}

func generateKeyPair(filename string) error {

	newRootKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	if err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return errors.New(fmt.Sprintf("%v\nUsage: bazo_client accTx <header> <fee> <privKey> <keyOutput>", err))
	}

	var rootPublicKey [64]byte
	rootPubKey1, rootPubKey2 := newRootKey.PublicKey.X.Bytes(), newRootKey.PublicKey.Y.Bytes()
	copy(rootPublicKey[32-len(rootPubKey1):32], rootPubKey1)
	copy(rootPublicKey[64-len(rootPubKey2):], rootPubKey2)

	hash := serializeHashContent(rootPublicKey)

	_, err = file.WriteString(string(newRootKey.X.Text(16)) + "\n")
	_, err2 := file.WriteString(string(newRootKey.Y.Text(16)) + "\n")
	_, err3 := file.WriteString(string(newRootKey.D.Text(16)) + "\n")
	fmt.Printf("BeneficiaryHash: %x\n", hash)

	if err != nil || err2 != nil || err3 != nil {
		return errors.New("Failed to write key to file\nUsage: bazo_client accTx <header> <fee> <privKey> <keyOutput>")
	}

	return nil
}
