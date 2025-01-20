package physics

import (
	"math"

	"github.com/EliCDavis/vector/vector3"
	"github.com/alexanderi96/go-fluid-simulator/physics/material"
	"github.com/google/uuid"
)

// NewTestSphere crea una sfera di test standard di 1m³
func NewTestSphere(position vector3.Vector[float64]) *Unit {
	// Per un volume di 1m³, calcoliamo il raggio
	radius := math.Pow((3.0 / (4.0 * math.Pi)), 1.0/3.0)

	sphere := &Unit{
		Id:             uuid.New(),
		Position:       position,
		Velocity:       vector3.Zero[float64](),
		Mass:           1.0, // 1 kg
		Radius:         radius,
		MassMultiplier: 1.0,
		Acceleration:   vector3.Zero[float64](),
		Heat:           20.0, // Temperatura ambiente di default
		Composition: material.NewComposition(map[*material.Material]float64{
			&material.Material{
				Name:                 "Test Material",
				Density:              1.0,  // kg/m³
				SpecificHeatCapacity: 450,  // J/(kg⋅K)
				ThermalConductivity:  80.2, // W/(m⋅K)
				Emissivity:           0.3,
				BaseColor:            [3]float64{0.6, 0.6, 0.6},
				Elasticity:           0.7,
			}: 1.0,
		}),
		Mesh: new(PointLightMesh),
	}
	sphere.NewPointLightMesh()
	return sphere
}

// NewEarth crea una nuova istanza della Terra con le proprietà standard
func NewEarth(position vector3.Vector[float64]) *Unit {
	earth := &Unit{
		Id:             uuid.New(),
		Position:       position,
		Velocity:       vector3.Zero[float64](),
		Mass:           EarthMass,
		Radius:         EarthRadius,
		MassMultiplier: 1.0,
		Acceleration:   vector3.Zero[float64](),
		Heat:           20.0, // Temperatura ambiente di default
		Composition: material.NewComposition(map[*material.Material]float64{
			&material.Material{
				Name:                 "Earth",
				Density:              EarthDensity,
				SpecificHeatCapacity: 1000, // Valore approssimato
				ThermalConductivity:  2.0,  // W/(m⋅K)
				Emissivity:           0.95, // Valore medio approssimato
				BaseColor:            [3]float64{0.2, 0.5, 0.8},
				Elasticity:           0.1,
			}: 1.0,
		}),
		Mesh: new(PointLightMesh),
	}
	earth.NewPointLightMesh()
	return earth
}

// NewMoon crea una nuova istanza della Luna con le proprietà standard
func NewMoon(position vector3.Vector[float64]) *Unit {
	moon := &Unit{
		Id:             uuid.New(),
		Position:       position,
		Velocity:       vector3.Zero[float64](),
		Mass:           MoonMass,
		Radius:         MoonRadius,
		MassMultiplier: 1.0,
		Acceleration:   vector3.Zero[float64](),
		Heat:           20.0, // Temperatura ambiente di default
		Composition: material.NewComposition(map[*material.Material]float64{
			&material.Material{
				Name:                 "Moon",
				Density:              MoonDensity,
				SpecificHeatCapacity: 920, // J/(kg⋅K)
				ThermalConductivity:  2.0, // W/(m⋅K)
				Emissivity:           0.97,
				BaseColor:            [3]float64{0.8, 0.8, 0.8},
				Elasticity:           0.1,
			}: 1.0,
		}),
		Mesh: new(PointLightMesh),
	}
	moon.NewPointLightMesh()
	return moon
}

// NewSun crea una nuova istanza del Sole con le proprietà standard
func NewSun(position vector3.Vector[float64]) *Unit {
	sun := &Unit{
		Id:             uuid.New(),
		Position:       position,
		Velocity:       vector3.Zero[float64](),
		Mass:           SunMass,
		Radius:         SunRadius,
		MassMultiplier: 1.0,
		Acceleration:   vector3.Zero[float64](),
		Heat:           20.0, // Temperatura ambiente di default
		Composition: material.NewComposition(map[*material.Material]float64{
			&material.Material{
				Name:                 "Sun",
				Density:              SunDensity,
				SpecificHeatCapacity: 12000, // Valore approssimato
				ThermalConductivity:  100,   // Valore approssimato
				Emissivity:           1.0,
				BaseColor:            [3]float64{1.0, 0.9, 0.0},
				Elasticity:           0.1,
			}: 1.0,
		}),
		Mesh: new(PointLightMesh),
	}
	sun.NewPointLightMesh()
	return sun
}
