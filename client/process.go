package client

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/mchetelat/bazo_miner/protocol"
	"math/big"
	"os"
	"strconv"
	"strings"
)

const (
	ARGS_MSG = "Wrong number of arguments."
)

func parseAccTx(args []string) (protocol.Transaction, error) {
	accTxUsage := "\nUsage: bazo_client accTx <header> <fee> <privKey> <keyOutput>"

	if len(args) != 4 {
		return nil, errors.New(fmt.Sprintf("%v%v", ARGS_MSG, accTxUsage))
	}

	header, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, accTxUsage))
	}

	fee, err := strconv.Atoi(args[1])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, accTxUsage))
	}

	_, privKey, err := extractKeyFromFile(args[2])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, accTxUsage))
	}

	tx, newKey, err := protocol.ConstrAccTx(byte(header), uint64(fee), &privKey)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, accTxUsage))
	}

	if tx == nil {
		return nil, errors.New(fmt.Sprintf("Transaction encoding failed.%v", accTxUsage))
	}

	//Write the public key to the given textfile
	if _, err = os.Stat(args[3]); !os.IsNotExist(err) {
		return nil, errors.New(fmt.Sprintf("Output file exists.%v", accTxUsage))
	}

	file, err := os.Create(args[3])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, accTxUsage))
	}

	_, err = file.WriteString(string(newKey.X.Text(16)) + "\n")
	_, err2 := file.WriteString(string(newKey.Y.Text(16)) + "\n")
	_, err3 := file.WriteString(string(newKey.D.Text(16)) + "\n")

	if err != nil || err2 != nil || err3 != nil {
		return nil, errors.New(fmt.Sprintf("Failed to write key to file%v", accTxUsage))
	}

	return tx, nil
}

func parseFundsTx(args []string) (protocol.Transaction, error) {
	fundsTxUsage := "\nUsage: bazo_client fundsTx <header> <amount> <fee> <txCnt> <fromHash> <toHash> <privKey>"

	var (
		fromPubKey, toPubKey [64]byte
	)

	if len(args) != 7 {
		return nil, errors.New(fmt.Sprintf("%v%v", ARGS_MSG, fundsTxUsage))
	}

	header, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, fundsTxUsage))
	}

	amount, err := strconv.Atoi(args[1])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, fundsTxUsage))
	}

	fee, err := strconv.Atoi(args[2])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, fundsTxUsage))
	}

	txCnt, err := strconv.Atoi(args[3])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, fundsTxUsage))
	}

	hashFromFile, err := os.Open(args[4])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, fundsTxUsage))
	}

	reader := bufio.NewReader(hashFromFile)
	//We only need the public key
	pub1, err := reader.ReadString('\n')
	pub2, err2 := reader.ReadString('\n')
	if err != nil || err2 != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, fundsTxUsage))
	}

	pub1Int, _ := new(big.Int).SetString(strings.Split(pub1, "\n")[0], 16)
	pub2Int, _ := new(big.Int).SetString(strings.Split(pub2, "\n")[0], 16)
	copy(fromPubKey[0:32], pub1Int.Bytes())
	copy(fromPubKey[32:64], pub2Int.Bytes())

	hashToFile, err := os.Open(args[5])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, fundsTxUsage))
	}

	reader.Reset(hashToFile)
	//We only need the public key
	pub1, err = reader.ReadString('\n')
	pub2, err2 = reader.ReadString('\n')
	if err != nil || err2 != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, fundsTxUsage))
	}

	pub1Int, _ = new(big.Int).SetString(strings.Split(pub1, "\n")[0], 16)
	pub2Int, _ = new(big.Int).SetString(strings.Split(pub2, "\n")[0], 16)
	copy(toPubKey[0:32], pub1Int.Bytes())
	copy(toPubKey[32:64], pub2Int.Bytes())

	_, privKey, err := extractKeyFromFile(args[6])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, fundsTxUsage))
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

	fmt.Printf("%x")

	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", err, fundsTxUsage))
	}

	if tx == nil {
		return nil, errors.New(fmt.Sprintf("Transaction encoding failed.%v", fundsTxUsage))
	}

	return tx, nil
}

func parseConfigTx(args []string) (protocol.Transaction, error) {
	configTxUsage := "\nUsage: bazo_client configTx <header> <id> <payload> <fee> <txCnt> <privKey>"

	if len(args) != 6 {
		return nil, errors.New(fmt.Sprintf("%v%v", ARGS_MSG, configTxUsage))
	}

	header, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", ARGS_MSG, configTxUsage))
	}

	id, err := strconv.Atoi(args[1])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", ARGS_MSG, configTxUsage))
	}

	payload, err := strconv.Atoi(args[2])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", ARGS_MSG, configTxUsage))
	}

	fee, err := strconv.Atoi(args[3])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", ARGS_MSG, configTxUsage))
	}

	txCnt, err := strconv.Atoi(args[4])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", ARGS_MSG, configTxUsage))
	}

	_, privKey, err := extractKeyFromFile(args[5])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v%v", ARGS_MSG, configTxUsage))
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
		return nil, errors.New(fmt.Sprintf("%v%v", ARGS_MSG, configTxUsage))
	}

	if tx == nil {
		return nil, errors.New(fmt.Sprintf("Transaction encoding failed.%v", configTxUsage))
	}

	return tx, nil
}
