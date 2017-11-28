package REST

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/mchetelat/bazo_client/client"
	"math/big"
	"net/http"
)

func GetAccountEndpoint(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)

	var pubKey [64]byte
	pubKeyInt, _ := new(big.Int).SetString(params["id"], 16)
	copy(pubKey[:], pubKeyInt.Bytes())

	acc, err := client.GetAccount(pubKey)
	if err != nil {
		json.NewEncoder(w).Encode(err.Error())
	} else {
		json.NewEncoder(w).Encode(acc.String())
	}
}
