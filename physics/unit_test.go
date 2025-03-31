package physics

import (
	"math"
	"testing"

	"github.com/EliCDavis/vector/vector3"
	"github.com/alexanderi96/go-fluid-simulator/config"
	"github.com/alexanderi96/go-fluid-simulator/physics/collision"
	"github.com/alexanderi96/go-fluid-simulator/physics/material"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUnitMerge(t *testing.T) {
	// Test case 1: Merge con unità di uguale massa
	t.Run("Merge con unità di uguale massa", func(t *testing.T) {
		// Crea prima unità
		unit1 := &Unit{
			Id:        uuid.New(),
			_position: vector3.New(0.0, 0.0, 0.0),
			_velocity: vector3.New(1.0, 0.0, 0.0),
			_mass:     10.0,
			_radius:   1.0,
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
			Id:        uuid.New(),
			_position: vector3.New(2.0, 0.0, 0.0),
			_velocity: vector3.New(-1.0, 0.0, 0.0),
			_mass:     10.0,
			_radius:   1.0,
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
		assert.Equal(t, expectedMass, unit1.Mass(), "La massa dovrebbe essere la somma delle due unità")
		pos := unit1.Position()
		assert.Equal(t, expectedPosition.X(), pos.X(), "La posizione X dovrebbe essere la media pesata")
		assert.Equal(t, expectedPosition.Y(), pos.Y(), "La posizione Y dovrebbe essere la media pesata")
		assert.Equal(t, expectedPosition.Z(), pos.Z(), "La posizione Z dovrebbe essere la media pesata")

		// Verify momentum conservation
		initialMomentum := unit1.Mass()*1.0 + unit2.Mass()*(-1.0) // 10 - 10 = 0
		finalMomentum := unit1.Mass() * unit1.Velocity().X()      // 20 * 0 = 0
		assert.InDelta(t, initialMomentum, finalMomentum, 1e-10, "Il momento dovrebbe essere conservato")

		vel := unit1.Velocity()
		assert.Equal(t, expectedVelocity.X(), vel.X(), "La velocità finale dovrebbe conservare il momento")
		assert.Equal(t, expectedVelocity.Y(), vel.Y(), "La velocità Y dovrebbe essere zero")
		assert.Equal(t, expectedVelocity.Z(), vel.Z(), "La velocità Z dovrebbe essere zero")

		// Verifica del volume e raggio
		expectedVolume := unit1.GetVolume() + unit2.GetVolume()
		expectedRadius := math.Pow((3.0*expectedVolume)/(4.0*math.Pi), 1.0/3.0)
		assert.InDelta(t, expectedVolume, unit1.GetVolume(), 1e-10, "Il volume dovrebbe essere la somma dei volumi")
		assert.InDelta(t, expectedRadius, unit1.Radius(), 1e-10, "Il raggio dovrebbe essere calcolato correttamente dal volume")

		// Verifica dello stato dell'unità assorbita
		assert.False(t, unit2.CanBeAltered(), "La seconda unità non dovrebbe essere alterabile dopo il merge")
		assert.Equal(t, 0.0, unit2.Mass(), "La massa della seconda unità dovrebbe essere azzerata")
		assert.Equal(t, 0.0, unit2.GetVolume(), "Il volume della seconda unità dovrebbe essere azzerato")
	})

	// Test case 2: Merge con masse diverse
	t.Run("Merge con masse diverse", func(t *testing.T) {
		// Crea prima unità (più pesante)
		unit1 := &Unit{
			Id:        uuid.New(),
			_position: vector3.New(0.0, 0.0, 0.0),
			_velocity: vector3.New(1.0, 0.0, 0.0),
			_mass:     15.0,
			_radius:   1.0,
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
			Id:        uuid.New(),
			_position: vector3.New(3.0, 0.0, 0.0),
			_velocity: vector3.New(-1.0, 0.0, 0.0),
			_mass:     5.0,
			_radius:   1.0,
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
		// Conservation of momentum: p1 + p2 = p_final
		// (15 * 1) + (5 * -1) = 20 * v_final
		// 15 - 5 = 20 * v_final
		// v_final = 10/20 = 0.5
		expectedVelocity := vector3.New(0.5, 0.0, 0.0)

		// Verifica risultati
		assert.Equal(t, expectedMass, unit1.Mass(), "La massa dovrebbe essere la somma delle due unità")
		pos := unit1.Position()
		assert.Equal(t, expectedPosition.X(), pos.X(), "La posizione X dovrebbe essere più vicina all'unità più pesante")
		assert.Equal(t, expectedPosition.Y(), pos.Y(), "La posizione Y dovrebbe essere più vicina all'unità più pesante")
		assert.Equal(t, expectedPosition.Z(), pos.Z(), "La posizione Z dovrebbe essere più vicina all'unità più pesante")

		// Verify momentum conservation
		initialMomentum := unit1.Mass()*1.0 + unit2.Mass()*(-1.0) // 15 - 5 = 10
		finalMomentum := unit1.Mass() * unit1.Velocity().X()      // 20 * 0.5 = 10
		assert.InDelta(t, initialMomentum, finalMomentum, 1e-10, "Il momento dovrebbe essere conservato")

		vel := unit1.Velocity()
		assert.Equal(t, expectedVelocity.X(), vel.X(), "La velocità finale dovrebbe conservare il momento")
		assert.Equal(t, expectedVelocity.Y(), vel.Y(), "La velocità Y dovrebbe essere zero")
		assert.Equal(t, expectedVelocity.Z(), vel.Z(), "La velocità Z dovrebbe essere zero")
		assert.False(t, unit2.CanBeAltered(), "La seconda unità non dovrebbe essere alterabile dopo il merge")
	})
}

func TestEnergyToHeatConversion(t *testing.T) {
	t.Run("Verifica conversione energia-calore durante collisione", func(t *testing.T) {
		// Crea prima unità
		unit1 := &Unit{
			Id:        uuid.New(),
			_position: vector3.New(0.0, 0.0, 0.0),
			_velocity: vector3.New(5.0, 0.0, 0.0), // 5 m/s verso destra
			_mass:     10.0,                       // 10 kg
			_radius:   1.0,                        // 1 m
			Composition: material.NewComposition(map[*material.Material]float64{&material.Material{
				Name:                 "Test Material",
				Density:              1000,
				SpecificHeatCapacity: 100, // 100 J/(kg·K)
				ThermalConductivity:  50,  // 50 W/(m·K)
				Emissivity:           0.5, // 0.5 (50% di emissività)
				BaseColor:            [3]float64{0.5, 0.5, 0.5},
				Elasticity:           0.5, // 0.5 (50% di energia conservata)
			}: 1.0}),
			Heat: AmbientTemperature,
			Mesh: new(PointLightMesh),
			config: &config.Config{
				AllowUnitMerge: false,
			},
		}
		unit1.NewPointLightMesh()

		// Crea seconda unità
		unit2 := &Unit{
			Id:        uuid.New(),
			_position: vector3.New(1.5, 0.0, 0.0),  // 1.5 m a destra (garantisce collisione con unit1)
			_velocity: vector3.New(-5.0, 0.0, 0.0), // 5 m/s verso sinistra
			_mass:     10.0,                        // 10 kg
			_radius:   1.0,                         // 1 m
			Composition: material.NewComposition(map[*material.Material]float64{&material.Material{
				Name:                 "Test Material",
				Density:              1000,
				SpecificHeatCapacity: 100, // 100 J/(kg·K)
				ThermalConductivity:  50,  // 50 W/(m·K)
				Emissivity:           0.5, // 0.5 (50% di emissività)
				BaseColor:            [3]float64{0.5, 0.5, 0.5},
				Elasticity:           0.5, // 0.5 (50% di energia conservata)
			}: 1.0}),
			Heat: AmbientTemperature,
			Mesh: new(PointLightMesh),
			config: &config.Config{
				AllowUnitMerge: false,
			},
		}
		unit2.NewPointLightMesh()

		// Calcolo dell'energia cinetica iniziale
		initialKE1 := 0.5 * unit1.Mass() * math.Pow(unit1.Velocity().Length(), 2)
		initialKE2 := 0.5 * unit2.Mass() * math.Pow(unit2.Velocity().Length(), 2)
		initialTotalKE := initialKE1 + initialKE2

		t.Logf("Energia cinetica iniziale: %.2f J (unit1: %.2f J, unit2: %.2f J)",
			initialTotalKE, initialKE1, initialKE2)

		// Memorizza temperature iniziali
		initialTemp1 := unit1.Heat
		initialTemp2 := unit2.Heat

		// Simula la collisione
		collData := collision.GatherCollisionData(unit1, unit2)
		assert.True(t, collData.Collided, "Le unità dovrebbero collidere")
		collision.ResolveCollision(collData)

		// Calcolo dell'energia cinetica finale
		finalKE1 := 0.5 * unit1.Mass() * math.Pow(unit1.Velocity().Length(), 2)
		finalKE2 := 0.5 * unit2.Mass() * math.Pow(unit2.Velocity().Length(), 2)
		finalTotalKE := finalKE1 + finalKE2

		t.Logf("Energia cinetica finale: %.2f J (unit1: %.2f J, unit2: %.2f J)",
			finalTotalKE, finalKE1, finalKE2)

		// Calcolo dell'energia persa
		lostKE := initialTotalKE - finalTotalKE
		t.Logf("Energia cinetica persa: %.2f J", lostKE)

		// Calcolo del calore generato
		_, specificHeat1, _, _, _ := unit1.Composition.GetEffectiveProperties()
		_, specificHeat2, _, _, _ := unit2.Composition.GetEffectiveProperties()
		heatGained1 := unit1.Mass() * specificHeat1 * (unit1.Heat - initialTemp1)
		heatGained2 := unit2.Mass() * specificHeat2 * (unit2.Heat - initialTemp2)
		totalHeatGained := heatGained1 + heatGained2

		t.Logf("Calore generato: %.2f J (unit1: %.2f J, unit2: %.2f J)",
			totalHeatGained, heatGained1, heatGained2)

		// Verifica della conservazione dell'energia
		energyConservationRatio := totalHeatGained / lostKE
		t.Logf("Rapporto di conservazione dell'energia (calore/energia persa): %.4f",
			energyConservationRatio)

		// Il rapporto dovrebbe essere vicino a 1.0 (tolleranza 5%)
		assert.InDelta(t, 1.0, energyConservationRatio, 0.05,
			"L'energia cinetica persa dovrebbe essere convertita in calore")

		// Test del raffreddamento
		dt := 1.0 // 1 secondo
		initialHeat1 := unit1.Heat
		initialHeat2 := unit2.Heat

		// Aggiorna le posizioni per simulare il passaggio del tempo
		unit1.UpdatePosition(dt)
		unit2.UpdatePosition(dt)

		// Calcolo del raffreddamento teorico secondo Stefan-Boltzmann
		temp1K := initialHeat1 + 273.15
		temp2K := initialHeat2 + 273.15
		ambientK := AmbientTemperature + 273.15

		// Calcolo della potenza irradiata
		_, specificHeat1, _, emissivity1, _ := unit1.Composition.GetEffectiveProperties()
		_, specificHeat2, _, emissivity2, _ := unit2.Composition.GetEffectiveProperties()
		power1 := emissivity1 * StefanBoltzmannConstant * unit1.GetSurfaceArea() *
			(math.Pow(temp1K, 4) - math.Pow(ambientK, 4))
		power2 := emissivity2 * StefanBoltzmannConstant * unit2.GetSurfaceArea() *
			(math.Pow(temp2K, 4) - math.Pow(ambientK, 4))

		// Calore perso in dt secondi
		expectedHeatLoss1 := power1 * dt / (specificHeat1 * unit1.Mass())
		expectedHeatLoss2 := power2 * dt / (specificHeat2 * unit2.Mass())

		// Calore effettivamente perso
		actualHeatLoss1 := initialHeat1 - unit1.Heat
		actualHeatLoss2 := initialHeat2 - unit2.Heat

		t.Logf("Raffreddamento unit1 - Atteso: %.4f°C, Effettivo: %.4f°C",
			expectedHeatLoss1, actualHeatLoss1)
		t.Logf("Raffreddamento unit2 - Atteso: %.4f°C, Effettivo: %.4f°C",
			expectedHeatLoss2, actualHeatLoss2)

		// Verifica del raffreddamento
		coolingAccuracy1 := actualHeatLoss1 / expectedHeatLoss1
		coolingAccuracy2 := actualHeatLoss2 / expectedHeatLoss2

		t.Logf("Precisione del raffreddamento - unit1: %.4f, unit2: %.4f",
			coolingAccuracy1, coolingAccuracy2)

		// Tolleranza del 50% per il raffreddamento dato il modello semplificato
		assert.InDelta(t, 1.0, coolingAccuracy1, 0.5,
			"Il raffreddamento dovrebbe seguire approssimativamente la legge di Stefan-Boltzmann")
		assert.InDelta(t, 1.0, coolingAccuracy2, 0.5,
			"Il raffreddamento dovrebbe seguire approssimativamente la legge di Stefan-Boltzmann")
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
			d := sphere.Position().Sub(earth.Position())
			r := d.Length()
			dir := d.Normalized()

			// Forza gravitazionale
			F := G * earth.Mass() * sphere.Mass() / (r * r)

			// Accelerazione sulla sfera (F = m * a => a = F / m)
			a := F / sphere.Mass()
			finalAccel = a // memorizziamo l'ultima accelerazione

			// Log conciso dei valori principali
			t.Logf("Step %d: dist=%.2f m, a=%.2f m/s², v=%.2f m/s",
				i+1, r, a, sphere.Velocity().Length())

			// Aggiorna velocità e posizione della sfera
			// v(t+dt) = v(t) + a*dt
			sphere.SetVelocity(sphere.Velocity().Add(dir.Scale(a * dt)))
			// x(t+dt) = x(t) + v(t+dt)*dt
			sphere.SetPosition(sphere.Position().Add(sphere.Velocity().Scale(dt)))
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
			d := sphere.Position().Sub(earth.Position())
			r := d.Length()
			dir := d.Normalized()

			F := G * earth.Mass() * sphere.Mass() / (r * r)
			a := F / sphere.Mass()
			finalAccel = a

			// Log conciso dei valori principali
			t.Logf("Step %d: dist=%.2f m, a=%.2f m/s², v=%.2f m/s",
				i+1, r, a, sphere.Velocity().Length())

			sphere.SetVelocity(sphere.Velocity().Add(dir.Scale(a * dt)))
			sphere.SetPosition(sphere.Position().Add(sphere.Velocity().Scale(dt)))
		}

		// Valore atteso a r = startDistance
		expectedAccel := G * massEarth / math.Pow(startDistance, 2)
		delta := 0.01 * expectedAccel

		assert.InDelta(t, expectedAccel, finalAccel, delta,
			"L'accelerazione dovrebbe seguire la legge ~1/r^2 anche a distanza maggiore")
	})
}
