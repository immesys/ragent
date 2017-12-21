package ragentlib

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/immesys/bw2/crypto"
	"github.com/immesys/bw2/objects"
)

var ourEntity *objects.Entity

func DoClient(relayEntityFile string, remote string, remotevk string, listenaddr string) {
	econtents, err := ioutil.ReadFile(relayEntityFile)
	if err != nil {
		panic(err)
	}
	DoClientER(econtents, remote, remotevk, listenaddr)
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
func DoClientER(econtents []byte, remote string, remotevk string, listenaddr string) {
	enti, err := objects.NewEntity(objects.ROEntityWKey, econtents[1:])
	errc := color.New(color.FgRed, color.Bold)
	if err != nil {
		errc.Printf("invalid auth: %s\n", err)
		os.Exit(1)
	}
	ourEntity = enti.(*objects.Entity)
	ln, err := net.Listen("tcp", listenaddr)
	if err != nil {
		errc.Printf("could not connect to init ragent client: %s\n", err)
		os.Exit(1)
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
	fmt.Printf("trying connection to %s\n", remote)
	roots := x509.NewCertPool()
	errc := color.New(color.FgRed, color.Bold)
	conn, err := tls.Dial("tcp", remote, &tls.Config{
		InsecureSkipVerify: true,
		RootCAs:            roots,
	})
	if err != nil {
		errc.Printf("bosswave connection error: %s\n", err)
		os.Exit(1)
	}
	expectedVK, err := crypto.UnFmtKey(remotevk)
	if err != nil {
		errc.Printf("handshake error error : %s\n", err)
		os.Exit(1)
	}
	cs := conn.ConnectionState()
	if len(cs.PeerCertificates) != 1 {
		panic("peer gave no certs")
	}
	proof := make([]byte, 96)
	fmt.Printf("starting to wait for proof\n")
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
	ctx, cancel := context.WithCancel(context.Background())
	go copysimplex("remote->local", conn, lconn, cancel)
	go copysimplex("local->remote", lconn, conn, cancel)
	<-ctx.Done()
	fmt.Println("relay terminated")
	lconn.Close()
	conn.Close()
}

func copysimplex(desc string, a, b net.Conn, cancel func()) {
	total := 0
	last := time.Now()
	buf := make([]byte, 4096)

	for {
		count, err := a.Read(buf)
		if count == 0 || err != nil {
			fmt.Printf("read error: %v\n", err)
			cancel()
			return
		}
		b.SetWriteDeadline(time.Now().Add(2 * time.Minute))
		_, err = b.Write(buf[:count])
		if err != nil {
			fmt.Printf("write error: %v\n", err)
			cancel()
			return
		}
		total += count
		if time.Now().Sub(last) > 15*time.Second {
			fmt.Println(desc, total, "bytes transferred")
			last = time.Now()
		}
	}
}
