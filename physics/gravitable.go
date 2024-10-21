package physics

import "github.com/EliCDavis/vector/vector3"

type Gravitable interface {
	GetPosition() vector3.Vector[float64]
	GetMass() float64
	ApplyForce(force vector3.Vector[float64])
	GetUnit() *Unit
}
