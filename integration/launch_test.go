package integration

import (
	"context"
	"fmt"
	"math"
	"testing"

	krpcgo "github.com/atburke/krpc-go"
	"github.com/atburke/krpc-go/krpc"
	"github.com/atburke/krpc-go/spacecenter"
	"github.com/stretchr/testify/require"
)

// TestLaunch starts from the space center, loads the Kerbal, X, and launches
// it into orbit.
func TestLaunch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	client := krpcgo.NewKRPCClient(krpcgo.KRPCClientConfig{})
	require.NoError(t, client.Connect(ctx))

	krpcService := krpc.NewKRPC(client)
	// require.NoError(t, krpcService.SetPaused(false))
	t.Cleanup(func() {
		require.NoError(t, krpcService.SetPaused(true))
	})

	gamescene, err := krpcService.CurrentGameScene()
	require.NoError(t, err)
	fmt.Println("got game scene")
	require.Equal(t, krpc.GameScene_Flight, gamescene, "Test should be run from the launch pad.")
	sc := spacecenter.NewSpaceCenter(client)

	vessel, err := sc.ActiveVessel()
	require.NoError(t, err)
	fmt.Println("got vessel")
	name, err := vessel.Name()
	require.NoError(t, err)
	fmt.Printf("Launching vessel %q\n", name)

	rf, err := vessel.SurfaceReferenceFrame()
	require.NoError(t, err)
	fmt.Println("got reference frame")
	flight, err := vessel.Flight(rf)
	require.NoError(t, err)
	fmt.Println("got flight")
	orbit, err := vessel.Orbit()
	require.NoError(t, err)
	fmt.Println("got orbit")

	altitudeStream, err := flight.StreamMeanAltitude()
	require.NoError(t, err)
	apoapsisStream, err := orbit.StreamApoapsisAltitude()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, altitudeStream.Close())
		require.NoError(t, apoapsisStream.Close())
	})

	control, err := vessel.Control()
	require.NoError(t, err)
	require.NoError(t, control.SetSAS(false))
	require.NoError(t, control.SetRCS(false))
	require.NoError(t, control.SetThrottle(1.0))

	autopilot, err := vessel.AutoPilot()
	require.NoError(t, err)

	_, err = control.ActivateNextStage()
	require.NoError(t, err)
	require.NoError(t, autopilot.Engage())
	require.NoError(t, autopilot.TargetPitchAndHeading(90.0, 90))

	// Autostaging
	go func() {
		stage, err := control.CurrentStage()
		require.NoError(t, err)

		for {
			fmt.Printf("current stage is %v\n", stage)
			resources, err := vessel.ResourcesInDecoupleStage(stage-1, false)
			require.NoError(t, err)
			names, err := resources.Names()
			require.NoError(t, err)
			fmt.Println(names)
			amountStream, err := resources.StreamAmount("LiquidFuel")
			require.NoError(t, err)

		readAmount:
			for {
				select {
				case amount := <-amountStream.C:
					if amount < 0.1 {
						fmt.Println("ran out of liquid fuel in current stage")
						_, err = control.ActivateNextStage()
						require.NoError(t, amountStream.Close())
						require.NoError(t, err)
						stage--
						if stage == 0 {
							return
						}
						break readAmount
					}
				case <-ctx.Done():
					return
				}
			}

		}
	}()

	turnStartAltitude := 250.0
	turnEndAltitude := 45000.0
	targetAltitude := 150000.0

	turnAngle := 0.0
	var apoapsis float64

	for apoapsis < 0.9*targetAltitude {
		select {
		case altitude := <-altitudeStream.C:
			if altitude < turnStartAltitude || altitude > turnEndAltitude {
				continue
			}
			frac := (altitude - turnStartAltitude) / (turnEndAltitude - turnStartAltitude)
			newTurnAngle := frac * 90
			if math.Abs(newTurnAngle-turnAngle) > 0.5 {
				turnAngle = newTurnAngle
				require.NoError(t, autopilot.TargetPitchAndHeading(float32(90-turnAngle), 90))
			}
		case apoapsis = <-apoapsisStream.C:
		case <-ctx.Done():
			return
		}
	}
}
