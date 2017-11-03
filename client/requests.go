package client

import (
	"fmt"
	"github.com/mchetelat/bazo_miner/p2p"
	"github.com/mchetelat/bazo_miner/protocol"
)

func requestBlock(blockHash [32]byte) (block *protocol.Block) {

	packet := p2p.BuildPacket(p2p.BLOCK_REQ, blockHash[:])
	conn.Write(packet)

	header, payload, err := rcvData(conn)
	if err != nil {
		fmt.Printf("Disconnected: %v\n", err)
		return
	}

	if header.TypeID == p2p.BLOCK_RES {
		block = block.Decode(payload)
	}

	return block
}

func requestTx(txType uint8, txHash [32]byte) (tx protocol.Transaction) {

	packet := p2p.BuildPacket(txType, txHash[:])
	conn.Write(packet)

	header, payload, err := rcvData(conn)
	if err != nil {
		fmt.Printf("Disconnected: %v\n", err)
		return
	}

	switch header.TypeID {
	case p2p.ACC_RES:
		var accTx *protocol.AccTx
		accTx = accTx.Decode(payload)
		tx = accTx
	case p2p.FUNDSTX_RES:
		var fundsTx *protocol.FundsTx
		fundsTx = fundsTx.Decode(payload)
		tx = fundsTx
	}

	return tx
}

func requestSPVHeader(blockHash []byte) (spvHeader *protocol.SPVHeader) {

	packet := p2p.BuildPacket(p2p.BLOCK_HEADER_REQ, blockHash)
	conn.Write(packet)

	header, payload, err := rcvData(conn)
	if err != nil {
		fmt.Printf("Disconnected: %v\n", err)
		return
	}

	if header.TypeID == p2p.BlOCK_HEADER_RES {
		spvHeader = spvHeader.SPVDecode(payload)
	}

	return spvHeader
}
