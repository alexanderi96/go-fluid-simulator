package material

// Material represents the physical properties of a substance
type Material struct {
	Name                 string
	Density              float64    // kg/m³
	SpecificHeatCapacity float64    // J/(kg·K)
	ThermalConductivity  float64    // W/(m·K)
	Emissivity           float64    // 0-1
	BaseColor            [3]float64 // RGB values 0-1
	Elasticity           float64    // 0-1
}

// Common materials
var (
	Iron = Material{
		Name:                 "Iron",
		Density:              7874,
		SpecificHeatCapacity: 450,
		ThermalConductivity:  80.2,
		Emissivity:           0.3,
		BaseColor:            [3]float64{0.6, 0.6, 0.6},
		Elasticity:           0.7,
	}

	Copper = Material{
		Name:                 "Copper",
		Density:              8960,
		SpecificHeatCapacity: 386,
		ThermalConductivity:  401,
		Emissivity:           0.03,
		BaseColor:            [3]float64{0.85, 0.45, 0.2},
		Elasticity:           0.75,
	}

	Ice = Material{
		Name:                 "Ice",
		Density:              917,
		SpecificHeatCapacity: 2108,
		ThermalConductivity:  2.18,
		Emissivity:           0.97,
		BaseColor:            [3]float64{0.8, 0.9, 0.95},
		Elasticity:           0.3,
	}
)

// Composition represents the material makeup of a unit
type Composition struct {
	Materials map[*Material]float64 // Material -> Volume Fraction (0-1)
}

// NewComposition creates a new composition with the given materials and their volume fractions
func NewComposition(fractions map[*Material]float64) *Composition {
	// Normalize fractions to ensure they sum to 1
	total := 0.0
	for _, fraction := range fractions {
		total += fraction
	}

	normalized := make(map[*Material]float64)
	for material, fraction := range fractions {
		normalized[material] = fraction / total
	}

	return &Composition{
		Materials: normalized,
	}
}

// GetEffectiveProperties calculates the weighted average properties of the composition
func (c *Composition) GetEffectiveProperties() (density, specificHeat, thermalConductivity, emissivity, elasticity float64) {
	for material, fraction := range c.Materials {
		density += material.Density * fraction
		specificHeat += material.SpecificHeatCapacity * fraction
		thermalConductivity += material.ThermalConductivity * fraction
		emissivity += material.Emissivity * fraction
		elasticity += material.Elasticity * fraction
	}
	return
}

// GetColor calculates the weighted average color of the composition
func (c *Composition) GetColor() [3]float64 {
	var color [3]float64
	for material, fraction := range c.Materials {
		for i := 0; i < 3; i++ {
			color[i] += material.BaseColor[i] * fraction
		}
	}
	return color
}
