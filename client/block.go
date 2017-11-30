package client

import (
	"errors"
	"fmt"
	"github.com/mchetelat/bazo_miner/miner"
	"github.com/mchetelat/bazo_miner/protocol"
)

func validateMerkleRoot(block *protocol.Block) error {
	var txHashSlice [][32]byte

	for _, txHash := range block.AccTxData {
		txHashSlice = append(txHashSlice, txHash)
	}
	for _, txHash := range block.FundsTxData {
		txHashSlice = append(txHashSlice, txHash)
	}
	for _, txHash := range block.ConfigTxData {
		txHashSlice = append(txHashSlice, txHash)
	}

	if block.MerkleRoot != miner.BuildMerkleTree(txHashSlice) {
		return errors.New(fmt.Sprintf("Block %x cannot be validated: Expected Merkle root cannot be recalculated.\n", block.Hash))
	}

	return nil
}
