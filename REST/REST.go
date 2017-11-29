package REST

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

var (
	logger *log.Logger
)

func Init() {
	logger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

	router := mux.NewRouter()

	getEndpoints(router)

	log.Fatal(http.ListenAndServe(":8001", router))
}

func getEndpoints(router *mux.Router) {
	router.HandleFunc("/account/{id}", GetAccountEndpoint).Methods("GET")
	router.HandleFunc("/createFundsTx/{amount}/{fee}/{txCnt}/{fromPub}/{toPub}/{fromPriv}", CreateFundsTxEndpoint).Methods("POST")
	router.HandleFunc("/sendFundsTx/{signedTx}", SendFundsTxEndpoint).Methods("POST")
}
