package client

import (
	"github.com/mchetelat/bazo_miner/miner"
	"github.com/mchetelat/bazo_miner/p2p"
	"github.com/mchetelat/bazo_miner/protocol"
	"encoding/hex"
	"time"
)

var (
	//If miner code is not available, a network request must be implemented
	parameters = miner.NewDefaultParameters()
	allBockHeaders []*protocol.SPVHeader
)

func initState() {
	loadAllBlockHeaders()
	allBockHeaders = miner.InvertSPVHeaderSlice(allBockHeaders)
}

func getAccState(pubKey [64]byte) *Account {
	//Initialize new account with empty address
	acc := Account{hex.EncodeToString(pubKey[:]), 0, 0, false, false}

	//Set default params
	parameters = miner.NewDefaultParameters()

	if rootAcc := reqRootAccFromHash(serializeHashContent(pubKey)); rootAcc != nil {
		acc.isRoot = true
	}

	isAccCreated(pubKey)
	balance, err := getBalance(pubKey)
	if err != nil {
	}

	if Acc.Address != [64]byte{} {
		logger.Println(Acc.String())
	} else {
		logger.Println("Account does not exist.")
	}

}

func isAccCreated(pubKey [64]byte) (bool, error) {

}

func getBalance(acc *Account) (balance uint64, err error) {
	var pubKey [64]byte
	hex.Decode(pubKey[:64], []byte(acc.Address))
	pubKeyHash := serializeHashContent(pubKey)

	//Get blocks if the Acc address:
	//* issued an Acc
	//* got issued as an Acc
	//* created funds
	//* received funds
	//* is beneficiary
	//* nr of configTx in block is > 0 (in order to maintain params in light-client)
	relevantBlocks, err := getRelevantBlocks(pubKey)
	if err != nil {
		return balance, err
	}

	for _, block := range relevantBlocks {
		//Collect block reward
		if block.Beneficiary == serializeHashContent(pubKey) {
			balance += parameters.Block_reward
		}

		//Check if Account was issued and collect fee
		for _, txHash := range block.AccTxData {
			tx := requestTx(p2p.ACCTX_REQ, txHash)
			AccTx := tx.(*protocol.AccTx)

			if AccTx.PubKey == pubKey {
				if rootAcc := reqRootAccFromHash(pubKeyHash); rootAcc != nil {
					isRootAcc = true
				}
			}

			if block.Beneficiary == pubKeyHash {
				Acc.Balance += AccTx.Fee
			}
		}

		//Update config parameters and collect fee
		for _, txHash := range block.ConfigTxData {
			tx := requestTx(p2p.CONFIGTX_REQ, txHash)
			configTx := tx.(*protocol.ConfigTx)
			configTxSlice := []*protocol.ConfigTx{configTx}

			if block.Beneficiary == pubKeyHash {
				Acc.Balance += configTx.Fee
			}

			miner.CheckAndChangeParameters(&parameters, &configTxSlice)
		}

		//Balance funds and collect fee
		for _, txHash := range block.FundsTxData {
			tx := requestTx(p2p.FUNDSTX_REQ, txHash)
			fundsTx := tx.(*protocol.FundsTx)
			//If Acc is no root, balance funds
			if !isRootAcc {
				if fundsTx.From == pubKeyHash {
					Acc.Balance -= fundsTx.Amount
					Acc.Balance -= fundsTx.Fee
				} else if fundsTx.To == pubKeyHash {
					Acc.Balance += fundsTx.Amount
				}
			}

			if block.Beneficiary == pubKeyHash {
				Acc.Balance += fundsTx.Fee
			}
		}
	}

	return nil
}

func getRelevantBlocks(pubKey [64]byte) (relevantBlocks []*protocol.Block, err error) {
	for _, blockHash := range getRelevantBlockHashes(pubKey) {
		block := requestBlock(blockHash)

		//Validate block integrity
		err := validateMerkleRoot(block)
		if err != nil {
			return nil, err
		}

		relevantBlocks = append(relevantBlocks, block)
	}

	return relevantBlocks, nil
}

func getRelevantBlockHashes(pubKey [64]byte) (relevantBlockHashes [][32]byte) {
	pubKeyHash := serializeHashContent(pubKey)
	for _, spvHeader := range allBockHeaders {
		if spvHeader.BloomFilter.Test(pubKeyHash[:]) || spvHeader.Beneficiary == pubKeyHash || spvHeader.NrConfigTx > 0 {
			relevantBlockHashes = append(relevantBlockHashes, spvHeader.Hash)
		}
	}

	return relevantBlockHashes
}

func loadAllBlockHeaders() {
	spvHeader := requestSPVHeader(nil)
	allBockHeaders = append(allBockHeaders, spvHeader)
	prevHash := spvHeader.PrevHash

	for spvHeader.Hash != [32]byte{} {
		spvHeader = requestSPVHeader(prevHash[:])
		allBockHeaders = append(allBockHeaders, spvHeader)
		prevHash = spvHeader.PrevHash
	}
}