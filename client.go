package main

import (
	"context"

	libp2p "github.com/libp2p/go-libp2p"
	host "github.com/libp2p/go-libp2p-host"
)



//generates libp2p host
func generateHost(ctx context.Context, opts []libp2p.Option) (host.Host, error) {
	host, err := libp2p.New(ctx, opts...,)
	if err != nil {
		panic(err)
	}
	return host, err 
}




