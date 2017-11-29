package REST

import (
	"encoding/hex"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/mchetelat/bazo_client/client"
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

	amount, _ := strconv.Atoi(params["amount"])
	fee, _ := strconv.Atoi(params["fee"])
	txCnt, _ := strconv.Atoi(params["txCnt"])

	fromPubInt, _ := new(big.Int).SetString(params["fromPub"], 16)
	copy(fromPub[:], fromPubInt.Bytes())

	toPubInt, _ := new(big.Int).SetString(params["toPub"], 16)
	copy(toPub[:], toPubInt.Bytes())

	tx := protocol.FundsTx{
		Header: 0,
		Amount: uint64(amount),
		Fee:    uint64(fee),
		TxCnt:  uint32(txCnt),
		From:   client.SerializeHashContent(fromPub),
		To:     client.SerializeHashContent(toPub),
	}

	txHash := tx.Hash()

	js, err := json.Marshal(hex.EncodeToString(txHash[:]))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func SendFundsTxEndpoint(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	signedTx := params["signedTx"]
}
