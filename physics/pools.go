package physics

import "sync"

// Shared object pools for the physics package
var (
	// Pool for temporary vector calculations
	VectorPool = sync.Pool{
		New: func() interface{} {
			return make([]float64, 3)
		},
	}

	// Pool for near units slices
	NearUnitsPool = sync.Pool{
		New: func() interface{} {
			slice := make([]*Unit, 0, 16)
			return &slice
		},
	}
)
