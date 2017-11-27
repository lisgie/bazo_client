package client

import "fmt"

type Account struct {
	Address       [64]byte
	AddressString string
	Balance       uint64
	TxCnt         uint32
	isCreated     bool
	isRoot        bool
}

func (acc Account) String() string {
	addressHash := serializeHashContent(acc.Address)
	return fmt.Sprintf("Hash: %x, Address: %x, TxCnt: %v, Balance: %v, isCreated: %v, isRoot: %v", addressHash[0:12], acc.Address[0:8], acc.TxCnt, acc.Balance, acc.isCreated, acc.isRoot)
}