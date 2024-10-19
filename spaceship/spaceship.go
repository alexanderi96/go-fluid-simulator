package spaceship

import (
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/window"
)

type SpaceShip struct {
	Ship            *graphic.Mesh
	Speed           float32
	MaxSpeed        float32
	MaxEngineThrust float32
	Thrust          float32
	RotationSpeed   float32
	BreakingPower   float32
	Keys            map[window.Key]bool
	CameraOffset    *math32.Vector3
}

func SetupPlane(s *SpaceShip) {
	// Creiamo il corpo principale della nave
	bodyGeom := geometry.NewBox(1, 0.3, 2) // Corpo allungato e sottile
	bodyMat := material.NewStandard(math32.NewColor("Gray"))
	bodyMesh := graphic.NewMesh(bodyGeom, bodyMat)

	// Parte frontale appuntita (cockpit a forma di prisma)
	noseGeom := geometry.NewCone(0.2, 0.5, 4, 1, true) // Cono a base quadrata
	noseMat := material.NewStandard(math32.NewColor("Black"))
	noseMesh := graphic.NewMesh(noseGeom, noseMat)
	noseMesh.SetRotationX(math32.Pi / 2) // Ruotiamo il cono per farlo puntare in avanti
	noseMesh.SetPosition(0, 0.1, 1.2)    // Posizioniamo la parte frontale sulla nave

	// Creiamo le ali larghe, puntate verso il fronte e il basso
	wingGeom := geometry.NewBox(1.5, 0.05, 0.6) // Ali più larghe e sottili
	wingMat := material.NewStandard(math32.NewColor("Gray"))
	leftWingMesh := graphic.NewMesh(wingGeom, wingMat)
	rightWingMesh := graphic.NewMesh(wingGeom, wingMat)

	// Impostiamo l'inclinazione delle ali
	leftWingMesh.SetRotationZ(math32.DegToRad(20))   // Inclinazione verso il basso
	leftWingMesh.SetRotationY(math32.DegToRad(15))   // Inclinazione verso il fronte
	rightWingMesh.SetRotationZ(-math32.DegToRad(20)) // Inclinazione verso il basso
	rightWingMesh.SetRotationY(-math32.DegToRad(15)) // Inclinazione verso il fronte

	leftWingMesh.SetPosition(-0.9, -0.2, -0.3) // Posiziona l'ala sinistra
	rightWingMesh.SetPosition(0.9, -0.2, -0.3) // Posiziona l'ala destra

	// Aggiungiamo dei bracci anteriori (cannoni)
	armGeom := geometry.NewBox(0.05, 0.05, 0.6) // Bracci allungati
	armMat := material.NewStandard(math32.NewColor("Gray"))
	leftArmMesh := graphic.NewMesh(armGeom, armMat)
	rightArmMesh := graphic.NewMesh(armGeom, armMat)

	leftArmMesh.SetPosition(-0.4, 0, 0.8) // Posiziona il braccio sinistro verso la parte anteriore
	rightArmMesh.SetPosition(0.4, 0, 0.8) // Posiziona il braccio destro

	// Aggiungi tutte le componenti al corpo principale della nave
	bodyMesh.Add(noseMesh)      // Aggiungi il "cockpit" anteriore
	bodyMesh.Add(leftWingMesh)  // Aggiungi l'ala sinistra
	bodyMesh.Add(rightWingMesh) // Aggiungi l'ala destra
	bodyMesh.Add(leftArmMesh)   // Aggiungi il braccio sinistro
	bodyMesh.Add(rightArmMesh)  // Aggiungi il braccio destro

	// Aggiungi la nave alla scena
	s.Ship = bodyMesh

}

func UpdateMovement(s *SpaceShip) {
	if s.Keys[window.KeyW] && s.Thrust < s.MaxEngineThrust {
		s.Thrust += 0.1
	} else if s.Keys[window.KeyS] && s.Thrust > -s.MaxEngineThrust {
		s.Thrust += -0.1
	} else {
		s.Thrust = 0
	}

	s.Speed += s.Thrust
	s.Speed = math32.Clamp(s.Speed, -s.MaxSpeed, s.MaxSpeed)

	s.Ship.TranslateZ(s.Speed)

	if s.Keys[window.KeySpace] {
		if s.Speed > 0 {
			s.Speed = math32.Max(0, s.Speed-s.BreakingPower)
		} else if s.Speed < 0 {
			s.Speed = math32.Min(0, s.Speed+s.BreakingPower)
		}
	}
	if s.Keys[window.KeyE] {
		s.Ship.RotateY(-s.RotationSpeed)
	}
	if s.Keys[window.KeyQ] {
		s.Ship.RotateY(s.RotationSpeed)
	}
	if s.Keys[window.KeyA] {
		s.Ship.RotateZ(-s.RotationSpeed)
	}
	if s.Keys[window.KeyD] {
		s.Ship.RotateZ(s.RotationSpeed)
	}
	if s.Keys[window.KeyM] {
		s.Ship.RotateX(-s.RotationSpeed)
	}
	if s.Keys[window.KeyK] {
		s.Ship.RotateX(s.RotationSpeed)
	}
}
