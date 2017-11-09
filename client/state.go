package client

import (
	"github.com/mchetelat/bazo_miner/p2p"
	"github.com/mchetelat/bazo_miner/protocol"
)

func getAccState() error {
	for _, block := range getRelevantBlocks() {

		err := validateMerkleRoot(block)
		if err != nil {
			return err
		}

		if !acc_created {
			for _, txHash := range block.AccTxData {
				tx := requestTx(p2p.ACCTX_REQ, txHash)
				accTx := tx.(*protocol.AccTx)
				if accTx.PubKey == acc.Address {
					acc_created = true
				}
			}
		}

		for _, txHash := range block.FundsTxData {
			tx := requestTx(p2p.FUNDSTX_REQ, txHash)
			fundsTx := tx.(*protocol.FundsTx)
			if fundsTx.From == pubKeyHash {
				acc.Balance -= fundsTx.Amount
				acc.Balance -= fundsTx.Fee
			} else if fundsTx.To == pubKeyHash {
				acc.Balance += fundsTx.Amount
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
	spvHeader := requestSPVHeader(nil)

	if spvHeader.BloomFilter.Test(pubKeyHash[:]) {
		relevantBlockHashes = append(relevantBlockHashes, spvHeader.Hash)
	}

	prevHash := spvHeader.PrevHash

	for spvHeader.Hash != [32]byte{} {
		spvHeader = requestSPVHeader(prevHash[:])
		if spvHeader.BloomFilter.Test(pubKeyHash[:]) {
			relevantBlockHashes = append(relevantBlockHashes, spvHeader.Hash)
		}

		prevHash = spvHeader.PrevHash
	}

	return relevantBlockHashes
}
