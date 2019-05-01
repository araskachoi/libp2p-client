package main

import (
	"context"
	"errors"
	"math/rand"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	inet "github.com/libp2p/go-libp2p-net"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	host "github.com/libp2p/go-libp2p-host"
)

var BootstrapPeers = dht.DefaultBootstrapPeers

const BootstrapConnections = 4

func bootstrapPeerInfo() ([]*pstore.PeerInfo, error) {
	pis := make([]*pstore.PeerInfo, 0, len(BootstrapPeers))
	for _, a := range BootstrapPeers {
		pi, err := pstore.InfoFromP2pAddr(a)
		if err != nil {
			return nil, err
		}
		pis = append(pis, pi)
	}
	return pis, nil
}

func shufflePeerInfos(peers []*pstore.PeerInfo) {
	for i := range peers {
		j := rand.Intn(i + 1)
		peers[i], peers[j] = peers[j], peers[i]
	}
}

func Bootstrap(host host.Host) error {
	pis, err := bootstrapPeerInfo()
	if err != nil {
		return err
	}

	for _, pi := range pis {
		host.Peerstore().AddAddrs(pi.ID, pi.Addrs, pstore.PermanentAddrTTL)
	}

	count := connectBootstrapPeers(pis, BootstrapConnections)
	if count == 0 {
		return errors.New("Failed to connect to bootstrap peers")
	}

	go keepBootstrapConnections(pis)

	if dht != nil {
		return dht.Bootstrap(ctx)
	}

	return nil
}

func connectBootstrapPeers(pis []*pstore.PeerInfo, toconnect int) int {
	count := 0

	shufflePeerInfos(pis)

	ctx, cancel := context.WithTimeout(d.ctx, 60*time.Second)
	defer cancel()

	for _, pi := range pis {
		if d.host.Network().Connectedness(pi.ID) == inet.Connected {
			continue
		}
		err := d.host.Connect(ctx, *pi)
		if err != nil {
			log.Debugf("Error connecting to bootstrap peer %s: %s", pi.ID, err.Error())
		} else {
			d.host.ConnManager().TagPeer(pi.ID, "bootstrap", 1)
			count++
			toconnect--
		}
		if toconnect == 0 {
			break
		}
	}

	return count

}

func (d *Daemon) keepBootstrapConnections(pis []*pstore.PeerInfo) {
	ticker := time.NewTicker(15 * time.Minute)
	for {
		<-ticker.C

		conns := d.host.Network().Conns()
		if len(conns) >= BootstrapConnections {
			continue
		}

		toconnect := BootstrapConnections - len(conns)
		d.connectBootstrapPeers(pis, toconnect)
	}
}
