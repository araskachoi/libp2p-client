package main

import (
	// "fmt"
	"context"
)

func main() {
	// The context governs the lifetime of the libp2p node
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// listenPort := 9999
	seed := int64(0)

	getOptions(ctx)

	
}