package main

import (
	"fmt"
	"github.com/mchetelat/bazo_client/client"
	"os"
)

func main() {

	if len(os.Args) >= 2 {
		if os.Args[1] == "accTx" || os.Args[1] == "fundsTx" || os.Args[1] == "configTx" {
			client.Process(os.Args[1:])
		} else {
			client.Init(os.Args[1])
		}
	} else {
		fmt.Printf("%v\n", client.USAGE_MSG)
	}
}
