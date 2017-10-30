package main

import (
	"fmt"
	"github.com/mchetelat/bazo_client/client"
	"os"
)

func main() {

	switch len(os.Args) {
	case 2:
		if os.Args[1] == "accTx" || os.Args[1] == "fundsTx" || os.Args[1] == "configTx" {
			client.Process(os.Args[1:])
		} else {
			client.Init(os.Args[1])
		}
	default:
		fmt.Printf("Usage: bazo_client [pubKey|accTx|fundsTx|configTx] ...\n")
	}
}
