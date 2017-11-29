package REST

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"github.com/gorilla/mux"
	"github.com/mchetelat/bazo_client/client"
	"github.com/mchetelat/bazo_miner/p2p"
	"github.com/mchetelat/bazo_miner/protocol"
	"math/big"
	"net/http"
	"strconv"
)

func CreateAccTxEndpoint(w http.ResponseWriter, req *http.Request) {

}

func CreateConfigTxEndpoint(w http.ResponseWriter, req *http.Request) {

}

func CreateFundsTxEndpoint(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)

	var fromPub [64]byte
	var toPub [64]byte

	//TODO: Error logging
	amount, _ := strconv.Atoi(params["amount"])
	fee, _ := strconv.Atoi(params["fee"])
	txCnt, _ := strconv.Atoi(params["txCnt"])

	fromPubInt, _ := new(big.Int).SetString(params["fromPub"], 16)
	copy(fromPub[:], fromPubInt.Bytes())

	toPubInt, _ := new(big.Int).SetString(params["toPub"], 16)
	copy(toPub[:], toPubInt.Bytes())

	fromPubInt1 := new(big.Int)
	fromPubInt1.SetBytes(fromPub[:32])
	fromPubInt2 := new(big.Int)
	fromPubInt2.SetBytes(fromPub[32:])

	pubKey := ecdsa.PublicKey{
		elliptic.P256(),
		fromPubInt1,
		fromPubInt2,
	}

	fromPrivInt, _ := new(big.Int).SetString(params["fromPriv"], 16)

	fromPriv := ecdsa.PrivateKey{
		pubKey,
		fromPrivInt,
	}

	tx, _ := protocol.ConstrFundsTx(
		byte(0),
		uint64(amount),
		uint64(fee),
		uint32(txCnt),
		client.SerializeHashContent(fromPub[:]),
		client.SerializeHashContent(toPub[:]),
		&fromPriv,
	)

	client.SendTx(tx, p2p.FUNDSTX_BRDCST)
}
