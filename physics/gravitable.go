package physics

import "github.com/EliCDavis/vector/vector3"

// Gravitable defines objects affected by gravity
type Gravitable interface {
	GetPosition() vector3.Vector[float64]
	GetMass() float64
	GetUnit() *Unit
}
