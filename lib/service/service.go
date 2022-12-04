// Package service provides some definitions needed to generate services.
package service

import krpcgo "github.com/atburke/krpc-go"

type Enum interface {
	Value() int32
}

type SettableEnum interface {
	SetValue(int32)
}

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
	Client *krpcgo.KRPCClient
}

// ID gets the instance's ID.
func (c *BaseClass) ID() uint64 {
	return c.id
}

// SetID sets the instance's ID.
func (c *BaseClass) SetID(id uint64) {
	c.id = id
}
