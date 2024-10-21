package spaceship

import (
	"log"

	"github.com/EliCDavis/vector/vector3"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/loader/obj"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/window"
)

type SpaceShip struct {
	Ship            *core.Node
	Speed           float32
	MaxSpeed        float32
	MaxEngineThrust float32
	Thrust          float32
	RotationSpeed   float32
	BreakingPower   float32
	Position        vector3.Vector[float64]
	Acceleration    vector3.Vector[float64]
	Mass            float64
	Keys            map[window.Key]bool
	CameraOffset    *math32.Vector3
}

func (s *SpaceShip) ApplyForce(f vector3.Vector[float64]) {
	s.Acceleration = s.Acceleration.Add(f.Scale(1 / s.Mass))
}

// func (s *SpaceShip) UpdatePosition(dt float64) {
// 	s.Position = s.Position.Add(s.Velocity.Scale(dt))
// 	s.Velocity = s.Velocity.Add(s.Acceleration.Scale(dt))

// 	s.Acceleration = vector3.Zero[float64]()

// 	s.Mesh.SetPosition(s.Position.ToFloat32().X(), s.Position.ToFloat32().Y(), s.Position.ToFloat32().Z())

// }

func (s *SpaceShip) SetupShip() {
	// Corpo principale: un cilindro conica per dare un aspetto più fluido e aerodinamico
	bodyGeom := geometry.NewCylinder(0.3, 0.2, 2, 32, true, true) // Più segmenti per rotondità e una base più piccola
	bodyMat := material.NewStandard(math32.NewColor("Silver"))
	bodyMesh := graphic.NewMesh(bodyGeom, bodyMat)
	bodyMesh.SetRotationZ(math32.Pi / 2) // Ruotiamo il cilindro orizzontalmente

	// Cockpit: utilizziamo una sfera per un aspetto moderno e liscio
	cockpitGeom := geometry.NewSphere(0.4, 16, 16) // Più segmenti per una sfera più liscia
	cockpitMat := material.NewStandard(math32.NewColor("Black"))
	cockpitMesh := graphic.NewMesh(cockpitGeom, cockpitMat)
	cockpitMesh.SetPosition(0, 0, 1.1)

	// Ali: usiamo dei pannelli trapezoidali per un look più avanzato
	wingGeom := geometry.NewBox(1.5, 0.05, 0.5) // Ali più ampie e sottili per dare l'idea di velocità
	wingMat := material.NewStandard(math32.NewColor("DarkGray"))
	leftWingMesh := graphic.NewMesh(wingGeom, wingMat)
	rightWingMesh := graphic.NewMesh(wingGeom, wingMat)

	// Rotazione e posizionamento delle ali
	leftWingMesh.SetRotationY(math32.DegToRad(10)) // Leggera angolazione verso il basso
	rightWingMesh.SetRotationY(-math32.DegToRad(10))
	leftWingMesh.SetPosition(-0.9, -0.3, 0) // Ali spostate indietro per bilanciamento
	rightWingMesh.SetPosition(0.9, -0.3, 0)

	// Cannoni: piccoli cilindri allineati sotto le ali
	cannonGeom := geometry.NewCylinder(0.05, 0.05, 1, 16, true, true)
	cannonMat := material.NewStandard(math32.NewColor("Gray"))
	leftCannonMesh := graphic.NewMesh(cannonGeom, cannonMat)
	rightCannonMesh := graphic.NewMesh(cannonGeom, cannonMat)

	// Posizionamento cannoni
	leftCannonMesh.SetRotationZ(math32.Pi / 2)
	rightCannonMesh.SetRotationZ(math32.Pi / 2)
	leftCannonMesh.SetPosition(-0.7, -0.2, 0.5)
	rightCannonMesh.SetPosition(0.7, -0.2, 0.5)

	// Motori posteriori: cilindri inclinati per un look più dinamico
	engineGeom := geometry.NewCylinder(0.15, 0.1, 1, 16, true, true)
	engineMat := material.NewStandard(math32.NewColor("DarkGray"))
	leftEngineMesh := graphic.NewMesh(engineGeom, engineMat)
	rightEngineMesh := graphic.NewMesh(engineGeom, engineMat)

	// Posizionamento motori
	leftEngineMesh.SetRotationX(-math32.DegToRad(45))
	rightEngineMesh.SetRotationX(-math32.DegToRad(45))
	leftEngineMesh.SetPosition(-0.4, 0, -1.2)
	rightEngineMesh.SetPosition(0.4, 0, -1.2)

	// Dettagli aggiuntivi: antenne e componenti di sensoristica
	antennaGeom := geometry.NewCylinder(0.02, 0.02, 1, 8, true, true)
	antennaMat := material.NewStandard(math32.NewColor("Black"))
	antennaMesh := graphic.NewMesh(antennaGeom, antennaMat)
	antennaMesh.SetRotationZ(math32.Pi / 2)
	antennaMesh.SetPosition(0, 0.5, -0.8)

	// Assemblaggio della navicella
	bodyMesh.Add(cockpitMesh)
	bodyMesh.Add(leftWingMesh)
	bodyMesh.Add(rightWingMesh)
	bodyMesh.Add(leftCannonMesh)
	bodyMesh.Add(rightCannonMesh)
	bodyMesh.Add(leftEngineMesh)
	bodyMesh.Add(rightEngineMesh)
	bodyMesh.Add(antennaMesh)

	s.Mass = 10
	s.Ship = &bodyMesh.Node
}

func (s *SpaceShip) LoadShip() {
	// Decodes obj file and associated mtl file
	dec, err := obj.Decode("./assets/3d/spaceship/SpaceShip.obj", "./assets/3d/spaceship/SpaceShip.mtl")
	if err != nil {
		log.Panic(err)
	}

	// Creates a new node with all the objects in the decoded file and adds it to the scene
	group, err := dec.NewGroup()
	if err != nil {
		log.Panic(err)
	}

	// material := material.NewStandard(math32.NewColor("red"))

	// for _, child := range s.Ship.Children() {

	// 	if mesh, ok := child.(*graphic.Mesh); ok {
	// 		mesh.SetMaterial(material)
	// 	}
	// }

	s.Ship = group
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
