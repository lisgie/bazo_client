package client

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"github.com/mchetelat/bazo_miner/miner"
	"github.com/mchetelat/bazo_miner/protocol"
	"testing"
)

func TestValidateMerkleRoot(t *testing.T) {
	var hashSlice1 [][32]byte
	var hashSlice2 [][32]byte
	var hashSlice3 [][32]byte
	var tx1 *protocol.FundsTx
	var tx2 *protocol.AccTx
	var tx3 *protocol.ConfigTx

	//Generating a private key and prepare data
	privA, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	tx1, _ = protocol.ConstrFundsTx(0x01, 23, 1, 0, [32]byte{'0'}, [32]byte{'1'}, privA)
	tx2, _, _ = protocol.ConstrAccTx(0, 23, privA)
	tx3, _ = protocol.ConstrConfigTx(0x02, 2, 5000, 34, 0, privA)

	hashSlice1 = append(hashSlice1, tx1.Hash())
	hashSlice2 = append(hashSlice2, tx2.Hash())
	hashSlice3 = append(hashSlice3, tx3.Hash())

	block := new(protocol.Block)

	block.AccTxData = hashSlice1
	block.FundsTxData = hashSlice2
	block.ConfigTxData = hashSlice3

	block.MerkleRoot = miner.BuildMerkleTree(block.AccTxData, block.FundsTxData, block.ConfigTxData)

	if err := validateMerkleRoot(block); err != nil {
		t.Error("Expected valid block merkleroot, got error")
	}

	fakedBlock := block

	//Alter tx2, set header to 1
	tx2, _, _ = protocol.ConstrAccTx(1, 23, privA)
	hashSlice2[0] = tx2.Hash()
	fakedBlock.AccTxData = hashSlice2

	if err := validateMerkleRoot(fakedBlock); err == nil {
		t.Error("Expected invalid block merkleroot, got valid")
	}
}
