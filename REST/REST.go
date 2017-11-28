package REST

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func Init() {
	router := mux.NewRouter()
	getEndpoints(router)
	//router.HandleFunc("/people/{id}", GetPersonEndpoint).Methods("GET")
	//router.HandleFunc("/people/{id}", CreatePersonEndpoint).Methods("POST")
	//router.HandleFunc("/people/{id}", DeletePersonEndpoint).Methods("DELETE")
	log.Fatal(http.ListenAndServe(":8001", router))
}

func getEndpoints(router *mux.Router) {
	router.HandleFunc("/account/{id}", GetAccountEndpoint).Methods("GET")
}
