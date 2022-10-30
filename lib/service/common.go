package service

import "github.com/atburke/krpc-go/lib/client"

// BaseClass is the base for all classes.
type BaseClass struct {
	// ID is the struct's id.
	ID uint64
	// Client is a kRPC client.
	Client *client.KRPCClient
}
