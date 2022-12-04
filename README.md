# krpc-go
krpc-go is a Go client for [kRPC](https://github.com/krpc/krpc), a [Kerbal Space Program](https://www.kerbalspaceprogram.com/) mod for controlling the game with an external program.

## Installation

```sh
go get github.com/atburke/krpc-go
```

## Getting Started

This sample program will launch a vessel sitting on the launchpad. Error handling is omitted for brevity.

```go
package main

import (
    "context"

    krpcgo "github.com/atburke/krpc-go"
    "github.com/atburke/krpc-go/krpc"
    "github.com/atburke/krpc-go/spacecenter"
)

func main() {
    // Connect to the kRPC server with all default parameters.
    client := krpcgo.DefaultKRPCClient()
    client.Connect(context.Background())
    defer client.Close()

    sc := spacecenter.New(client)
    vessel, _ := sc.ActiveVessel()
    control, _ := vessel.Control()

    control.SetSAS(true)
    control.SetRCS(false)
    control.SetThrottle(1.0)
    control.ActivateNextStage()
}
```

### Types

This section describes type mappings from the kRPC protocol.

- Primitives are mapped to Go primitives.
- Arrays are mapped to slices. Dictionaries and sets are mapped to maps.
- Tuples are mapped to a special tuple type in the `types` package. For example, a tuple of strings would map to `types.Tuple3[string, string, string]`.
  - `types` also contains some convenience types that can be converted to/from the appropriate tuple, such as `types.Vector2D`, `types.Vector3D`, and `types.Color`.
- Classes and enums are mapped to local structs and constants defined in the appropriate service. For example, a Vessel will be mapped to a `*spacecenter.Vessel`, and a GameScene will be mapped to a `krpc.GameScene`.
- Existing protobuf types can be found in the `types` package. For example, a Status will be mapped to a `*types.Status`.

### Streams

krpc-go uses Go's built-in channels to handle streams. 

Here's an example of using streams to autostage a vessel until a specific stage is reached.

```go
func AutoStageUntil(vessel *spacecenter.Vessel, stopStage int32) {
    go func() {
        control, _ := vessel.Control()
        stage, _ := control.CurrentStage()

        for stage > stopStage {
            resources, _ := vessel.ResourcesInDecoupleStage(stage-1, false)
            amountStream, _ := resources.AmountStream("LiquidFuel")

            // Wait until this stage runs out of liquid fuel.
            for amount := <-amountStream.C; amount > 0.1 {}

            amountStream.Close()
            control.ActivateNextStage()
            stage--
        }
    }()
}
```

### More examples

See tests in `integration/` for more usage examples.

## Building

TODO

## Links

TODO krpc-go docs link
[kRPC documentation](https://krpc.github.io/krpc/index.html)
