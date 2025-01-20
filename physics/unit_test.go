package physics

import (
	"math"
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
			Mesh: new(PointLightMesh),
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
			Mesh: new(PointLightMesh),
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

		// Verifica del volume e raggio
		expectedVolume := unit1.GetVolume() + unit2.GetVolume()
		expectedRadius := math.Pow((3.0*expectedVolume)/(4.0*math.Pi), 1.0/3.0)
		assert.InDelta(t, expectedVolume, unit1.GetVolume(), 1e-10, "Il volume dovrebbe essere la somma dei volumi")
		assert.InDelta(t, expectedRadius, unit1.GetRadius(), 1e-10, "Il raggio dovrebbe essere calcolato correttamente dal volume")

		// Verifica dello stato dell'unità assorbita
		assert.False(t, unit2.CanBeAltered(), "La seconda unità non dovrebbe essere alterabile dopo il merge")
		assert.Equal(t, 0.0, unit2.GetMass(), "La massa della seconda unità dovrebbe essere azzerata")
		assert.Equal(t, 0.0, unit2.GetVolume(), "Il volume della seconda unità dovrebbe essere azzerato")
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
			Mesh: new(PointLightMesh),
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
			Mesh: new(PointLightMesh),
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

func TestGravitationalAcceleration(t *testing.T) {
	t.Run("Accelerazione vicino alla superficie terrestre", func(t *testing.T) {
		// Parametri "Terra"
		massEarth := EarthMass
		radiusEarth := EarthRadius

		// Crea "Terra"
		earth := NewEarth(vector3.Zero[float64]())

		// Crea "Sfera" di test
		sphere := NewTestSphere(vector3.New(radiusEarth+100.0, 0.0, 0.0)) // 100 m sopra la superficie

		// Simulazione semplificata: 10 step di 1 secondo
		dt := 1.0
		steps := 10

		var finalAccel float64
		for i := 0; i < steps; i++ {
			// Calcola distanza e direzione
			d := sphere.Position.Sub(earth.Position)
			r := d.Length()
			dir := d.Normalized()

			// Forza gravitazionale
			F := G * earth.Mass * sphere.Mass / (r * r)

			// Accelerazione sulla sfera (F = m * a => a = F / m)
			a := F / sphere.Mass
			finalAccel = a // memorizziamo l'ultima accelerazione

			// Log conciso dei valori principali
			t.Logf("Step %d: dist=%.2f m, a=%.2f m/s², v=%.2f m/s",
				i+1, r, a, sphere.Velocity.Length())

			// Aggiorna velocità e posizione della sfera
			// v(t+dt) = v(t) + a*dt
			sphere.Velocity = sphere.Velocity.Add(dir.Scale(a * dt))
			// x(t+dt) = x(t) + v(t+dt)*dt
			sphere.Position = sphere.Position.Add(sphere.Velocity.Scale(dt))
		}

		// Calcolo dell'accelerazione teorica a r = (R_terra + 100 m)
		expectedAccel := G * massEarth / math.Pow(radiusEarth+100.0, 2.0)
		// Tolleranza relativa (es. 1%)
		delta := 0.01 * expectedAccel

		assert.InDelta(t, expectedAccel, finalAccel, delta,
			"L'accelerazione dovrebbe essere vicina a G*M / (r^2)")
	})

	t.Run("Verifica 1/r^2 a distanza maggiore", func(t *testing.T) {
		// Parametri "Terra"
		massEarth := EarthMass
		radiusEarth := EarthRadius

		// Crea "Terra"
		earth := NewEarth(vector3.Zero[float64]())

		// Crea "Sfera" di test
		startDistance := radiusEarth * 10.0 // 10 raggi terrestri di distanza
		sphere := NewTestSphere(vector3.New(startDistance, 0.0, 0.0))

		dt := 1.0
		steps := 10
		var finalAccel float64
		for i := 0; i < steps; i++ {
			d := sphere.Position.Sub(earth.Position)
			r := d.Length()
			dir := d.Normalized()

			F := G * earth.Mass * sphere.Mass / (r * r)
			a := F / sphere.Mass
			finalAccel = a

			// Log conciso dei valori principali
			t.Logf("Step %d: dist=%.2f m, a=%.2f m/s², v=%.2f m/s",
				i+1, r, a, sphere.Velocity.Length())

			sphere.Velocity = sphere.Velocity.Add(dir.Scale(a * dt))
			sphere.Position = sphere.Position.Add(sphere.Velocity.Scale(dt))
		}

		// Valore atteso a r = startDistance
		expectedAccel := G * massEarth / math.Pow(startDistance, 2)
		delta := 0.01 * expectedAccel

		assert.InDelta(t, expectedAccel, finalAccel, delta,
			"L'accelerazione dovrebbe seguire la legge ~1/r^2 anche a distanza maggiore")
	})
}
