package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"net"

	"github.com/immesys/bw2/crypto"
	"github.com/immesys/bw2/objects"
)

func doclient(relayEntityFile string, remote string, remotevk string, listenaddr string) {
	econtents, err := ioutil.ReadFile(relayEntityFile)
	if err != nil {
		panic(err)
	}
	enti, err := objects.NewEntity(objects.ROEntityWKey, econtents[1:])
	if err != nil {
		panic(err)
	}
	ourEntity = enti.(*objects.Entity)
	ln, err := net.Listen("tcp", listenaddr)
	if err != nil {
		panic(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
		}
		go proxyclient(conn, remote, remotevk)
	}
}
func proxyclient(lconn net.Conn, remote, remotevk string) {
	roots := x509.NewCertPool()
	conn, err := tls.Dial("tcp", remote, &tls.Config{
		InsecureSkipVerify: true,
		RootCAs:            roots,
	})
	if err != nil {
		panic(err)
	}
	expectedVK, err := crypto.UnFmtKey(remotevk)
	if err != nil {
		panic(err)
	}
	cs := conn.ConnectionState()
	if len(cs.PeerCertificates) != 1 {
		panic("peer gave no certs")
	}
	proof := make([]byte, 96)
	_, err = io.ReadFull(conn, proof)
	if err != nil {
		panic("failed to read proof: " + err.Error())
	}
	proofOK := crypto.VerifyBlob(proof[:32], proof[32:], cs.PeerCertificates[0].Raw)
	if !proofOK {
		panic("peer verification failed")
	}
	fmt.Println("remote proof valid for: ", crypto.FmtKey(proof[:32]))
	if !bytes.Equal(proof[:32], expectedVK) {
		panic("peer has a different VK")
	}
	//Now we read a nonce from them
	nonce := make([]byte, 32)
	_, err = io.ReadFull(conn, nonce)
	if err != nil {
		panic("failed to read nonce: " + err.Error())
	}
	//And sign it
	copy(proof[:32], ourEntity.GetVK())
	crypto.SignBlob(ourEntity.GetSK(), ourEntity.GetVK(), proof[32:], nonce)
	conn.Write(proof)
	//Now read back the state from them:
	state := make([]byte, 4)
	_, err = io.ReadFull(conn, state)
	if err != nil {
		panic("failed to read state: " + err.Error())
	}
	if string(state) != "OKAY" {
		panic(state)
	}
	fmt.Println("remote sent OKAY, beginning relay")
	go copysimplex("remote->local", conn, lconn)
	copysimplex("local->remote", lconn, conn)
	fmt.Println("relay terminated")
	lconn.Close()
	conn.Close()
}
