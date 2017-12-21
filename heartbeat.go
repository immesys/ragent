package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/immesys/bw2bind"
)

func beat(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	conn, err := net.DialTimeout("tcp", "127.0.0.1:28590", 2*time.Second)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("ragent conn failed"))
		return
	}
	conn.Close()

	cl, err := bw2bind.Connect("127.0.0.1:28589")
	if err != nil {
		fmt.Printf("Could not connect to agent: %s\n", err)
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("could not connect to agent: %s\n", err)))
		return
	}
	defer cl.Close()
	cip, err := cl.GetBCInteractionParams()
	if err != nil {
		fmt.Printf("Could not get BCIP: %s\n", err)
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("Could not get BCIP: %s\n", err)))
		return
	}

	if cip.Peers == 0 {
		fmt.Printf("Peers is zero\n")
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("Peers is zero")))
		return
	}

	if cip.CurrentAge > 10*time.Minute {
		fmt.Printf("Chain age is %s\n", cip.CurrentAge)
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("Chain age is %s\n", cip.CurrentAge)))
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("OK"))
}
func doheartbeat() {
	http.Handle("/healthz", http.HandlerFunc(beat))
	log.Fatal(http.ListenAndServe(":28591", nil))
}
