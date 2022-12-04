package types

import "math"

// Vector2D is a 2D vector.
type Vector2D struct {
	X, Y float64
}

// NewVector2D creates a 2D vector from components.
func NewVector2D(x, y float64) Vector2D {
	return Vector2D{
		X: x,
		Y: y,
	}
}

// Vector2DFromTuple creates a vector from a tuple.
func Vector2DFromTuple(t Tuple2[float64, float64]) Vector2D {
	return Vector2D{
		X: t.A,
		Y: t.B,
	}
}

// Tuple converts the vector into a tuple.
func (v Vector2D) Tuple() Tuple2[float64, float64] {
	return NewTuple2(v.X, v.Y)
}

// Scale scales the vector by a constant value.
func (v Vector2D) Scale(k float64) Vector2D {
	return NewVector2D(k*v.X, k*v.Y)
}

// Add adds two vectors.
func (v Vector2D) Add(v2 Vector2D) Vector2D {
	return NewVector2D(v.X+v2.X, v.Y+v2.Y)
}

// Dot computes the dot product between 2 vectors.
func (v Vector2D) Dot(v2 Vector2D) float64 {
	return v.X*v2.X + v.Y*v2.Y
}

// Length is the length (L2 norm) of the vector.
func (v Vector2D) Length() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

// AngleBetween is the angle between two vectors, in radians.
func (v Vector2D) AngleBetween(v2 Vector2D) float64 {
	return math.Acos(v.Dot(v2) / (v.Length() * v2.Length()))
}

// Vector3D is a 3D vector.
type Vector3D struct {
	X, Y, Z float64
}

// NewVector3D creates a vector from components.
func NewVector3D(x, y, z float64) Vector3D {
	return Vector3D{
		X: x,
		Y: y,
		Z: z,
	}
}

// Vector3DFromTuple creates a vector from a tuple.
func Vector3DFromTuple(t Tuple3[float64, float64, float64]) Vector3D {
	return Vector3D{
		X: t.A,
		Y: t.B,
		Z: t.C,
	}
}

// Tuple converts the vector into a tuple.
func (v Vector3D) Tuple() Tuple3[float64, float64, float64] {
	return NewTuple3(v.X, v.Y, v.Z)
}

// Scale scales the vector by a constant value.
func (v Vector3D) Scale(k float64) Vector3D {
	return NewVector3D(k*v.X, k*v.Y, k*v.Z)
}

// Add adds two vectors.
func (v Vector3D) Add(v2 Vector3D) Vector3D {
	return NewVector3D(v.X+v2.X, v.Y+v2.Y, v.Z+v2.Z)
}

// Dot computes the dot product between 2 vectors.
func (v Vector3D) Dot(v2 Vector3D) float64 {
	return v.X*v2.X + v.Y*v2.Y + v.Z*v2.Z
}

// Length is the length (L2 norm) of the vector.
func (v Vector3D) Length() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

// AngleBetween is the angle between two vectors, in radians.
func (v Vector3D) AngleBetween(v2 Vector3D) float64 {
	return math.Acos(v.Dot(v2) / (v.Length() * v2.Length()))
}

// Cross is the cross product between two vectors.
func (v Vector3D) Cross(v2 Vector3D) Vector3D {
	return NewVector3D(
		v.Y*v2.Z-v.Z*v2.Y,
		-(v.X*v2.Z - v.Z*v2.X),
		v.X*v2.Y-v.Y*v2.X,
	)
}
