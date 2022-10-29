package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/atburke/krpc-go/internal"
	"github.com/atburke/krpc-go/lib/client"
)

func main() {
	client := client.NewKRPCClient(client.KRPCClientConfig{})
	defer client.Close()

	err := client.Connect(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting: %v\n", err)
		os.Exit(1)
	}

	krpc := internal.NewBasicKRPC(client)

	status, err := krpc.GetStatus()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting status: %v\n", err)
		os.Exit(1)
	}
	out, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshalling json: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(out))
	services, err := krpc.GetServices()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting services: %v\n", err)
		os.Exit(1)
	}
	out, err = json.MarshalIndent(services, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshalling json: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(out))
}
