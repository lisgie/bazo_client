package client

import (
	"fmt"
	"encoding/hex"
	"github.com/mchetelat/bazo_miner/miner"
	"errors"
)

type Account struct {
	Address       [64]byte
	AddressString string
	Balance       uint64
	TxCnt         uint32
	isCreated     bool
	isRoot        bool
}

func GetAccount(pubKey [64]byte) (*Account, error) {
	//Initialize new account with empty address
	acc := Account{pubKey, hex.EncodeToString(pubKey[:]), 0, 0, false, false}

	//Set default params
	parameters = miner.NewDefaultParameters()

	acc.isCreated, _ = isAccCreated(&acc)
	if acc.isCreated == false {
		return nil, errors.New(fmt.Sprintf("Account %x has not yet been created.\n", acc.Address))
	}

	if rootAcc := reqRootAccFromHash(serializeHashContent(acc.Address)); rootAcc != nil {
		acc.isRoot = true
	}

	acc.Balance, err = getBalance(&acc)
	if err != nil {
		return &acc, errors.New(fmt.Sprintf("Could not calculate account (%x) balance: %v\n", acc.Address, err))
	}

	return &acc, nil
}

func (acc Account) String() string {
	addressHash := serializeHashContent(acc.Address)
	return fmt.Sprintf("Hash: %x, Address: %x, TxCnt: %v, Balance: %v, isCreated: %v, isRoot: %v", addressHash[0:12], acc.Address[0:8], acc.TxCnt, acc.Balance, acc.isCreated, acc.isRoot)
}
