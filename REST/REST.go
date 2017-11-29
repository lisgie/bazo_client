package REST

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func Init() {
	router := mux.NewRouter()

	getEndpoints(router)

	log.Fatal(http.ListenAndServe(":8001", router))
}

func getEndpoints(router *mux.Router) {
	router.HandleFunc("/account/{id}", GetAccountEndpoint).Methods("GET")
	router.HandleFunc("/fundsTx/{amount}/{fee}/{txCnt}/{fromPub}/{toPub}/{fromPriv}", CreateFundsTxEndpoint).Methods("GET")
}
