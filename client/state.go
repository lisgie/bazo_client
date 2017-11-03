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

		for _, txHash := range block.FundsTxData {
			tx := requestTx(p2p.FUNDSTX_REQ, txHash)
			fundsTx := tx.(*protocol.FundsTx)
			acc.Balance += fundsTx.Amount
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
