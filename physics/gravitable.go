package physics

import "github.com/EliCDavis/vector/vector3"

// Gravitable defines objects affected by gravity
type Gravitable interface {
	Position() vector3.Vector[float64]
	Mass() float64
	Unit() *Unit
}
