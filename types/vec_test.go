package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const delta = 0.0001

func requireVectorsEqual(t *testing.T, expected, actual any, msgAndArgs ...any) {
	cmp := func(e, a float64) {
		require.InDelta(t, e, a, delta, msgAndArgs...)
	}
	ev2, ev2ok := expected.(Vector2D)
	av2, av2ok := actual.(Vector2D)
	if ev2ok && av2ok {
		cmp(ev2.X, av2.X)
		cmp(ev2.Y, av2.Y)
		return
	}

	ev3, ev3ok := expected.(Vector3D)
	av3, av3ok := actual.(Vector3D)
	if ev3ok && av3ok {
		cmp(ev3.X, av3.X)
		cmp(ev3.Y, av3.Y)
		cmp(ev3.Z, av3.Z)
		return
	}

	require.Fail(t, "Arguments should be either both Vector2D or both Vector3D", msgAndArgs...)
}

func TestVectorScale(t *testing.T) {
	k := 3.0
	v2in := NewVector2D(3, 4)
	v2out := NewVector2D(9, 12)
	v3in := NewVector3D(3, 5, -7)
	v3out := NewVector3D(9, 15, -21)

	requireVectorsEqual(t, v2out, v2in.Scale(k))
	requireVectorsEqual(t, v3out, v3in.Scale(k))
}

func TestVectorAdd(t *testing.T) {
	v2left := NewVector2D(1, -2)
	v2right := NewVector2D(5.5, 4)
	v2out := NewVector2D(6.5, 2)
	requireVectorsEqual(t, v2out, v2left.Add(v2right))

	v3left := NewVector3D(9.3, 0.3, -29.8)
	v3right := NewVector3D(-10.1, 3, 3)
	v3out := NewVector3D(-0.8, 3.3, -26.8)
	requireVectorsEqual(t, v3out, v3left.Add(v3right))
}

func TestVectorDot(t *testing.T) {
	v2left := NewVector2D(-1, 2)
	v2right := NewVector2D(5, -1)
	v2out := -7.0
	require.InDelta(t, v2out, v2left.Dot(v2right), delta)

	v3left := NewVector3D(-7, 7, 4)
	v3right := NewVector3D(4, 17, -3)
	v3out := 79.0
	require.InDelta(t, v3out, v3left.Dot(v3right), delta)
}

func TestVectorLength(t *testing.T) {
	v2 := NewVector2D(3, -4)
	v2out := 5.0
	require.InDelta(t, v2out, v2.Length(), delta)

	v3 := NewVector3D(-11.5, 4.3, -5)
	v3out := 13.25669
	require.InDelta(t, v3out, v3.Length(), delta)
}

func TestVectorAngleBetween(t *testing.T) {
	v2left := NewVector2D(1, -3)
	v2right := NewVector2D(6, -4)
	v2out := 0.66104
	require.InDelta(t, v2out, v2left.AngleBetween(v2right), delta)

	v3left := NewVector3D(4, -2, 5)
	v3right := NewVector3D(11, 23, 4)
	v3out := 1.46663
	require.InDelta(t, v3out, v3left.AngleBetween(v3right), delta)
}

func TestVectorCross(t *testing.T) {
	vleft := NewVector3D(4, -6, -3)
	vright := NewVector3D(0.4, -4, -0.77)
	vout := NewVector3D(-7.38, 1.88, -13.6)
	requireVectorsEqual(t, vout, vleft.Cross(vright))
}
