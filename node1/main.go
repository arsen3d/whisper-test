package main

import (
	"crypto/ecdsa"
	"fmt"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/whisper/whisperv6"
)

func exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// Node2Encode ...
const Node2Encode = "50095ad7bd27b0c99e673a90b1f818b408ded5672ac578cf80799333be371c190c02a1c16b92bdd0894c5757c5c4987afe9422dbb22af8b00b41943db066add0"

func main() {
	var priKey *ecdsa.PrivateKey
	keyFile := "node1.key"
	if exists(keyFile) {
		priKey, _ = crypto.LoadECDSA(keyFile)
	} else {
		priKey, _ = crypto.GenerateKey()
		crypto.SaveECDSA(keyFile, priKey)
	}

	// set the log level to Trace
	log.Root().SetHandler(log.LvlFilterHandler(log.LvlTrace, log.StreamHandler(os.Stdout, log.TerminalFormat(true))))

	whisperv6Config := whisperv6.DefaultConfig
	whisperv6Config.MinimumAcceptedPOW = 0
	whisper := whisperv6.New(&whisperv6Config)

	p2pConfig := &p2p.Config{
		PrivateKey: priKey,
		MaxPeers:   10,
		ListenAddr: ":8000",
		Protocols:  whisper.Protocols(),
		Logger:     log.Root(),
	}
	srv := p2p.Server{Config: *p2pConfig}
	if err := srv.Start(); err != nil {
		fmt.Println("could not start server:", err)
		os.Exit(1)
	}

	log.Info("Node", "info", srv.NodeInfo())

	filter := whisperv6.Filter{
		AllowP2P: true,
		PoW:      0,
		KeySym:   []byte("whisperv6 message test.........."),
	}
	whisper.Subscribe(&filter)

	whisper.Start(&srv)

	go func() {
		for {
			time.Sleep(time.Second)
			for _, peer := range srv.Peers() {
				srv.Logger.Info("print peer info", "id", peer.ID(), "name", peer.String())

				n, err := discover.ParseNode(Node2Encode)
				if err != nil {
					srv.Logger.Error("ParseNode failed")
				}

				err = whisper.AllowP2PMessagesFromPeer(n.ID[:])
				if err != nil {
					srv.Logger.Error("AllowP2PMessagesFromPeer failed")
				}
				return
			}
		}
	}()

	go func() {
		for {
			time.Sleep(time.Second)
			srv.Logger.Info("filter.Retrieve")
			for _, msg := range filter.Retrieve() {
				srv.Logger.Info("recvd msg", "payload", string(msg.Payload[:]))
			}
		}
	}()

	select {}
}
