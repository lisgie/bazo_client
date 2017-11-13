package client

import (
	"github.com/mchetelat/bazo_miner/miner"
	"github.com/mchetelat/bazo_miner/p2p"
	"github.com/mchetelat/bazo_miner/protocol"
)

var (
	//If miner code is not available, a network request must be implemented
	parameters = miner.NewDefaultParameters()
)

func getAccState() error {
	//Get blocks if the acc address:
	//* issued an acc
	//* got issued as an acc
	//* created funds
	//* received funds
	//* is beneficiary
	//* nr of configTx in block is > 0 (in order to maintain params in light-client)
	for _, block := range getRelevantBlocks() {

		//Validate block integrity
		err := validateMerkleRoot(block)
		if err != nil {
			return err
		}

		//Collect block reward
		if block.Beneficiary == pubKeyHash {
			acc.Balance += parameters.Block_reward
		}

		//Check if account was issued and collect fee
		for _, txHash := range block.AccTxData {
			tx := requestTx(p2p.ACCTX_REQ, txHash)
			accTx := tx.(*protocol.AccTx)

			if accTx.PubKey == acc.Address {
				isAccCreated = true
			}

			if block.Beneficiary == pubKeyHash {
				acc.Balance += accTx.Fee
			}
		}

		//Update parameters and collect fee
		for _, txHash := range block.ConfigTxData {
			tx := requestTx(p2p.CONFIGTX_REQ, txHash)
			configTx := tx.(*protocol.ConfigTx)
			configTxSlice := []*protocol.ConfigTx{configTx}
			if block.Beneficiary == pubKeyHash {
				acc.Balance += configTx.Fee
			}

			miner.CheckAndChangeParameters(&parameters, &configTxSlice)
		}

		//Balance funds and collect fee
		for _, txHash := range block.FundsTxData {
			tx := requestTx(p2p.FUNDSTX_REQ, txHash)
			fundsTx := tx.(*protocol.FundsTx)

			//TODO: isAccRoot must reflect not only initRoot but all roots
			if isAccRoot == false {
				if fundsTx.From == pubKeyHash {
					acc.Balance -= fundsTx.Amount
					acc.Balance -= fundsTx.Fee
				} else if fundsTx.To == pubKeyHash {
					acc.Balance += fundsTx.Amount
				}
			}

			if block.Beneficiary == pubKeyHash {
				acc.Balance += fundsTx.Fee
			}
		}
	}

	return nil
}

func getRelevantBlocks() (relevantBlocks []*protocol.Block) {
	for _, blockHash := range getRelevantBlockHashes() {
		block := requestBlock(blockHash)
		relevantBlocks = append(relevantBlocks, block)
	}

	return relevantBlocks
}

func getRelevantBlockHashes() (relevantBlockHashes [][32]byte) {
	var spvHeaders []*protocol.SPVHeader
	spvHeader := requestSPVHeader(nil)
	spvHeaders = append(spvHeaders, spvHeader)
	prevHash := spvHeader.PrevHash

	for spvHeader.Hash != [32]byte{} {
		spvHeader = requestSPVHeader(prevHash[:])
		spvHeaders = append(spvHeaders, spvHeader)
		prevHash = spvHeader.PrevHash
	}

	//Order of SPVHeaders is important in order to calculate the correct parameter state for each block
	spvHeaders = miner.InvertSPVHeaderSlice(spvHeaders)

	for _, spvHeader := range spvHeaders {
		if spvHeader.BloomFilter.Test(pubKeyHash[:]) || spvHeader.Beneficiary == pubKeyHash || spvHeader.NrConfigTx > 0 {
			relevantBlockHashes = append(relevantBlockHashes, spvHeader.Hash)
		}
	}

	return relevantBlockHashes
}
