package integration

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	krpcgo "github.com/atburke/krpc-go"
	"github.com/atburke/krpc-go/krpc"
	"github.com/atburke/krpc-go/lib/api"
	"github.com/atburke/krpc-go/spacecenter"
	"github.com/stretchr/testify/require"
)

// TestLaunch starts from the space center, loads the Kerbal, X, and launches
// it into orbit. The procedure for launching the vessel into orbit is adapted
// from https://krpc.github.io/krpc/tutorials/launch-into-orbit.html.
func TestLaunch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	client := krpcgo.NewKRPCClient(krpcgo.KRPCClientConfig{})
	require.NoError(t, client.Connect(ctx))

	krpcService := krpc.NewKRPC(client)
	require.NoError(t, krpcService.SetPaused(false))
	t.Cleanup(func() {
		require.NoError(t, krpcService.SetPaused(true))
	})

	// Set stuff up
	gamescene, err := krpcService.CurrentGameScene()
	require.NoError(t, err)
	require.Equal(t, krpc.GameScene_Flight, gamescene, "Test should be run from the launch pad.")
	sc := spacecenter.NewSpaceCenter(client)

	vessel, err := sc.ActiveVessel()
	require.NoError(t, err)

	rf, err := vessel.SurfaceReferenceFrame()
	require.NoError(t, err)
	flight, err := vessel.Flight(rf)
	require.NoError(t, err)
	orbit, err := vessel.Orbit()
	require.NoError(t, err)

	altitudeStream, err := flight.MeanAltitudeStream()
	require.NoError(t, err)
	apoapsisStream, err := orbit.ApoapsisAltitudeStream()
	require.NoError(t, err)
	qStream, err := flight.DynamicPressureStream()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, altitudeStream.Close())
		require.NoError(t, apoapsisStream.Close())
		require.NoError(t, qStream.Close())
	})

	control, err := vessel.Control()
	require.NoError(t, err)
	require.NoError(t, control.SetSAS(false))
	require.NoError(t, control.SetRCS(false))
	require.NoError(t, control.SetThrottle(1.0))

	autopilot, err := vessel.AutoPilot()
	require.NoError(t, err)

	// Launch
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
			amountStream, err := resources.AmountStream("LiquidFuel")
			require.NoError(t, err)

		readAmount:
			for {
				select {
				case amount := <-amountStream.C:
					if amount < 0.1 {
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

	limitingThrottle := false

	for apoapsis < 0.9*targetAltitude {
		select {
		// Manage heading
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

			// Lazy Q limiting
		case q := <-qStream.C:
			if q >= 20000 && !limitingThrottle {
				limitingThrottle = true
				require.NoError(t, control.SetThrottle(0.5))
			} else if q < 20000 && limitingThrottle {
				limitingThrottle = false
				require.NoError(t, control.SetThrottle(1.0))
			}
		case <-ctx.Done():
			return
		}
	}

	// Fine tune apoapsis approach
	require.NoError(t, control.SetThrottle(0.25))
	for apoapsis < targetAltitude {
		select {
		case apoapsis = <-apoapsisStream.C:
		case <-ctx.Done():
			return
		}
	}
	require.NoError(t, control.SetThrottle(0))

	// Coast out of the atmosphere
	for apoapsis < 70500 {
		select {
		case apoapsis = <-apoapsisStream.C:
		case <-ctx.Done():
			return
		}
	}

	// Plan circularization
	body, err := orbit.Body()
	require.NoError(t, err)
	mu, err := body.GravitationalParameter()
	require.NoError(t, err)
	r, err := orbit.Apoapsis()
	require.NoError(t, err)
	a1, err := orbit.SemiMajorAxis()
	require.NoError(t, err)
	a2 := r
	v1 := math.Sqrt(float64(mu) * ((2 / r) - (1 / a1)))
	v2 := math.Sqrt(float64(mu) * ((2 / r) - (1 / a2)))
	deltaV := v2 - v1
	ut, err := sc.UT()
	require.NoError(t, err)
	timeToApoapsis, err := orbit.TimeToApoapsis()
	require.NoError(t, err)
	node, err := control.AddNode(ut+timeToApoapsis, float32(deltaV), 0, 0)
	require.NoError(t, err)

	// Calculate burn time
	f, err := vessel.AvailableThrust()
	require.NoError(t, err)
	rawISP, err := vessel.SpecificImpulse()
	require.NoError(t, err)
	isp := float64(rawISP * 9.82)
	m0, err := vessel.Mass()
	require.NoError(t, err)
	m1 := float64(m0) / math.Exp(deltaV/isp)
	flowRate := float64(f) / isp
	burnTime := (float64(m0) - m1) / flowRate

	// Orient ship
	require.NoError(t, control.SetRCS(true))
	nodeRF, err := node.ReferenceFrame()
	require.NoError(t, err)
	require.NoError(t, autopilot.SetReferenceFrame(nodeRF))
	require.NoError(t, autopilot.SetTargetDirection(api.NewVector3D(0, 1, 0).Tuple()))
	require.NoError(t, autopilot.Wait())

	// Wait until burn
	ut, err = sc.UT()
	require.NoError(t, err)
	timeToApoapsis, err = orbit.TimeToApoapsis()
	require.NoError(t, err)
	burnUT := ut + timeToApoapsis - (burnTime / 2)
	leadTime := float64(5)
	require.NoError(t, sc.WarpTo(burnUT-leadTime, 10, 1))

	// Execute burn
	timeToApoapsisStream, err := orbit.TimeToApoapsisStream()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, timeToApoapsisStream.Close())
	})
	for timeToApoapsis-(burnTime/2) > 0 {
		select {
		case timeToApoapsis = <-timeToApoapsisStream.C:
		case <-ctx.Done():
			return
		}
	}

	require.NoError(t, control.SetThrottle(1.0))
	time.Sleep(time.Duration(math.Round((burnTime - 0.1) * float64(time.Second))))
	require.NoError(t, control.SetThrottle(0.05))

	remainingBurnStream, err := node.RemainingDeltaVStream()
	require.NoError(t, err)
	remainingBurn := <-remainingBurnStream.C
	for remainingBurn > 5 {
		select {
		case remainingBurn = <-remainingBurnStream.C:
		case <-ctx.Done():
			return
		}
	}

	require.NoError(t, control.SetThrottle(0))
	require.NoError(t, node.Remove())

}
