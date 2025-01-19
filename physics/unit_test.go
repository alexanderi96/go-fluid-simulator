package physics

import (
	"testing"

	"github.com/EliCDavis/vector/vector3"
	"github.com/alexanderi96/go-fluid-simulator/physics/material"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUnitMerge(t *testing.T) {
	// Test case 1: Merge con unità di uguale massa
	t.Run("Merge con unità di uguale massa", func(t *testing.T) {
		// Crea prima unità
		unit1 := &Unit{
			Id:       uuid.New(),
			Position: vector3.New(0.0, 0.0, 0.0),
			Velocity: vector3.New(1.0, 0.0, 0.0),
			Mass:     10.0,
			Radius:   1.0,
			Composition: material.NewComposition(map[*material.Material]float64{&material.Material{
				Name:                 "Test Iron",
				Density:              7874,
				SpecificHeatCapacity: 450,
				ThermalConductivity:  80.2,
				Emissivity:           0.3,
				BaseColor:            [3]float64{0.6, 0.6, 0.6},
				Elasticity:           0.7,
			}: 1.0}),
			canBeAltered: true,
			volume:       (4.0 / 3.0) * 3.14159 * 1.0, // Pre-calcola il volume
			Mesh:         new(PointLightMesh),
		}
		unit1.NewPointLightMesh() // Initialize mesh properly

		// Crea seconda unità
		unit2 := &Unit{
			Id:       uuid.New(),
			Position: vector3.New(2.0, 0.0, 0.0),
			Velocity: vector3.New(-1.0, 0.0, 0.0),
			Mass:     10.0,
			Radius:   1.0,
			Composition: material.NewComposition(map[*material.Material]float64{&material.Material{
				Name:                 "Test Iron",
				Density:              7874,
				SpecificHeatCapacity: 450,
				ThermalConductivity:  80.2,
				Emissivity:           0.3,
				BaseColor:            [3]float64{0.6, 0.6, 0.6},
				Elasticity:           0.7,
			}: 1.0}),
			canBeAltered: true,
			volume:       (4.0 / 3.0) * 3.14159 * 1.0, // Pre-calcola il volume
			Mesh:         new(PointLightMesh),
		}
		unit2.NewPointLightMesh() // Initialize mesh properly

		// Esegui il merge
		unit1.Merge(unit2)

		// Valori attesi
		expectedMass := 20.0                           // Somma delle masse
		expectedPosition := vector3.New(1.0, 0.0, 0.0) // Media pesata delle posizioni (masse uguali)
		expectedVelocity := vector3.New(0.0, 0.0, 0.0) // Media pesata delle velocità (masse uguali)

		// Verifica risultati
		assert.Equal(t, expectedMass, unit1.Mass, "La massa dovrebbe essere la somma delle due unità")
		assert.Equal(t, expectedPosition.X(), unit1.Position.X(), "La posizione X dovrebbe essere la media pesata")
		assert.Equal(t, expectedPosition.Y(), unit1.Position.Y(), "La posizione Y dovrebbe essere la media pesata")
		assert.Equal(t, expectedPosition.Z(), unit1.Position.Z(), "La posizione Z dovrebbe essere la media pesata")
		assert.Equal(t, expectedVelocity.X(), unit1.Velocity.X(), "La velocità X dovrebbe essere la media pesata")
		assert.Equal(t, expectedVelocity.Y(), unit1.Velocity.Y(), "La velocità Y dovrebbe essere la media pesata")
		assert.Equal(t, expectedVelocity.Z(), unit1.Velocity.Z(), "La velocità Z dovrebbe essere la media pesata")
		assert.False(t, unit2.CanBeAltered(), "La seconda unità non dovrebbe essere alterabile dopo il merge")
	})

	// Test case 2: Merge con masse diverse
	t.Run("Merge con masse diverse", func(t *testing.T) {
		// Crea prima unità (più pesante)
		unit1 := &Unit{
			Id:       uuid.New(),
			Position: vector3.New(0.0, 0.0, 0.0),
			Velocity: vector3.New(1.0, 0.0, 0.0),
			Mass:     15.0,
			Radius:   1.0,
			Composition: material.NewComposition(map[*material.Material]float64{&material.Material{
				Name:                 "Test Iron",
				Density:              7874,
				SpecificHeatCapacity: 450,
				ThermalConductivity:  80.2,
				Emissivity:           0.3,
				BaseColor:            [3]float64{0.6, 0.6, 0.6},
				Elasticity:           0.7,
			}: 1.0}),
			canBeAltered: true,
			volume:       (4.0 / 3.0) * 3.14159 * 1.0, // Pre-calcola il volume
			Mesh:         new(PointLightMesh),
		}
		unit1.NewPointLightMesh() // Initialize mesh properly

		// Crea seconda unità (più leggera)
		unit2 := &Unit{
			Id:       uuid.New(),
			Position: vector3.New(3.0, 0.0, 0.0),
			Velocity: vector3.New(-1.0, 0.0, 0.0),
			Mass:     5.0,
			Radius:   1.0,
			Composition: material.NewComposition(map[*material.Material]float64{&material.Material{
				Name:                 "Test Iron",
				Density:              7874,
				SpecificHeatCapacity: 450,
				ThermalConductivity:  80.2,
				Emissivity:           0.3,
				BaseColor:            [3]float64{0.6, 0.6, 0.6},
				Elasticity:           0.7,
			}: 1.0}),
			canBeAltered: true,
			volume:       (4.0 / 3.0) * 3.14159 * 1.0, // Pre-calcola il volume
			Mesh:         new(PointLightMesh),
		}
		unit2.NewPointLightMesh() // Initialize mesh properly

		// Esegui il merge
		unit1.Merge(unit2)

		// Valori attesi
		expectedMass := 20.0 // Somma delle masse
		// La posizione dovrebbe essere più vicina all'unità più pesante
		// (15 * 0 + 5 * 3) / 20 = 0.75
		expectedPosition := vector3.New(0.75, 0.0, 0.0)
		// (15 * 1 + 5 * -1) / 20 = 0.5
		expectedVelocity := vector3.New(0.5, 0.0, 0.0)

		// Verifica risultati
		assert.Equal(t, expectedMass, unit1.Mass, "La massa dovrebbe essere la somma delle due unità")
		assert.Equal(t, expectedPosition.X(), unit1.Position.X(), "La posizione X dovrebbe essere più vicina all'unità più pesante")
		assert.Equal(t, expectedPosition.Y(), unit1.Position.Y(), "La posizione Y dovrebbe essere più vicina all'unità più pesante")
		assert.Equal(t, expectedPosition.Z(), unit1.Position.Z(), "La posizione Z dovrebbe essere più vicina all'unità più pesante")
		assert.Equal(t, expectedVelocity.X(), unit1.Velocity.X(), "La velocità X dovrebbe essere influenzata maggiormente dall'unità più pesante")
		assert.Equal(t, expectedVelocity.Y(), unit1.Velocity.Y(), "La velocità Y dovrebbe essere influenzata maggiormente dall'unità più pesante")
		assert.Equal(t, expectedVelocity.Z(), unit1.Velocity.Z(), "La velocità Z dovrebbe essere influenzata maggiormente dall'unità più pesante")
		assert.False(t, unit2.CanBeAltered(), "La seconda unità non dovrebbe essere alterabile dopo il merge")
	})
}
