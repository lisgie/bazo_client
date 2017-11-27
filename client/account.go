package client

type Account struct {
	Address   string
	Balance   uint64
	TxCnt     uint32
	isCreated bool
	isRoot    bool
}
