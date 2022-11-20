package api

// Tuple2 is a generic tuple with 2 elements.
type Tuple2[T, U any] struct {
	A T
	B U
}

// NewTuple2 creates a new Tuple2.
func NewTuple2[T, U any](a T, b U) Tuple2[T, U] {
	return Tuple2[T, U]{
		A: a,
		B: b,
	}
}

// Tuple3 is a generic tuple with 3 elements.
type Tuple3[T, U, V any] struct {
	A T
	B U
	C V
}

// NewTuple3 creates a new Tuple3.
func NewTuple3[T, U, V any](a T, b U, c V) Tuple3[T, U, V] {
	return Tuple3[T, U, V]{
		A: a,
		B: b,
		C: c,
	}
}

// Tuple4 is a generic tuple with 4 elements.
type Tuple4[T, U, V, W any] struct {
	A T
	B U
	C V
	D W
}

// NewTuple4 creates a new Tuple4.
func NewTuple4[T, U, V, W any](a T, b U, c V, d W) Tuple4[T, U, V, W] {
	return Tuple4[T, U, V, W]{
		A: a,
		B: b,
		C: c,
		D: d,
	}
}

// Real represents a real number.
type Real interface {
	float32 | float64
}

// Vector2D is a 2D vector.
type Vector2D struct {
	X, Y float64
}

// Vector2DFromTuple creates a vector from a tuple.
func Vector2DFromTuple[T Real](t Tuple2[float64, float64]) Vector2D {
	return Vector2D{
		X: t.A,
		Y: t.B,
	}
}

// Tuple converts the vector into a tuple.
func (v Vector2D) Tuple() Tuple2[float64, float64] {
	return NewTuple2(v.X, v.Y)
}

// Vector3D is a 3D vector.
type Vector3D struct {
	X, Y, Z float64
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

// Color is an RGB color.
type Color[T Real] struct {
	R, G, B T
}

// ColorFromTuple creates a color from a tuple.
func ColorFromTuple[T Real](t Tuple3[T, T, T]) Color[T] {
	return Color[T]{
		R: t.A,
		G: t.B,
		B: t.C,
	}
}

// Tuple converts the vector into a tuple.
func (c Color[T]) Tuple() Tuple3[T, T, T] {
	return NewTuple3(c.R, c.G, c.B)
}

// ToFloat64 converts the color to a float64 color.
func (c Color[T]) ToFloat64() Color[float64] {
	return Color[float64]{
		R: (float64)(c.R),
		G: (float64)(c.G),
		B: (float64)(c.B),
	}
}

// ToFloat32 converts the color to a float32 color.
func (c Color[T]) ToFloat32() Color[float32] {
	return Color[float32]{
		R: (float32)(c.R),
		G: (float32)(c.G),
		B: (float32)(c.B),
	}
}

// Quaternion is a quaternion.
type Quaternion struct {
	X, Y, Z, W float64
}

// QuaternionFromTuple creates a quaternion from a tuple.
func QuaternionFromTuple(t Tuple4[float64, float64, float64, float64]) Quaternion {
	return Quaternion{
		X: t.A,
		Y: t.B,
		Z: t.C,
		W: t.D,
	}
}

// Tuple converts the quaternion into a tuple.
func (q Quaternion) Tuple() Tuple4[float64, float64, float64, float64] {
	return NewTuple4(q.X, q.Y, q.Z, q.W)
}

// IdentityQuaternion returns the identity quaternion (0i+0j+0k+1).
func IdentityQuaternion() Quaternion {
	return Quaternion{0, 0, 0, 1}
}

// ConnectionClient holds the info for clients connected to a kRPC server.
type ConnectedClient struct {
	ID      [16]byte
	Name    string
	Address string
}

// ConnectedClientFromTuple creates a connected client from a tuple.
func ConnectedClientFromTuple(t Tuple3[[]byte, string, string]) ConnectedClient {
	var id [16]byte
	copy(id[:], t.A)
	return ConnectedClient{
		ID:      id,
		Name:    t.B,
		Address: t.C,
	}
}
