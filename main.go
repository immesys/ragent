package main

import (
	"fmt"
	"os"

	"github.com/immesys/ragent/ragentlib"
)

/*
 overview:
 client connects over TLS.
 we do typical proof that the ragent is a certain entity.
 client gives a dot hash from the ragent to the client on ragent/v1.0/full
 client will now forward all traffic to the local agent
*/
func main() {
	usage := func() {
		fmt.Println(`BOSSWAVE Remote Agent Relay`)
		fmt.Println(` server mode (accept TLS and relay to OOB):`)
		fmt.Println(`  ragent server <entityfile> <listenaddr> <agentaddr>`)
		fmt.Println(` client mode (accept OOB and relay over TLS):`)
		fmt.Println(`  ragent client <entityfile> <serveraddr> <servervk> <listenaddr>`)
		os.Exit(1)
	}
	if len(os.Args) < 5 {
		usage()
	}
	if os.Args[1] == "server" {
		if len(os.Args) != 5 {
			usage()
		}
		doserver(os.Args[2], os.Args[3], os.Args[4])
	} else if os.Args[1] == "client" {
		if len(os.Args) != 6 {
			usage()
		}
		ragentlib.DoClient(os.Args[2], os.Args[3], os.Args[4], os.Args[5])
	} else {
		usage()
	}

}
