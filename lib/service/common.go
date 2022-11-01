package service

import "github.com/atburke/krpc-go/lib/client"

type Class interface {
	// ID gets the instance's ID.
	ID() uint64
	// SetID sets the instance's ID.
	SetID(uint64)
}

// BaseClass is the base for all classes.
type BaseClass struct {
	// ID is the struct's id.
	id uint64
	// Client is a kRPC client.
	Client *client.KRPCClient
}

// ID gets the instance's ID.
func (c *BaseClass) ID() uint64 {
	return c.id
}

// SetID sets the instance's ID.
func (c *BaseClass) SetID(id uint64) {
	c.id = id
}
