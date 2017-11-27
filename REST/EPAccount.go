package REST

import (
	"net/http"
	"encoding/hex"
	"encoding/json"
	"github.com/mchetelat/bazo_client/client"
	"github.com/gorilla/mux"
	"fmt"
)

type FormattedAcc struct {
	Address string
	TxCnt uint32
	Balance uint64
}

func GetAccountEndpoint(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	fmt.Println(params["id"])

	var formattedAcc FormattedAcc

	formattedAcc.Address = hex.EncodeToString(client.Acc.Address[:])
	formattedAcc.TxCnt = client.Acc.TxCnt
	formattedAcc.Balance = client.Acc.Balance

	json.NewEncoder(w).Encode(formattedAcc)
}