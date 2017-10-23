package main

import (
	"fmt"
	"github.com/mchetelat/bazo_client/client"
	"os"
)

func main() {

	switch len(os.Args) {
	case 1:
		fmt.Println(fmt.Printf("Usage: bazo_client [pubKey|accTx|fundsTx|configTx] ...\n"))
	case 2:
		client.Init(os.Args[1])
	default:
		client.Process(os.Args[1:])
	}
}
