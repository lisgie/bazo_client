package client

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/mchetelat/bazo_miner/miner"
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

	//If Acc is Root in the bazo network state, we do not check for accTx, else we check
	if rootAcc := reqRootAccFromHash(serializeHashContent(acc.Address)); rootAcc != nil {
		acc.isCreated, acc.isRoot = true, true
	} else {
		if acc.isCreated, _ = isAccCreated(&acc); acc.isCreated == false {
			return nil, errors.New(fmt.Sprintf("Account %x does not exist.\n", acc.Address[:8]))
		}
	}

	acc.Balance, err = getBalance(&acc)
	if err != nil {
		return &acc, errors.New(fmt.Sprintf("Could not calculate account (%x) balance: %v\n", acc.Address[:8], err))
	}

	return &acc, nil
}

func (acc Account) String() string {
	addressHash := serializeHashContent(acc.Address)
	return fmt.Sprintf("Hash: %x, Address: %x, TxCnt: %v, Balance: %v, isCreated: %v, isRoot: %v", addressHash[:8], acc.Address[:8], acc.TxCnt, acc.Balance, acc.isCreated, acc.isRoot)
}
