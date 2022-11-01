package main

import (
	"github.com/atburke/krpc-go/api"
	"github.com/atburke/krpc-go/lib/encode"
)

func main() {
	t := api.NewTuple3("test", 1, 1.1)
	encode.Marshal(t)
}
