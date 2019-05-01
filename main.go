package main

import (
	"io"
	"context"
	"flag"
	"fmt"
	"log"
	"crypto/rand"
	"strings"
	mrand "math/rand"

	libp2p "github.com/libp2p/go-libp2p"
	relay "github.com/libp2p/go-libp2p-circuit"
	crypto "github.com/libp2p/go-libp2p-crypto"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	ps "github.com/libp2p/go-libp2p-pubsub"
	identify "github.com/libp2p/go-libp2p/p2p/protocol/identify"
	multiaddr "github.com/multiformats/go-multiaddr"

	_ "net/http/pprof"
)
/*
func main() {
	// The context governs the lifetime of the libp2p node
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	o, n, err := getOptions(ctx)
	fmt.Println(o, n , err)

	// listenPort := 9999
	// seed := int64(0)

}
*/

func main() {
	identify.ClientVersion = "p2pd/0.1"

	seed := flag.Int64("seed", 0, "seed to generate privKey")
	id := flag.String("id", "", "peer identity; private key file")
	peers := flag.String("staticPeers", "", "comma separated list of peers; defaults to no peers")
	connMgr := flag.Bool("connManager", false, "Enables the Connection Manager")
	connMgrLo := flag.Int("connLo", 256, "Connection Manager Low Water mark")
	connMgrHi := flag.Int("connHi", 512, "Connection Manager High Water mark")
	connMgrGrace := flag.Duration("connGrace", 120, "Connection Manager grace period (in seconds)")
	natPortMap := flag.Bool("natPortMap", false, "Enables NAT port mapping")
	pubsub := flag.Bool("pubsub", false, "Enables pubsub")
	pubsubRouter := flag.String("pubsubRouter", "gossipsub", "Specifies the pubsub router implementation")
	pubsubSign := flag.Bool("pubsubSign", true, "Enables pubsub message signing")
	pubsubSignStrict := flag.Bool("pubsubSignStrict", false, "Enables pubsub strict signature verification")
	gossipsubHeartbeatInterval := flag.Duration("gossipsubHeartbeatInterval", 0, "Specifies the gossipsub heartbeat interval")
	gossipsubHeartbeatInitialDelay := flag.Duration("gossipsubHeartbeatInitialDelay", 0, "Specifies the gossipsub initial heartbeat delay")
	relayEnabled := flag.Bool("relay", true, "Enables circuit relay")
	relayActive := flag.Bool("relayActive", false, "Enables active mode for relay")
	relayHop := flag.Bool("relayHop", false, "Enables hop for relay")
	hostAddrs := flag.String("hostAddrs", "", "comma separated list of multiaddrs the host should listen on")
	announceAddrs := flag.String("announceAddrs", "", "comma separated list of multiaddrs the host should announce to the network")
	noListen := flag.Bool("noListenAddrs", false, "sets the host to listen on no addresses")

	flag.Parse()

	var opts []libp2p.Option
	var staticPeers []multiaddr.Multiaddr

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	if *id != "" {
		var r io.Reader
		if *seed == int64(0) {
			r = rand.Reader
		} else {
			r = mrand.New(mrand.NewSource(*seed))
		}
		priv, _, err := crypto.GenerateEd25519Key(r)
		if err != nil {
			panic(err)
		}

		opts = append(opts, libp2p.Identity(priv))
	}

	if *hostAddrs != "" {
		addrs := strings.Split(*hostAddrs, ",")
		opts = append(opts, libp2p.ListenAddrStrings(addrs...))
	}

	if *announceAddrs != "" {
		addrs := strings.Split(*announceAddrs, ",")
		maddrs := make([]multiaddr.Multiaddr, 0, len(addrs))
		for _, a := range addrs {
			maddr, err := multiaddr.NewMultiaddr(a)
			if err != nil {
				log.Fatal(err)
			}
			maddrs = append(maddrs, maddr)
		}
		opts = append(opts, libp2p.AddrsFactory(func([]multiaddr.Multiaddr) []multiaddr.Multiaddr {
			return maddrs
		}))
	}

	if *connMgr {
		cm := connmgr.NewConnManager(*connMgrLo, *connMgrHi, *connMgrGrace)
		opts = append(opts, libp2p.ConnectionManager(cm))
	}

	if *natPortMap {
		opts = append(opts, libp2p.NATPortMap())
	}

	if *relayEnabled {
		var relayOpts []relay.RelayOpt
		if *relayActive {
			relayOpts = append(relayOpts, relay.OptActive)
		}
		if *relayHop {
			relayOpts = append(relayOpts, relay.OptHop)
		}
		// if *relayDiscovery {
		// 	relayOpts = append(relayOpts, relay.OptDiscovery)
		// }
		opts = append(opts, libp2p.EnableRelay(relayOpts...))
	}

	if *noListen {
		opts = append(opts, libp2p.NoListenAddrs)
	}

	host, err := generateHost(ctx, opts)
	if err != nil {
		panic(err)
	}

	if *pubsub {
		var psOpts []ps.Option
		if *gossipsubHeartbeatInterval > 0 {
			ps.GossipSubHeartbeatInterval = *gossipsubHeartbeatInterval
		}

		if *gossipsubHeartbeatInitialDelay > 0 {
			ps.GossipSubHeartbeatInitialDelay = *gossipsubHeartbeatInitialDelay
		}

		if *pubsubSign {
			psOpts = append(psOpts, ps.WithMessageSigning(*pubsubSign))

			if *pubsubSignStrict {
				psOpts = append(psOpts, ps.WithStrictSignatureVerification(*pubsubSignStrict))
			}
		}

		switch *pubsubRouter {
		case "floodsub":
			_, err := ps.NewFloodSub(ctx, host, psOpts...)
			if err != nil {
				panic(err)
			}
	
		case "gossipsub":
			_, err := ps.NewGossipSub(ctx, host, psOpts...)
			if err != nil {
				panic(err)
			}
	
		default:
			fmt.Errorf("unknown pubsub router: %s", *pubsubRouter)
			return
		}
	}

	if *peers != "" {
		for _, s := range strings.Split(*peers, ",") {
			ma, err := multiaddr.NewMultiaddr(s)
			if err != nil {
				log.Fatalf("error parsing bootstrap peer %q: %v", s, err)
			}
			staticPeers = append(staticPeers, ma)
		}
	}

	fmt.Println(host)
}