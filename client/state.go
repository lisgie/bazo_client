package client

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/mchetelat/bazo_miner/miner"
	"github.com/mchetelat/bazo_miner/p2p"
	"github.com/mchetelat/bazo_miner/protocol"
	"time"
)

var (
	//If miner code is not available, a network request must be implemented
	parameters     = miner.NewDefaultParameters()
	allBockHeaders []*protocol.SPVHeader
)

func initState() {
	loadAllBlockHeaders()
	allBockHeaders = miner.InvertSPVHeaderSlice(allBockHeaders)
}

func getAccState(pubKey [64]byte) (*Account, error) {
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

func isAccCreated(acc *Account) (bool, error) {
	relevantBlocks, err := getRelevantBlocks(acc.Address)
	if err != nil {
		return false, err
	}

	for _, block := range relevantBlocks {
		for _, txHash := range block.AccTxData {
			tx := requestTx(p2p.ACCTX_REQ, txHash)
			AccTx := tx.(*protocol.AccTx)
			if AccTx.PubKey == acc.Address {
				return true, nil
			}
		}
	}

	return false, nil
}

func getBalance(acc *Account) (balance uint64, err error) {
	pubKeyHash := serializeHashContent(acc.Address)

	//Get blocks if the Acc address:
	//* issued an Acc
	//* got issued as an Acc
	//* created funds
	//* received funds
	//* is beneficiary
	//* nr of configTx in block is > 0 (in order to maintain params in light-client)
	relevantBlocks, err := getRelevantBlocks(acc.Address)
	if err != nil {
		return balance, err
	}

	for _, block := range relevantBlocks {
		//Collect block reward
		if block.Beneficiary == serializeHashContent(acc.Address) {
			balance += parameters.Block_reward
		}

		//Check if Account was issued and collect fee
		for _, txHash := range block.AccTxData {
			tx := requestTx(p2p.ACCTX_REQ, txHash)
			AccTx := tx.(*protocol.AccTx)

			if block.Beneficiary == pubKeyHash {
				balance += AccTx.Fee
			}
		}

		//Update config parameters and collect fee
		for _, txHash := range block.ConfigTxData {
			tx := requestTx(p2p.CONFIGTX_REQ, txHash)
			configTx := tx.(*protocol.ConfigTx)
			configTxSlice := []*protocol.ConfigTx{configTx}

			if block.Beneficiary == pubKeyHash {
				balance += configTx.Fee
			}

			miner.CheckAndChangeParameters(&parameters, &configTxSlice)
		}

		//Balance funds and collect fee
		for _, txHash := range block.FundsTxData {
			tx := requestTx(p2p.FUNDSTX_REQ, txHash)
			fundsTx := tx.(*protocol.FundsTx)
			//If Acc is no root, balance funds
			if !acc.isRoot {
				if fundsTx.From == pubKeyHash {
					balance -= fundsTx.Amount
					balance -= fundsTx.Fee
				} else if fundsTx.To == pubKeyHash {
					balance += fundsTx.Amount
				}
			}

			if block.Beneficiary == pubKeyHash {
				balance += fundsTx.Fee
			}
		}
	}

	return balance, nil
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
