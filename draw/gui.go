package draw

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/alexanderi96/go-fluid-simulator/physics"
	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/math32"
)

func SetupHUD(s *physics.Simulation) {
	// Debug info (in alto a sinistra)
	s.Hud.FpsLabel = gui.NewLabel("FPS: 0")
	s.Hud.FpsLabel.SetPosition(10, 10)
	s.Scene.Add(s.Hud.FpsLabel)

	s.Hud.FtLabel = gui.NewLabel("FrameTime: 0")
	s.Hud.FtLabel.SetPosition(100, 10)
	s.Scene.Add(s.Hud.FtLabel)

	s.Hud.UnitLabel = gui.NewLabel("units: 0")
	s.Hud.UnitLabel.SetPosition(10, 25)
	s.Scene.Add(s.Hud.UnitLabel)

	s.Hud.SimDurationLabel = gui.NewLabel("Simulation duration: 0")
	s.Hud.SimDurationLabel.SetPosition(10, 40)
	s.Scene.Add(s.Hud.SimDurationLabel)

	s.Hud.RealDurationLabel = gui.NewLabel("Real duration: 0")
	s.Hud.RealDurationLabel.SetPosition(10, 55)
	s.Scene.Add(s.Hud.RealDurationLabel)

	// Info navicella (in basso al centro)
	// Calcola la posizione centrale dello schermo
	width := float32(800)  // Sostituisci con la larghezza effettiva della finestra
	height := float32(600) // Sostituisci con l'altezza effettiva della finestra
	centerX := width / 2
	bottomY := height - 120 // Spazio dal fondo dello schermo

	// Posiziona le etichette della navicella centrate in basso
	s.Hud.PositionLabel = gui.NewLabel("")
	s.Hud.PositionLabel.SetPosition(centerX-100, bottomY)
	s.Scene.Add(s.Hud.PositionLabel)

	s.Hud.SpeedLabel = gui.NewLabel("")
	s.Hud.SpeedLabel.SetPosition(centerX-100, bottomY+20)
	s.Scene.Add(s.Hud.SpeedLabel)

	s.Hud.DirectionLabel = gui.NewLabel("")
	s.Hud.DirectionLabel.SetPosition(centerX-100, bottomY+40)
	s.Scene.Add(s.Hud.DirectionLabel)

	s.Hud.OrientationLabel = gui.NewLabel("")
	s.Hud.OrientationLabel.SetPosition(centerX-100, bottomY+60)
	s.Scene.Add(s.Hud.OrientationLabel)

	s.Hud.StatusLabel = gui.NewLabel("")
	s.Hud.StatusLabel.SetPosition(centerX-100, bottomY+80)
	s.Scene.Add(s.Hud.StatusLabel)
}

func UpdateHUD(s *physics.Simulation, deltaTime time.Duration) {
	fps := 1.0 / float64(deltaTime.Seconds())
	s.Hud.FpsLabel.SetText("FPS: " + fmt.Sprintf("%.2f", fps))

	s.Hud.FtLabel.SetText("FrameTime: " + fmt.Sprintf("%.2f", s.Config.Frametime))

	s.Hud.UnitLabel.SetText("unit: " + fmt.Sprintf("%d", len(s.Fluid)))

	s.Hud.SimDurationLabel.SetText("Simulation duration: " + fmt.Sprintf("%.2f", s.Metrics.SimDuration))

	s.Hud.RealDurationLabel.SetText("Real duration: " + fmt.Sprintf("%.2f", -time.Until(s.AppStartTime).Seconds()))

	// Aggiorna posizione
	pos := s.SpaceShip.Ship.Position()
	s.Hud.PositionLabel.SetText(fmt.Sprintf("Position: X: %.1f Y: %.1f Z: %.1f", pos.X, pos.Y, pos.Z))

	// Aggiorna velocità
	s.Hud.SpeedLabel.SetText(fmt.Sprintf("Speed: %.1f units/s", s.SpaceShip.Speed))

	// Calcola e aggiorna la direzione
	forward := math32.NewVector3(0, 0, 1)
	matrix := s.SpaceShip.Ship.Matrix()
	forward.ApplyMatrix4(&matrix)
	forward.Normalize()

	// Determina le direzioni cardinali
	directions := []string{}
	if forward.Z > 0.3 {
		directions = append(directions, "North")
	}
	if forward.Z < -0.3 {
		directions = append(directions, "South")
	}
	if forward.X > 0.3 {
		directions = append(directions, "East")
	}
	if forward.X < -0.3 {
		directions = append(directions, "West")
	}
	if forward.Y > 0.3 {
		directions = append(directions, "Up")
	}
	if forward.Y < -0.3 {
		directions = append(directions, "Down")
	}
	directionText := strings.Join(directions, "-")
	if directionText == "" {
		directionText = "Neutral"
	}
	s.Hud.DirectionLabel.SetText(fmt.Sprintf("Direction: %s", directionText))

	// Calcola e aggiorna l'orientamento in gradi
	rot := s.SpaceShip.Ship.Rotation()
	s.Hud.OrientationLabel.SetText(fmt.Sprintf("Orientation - Pitch: %.1f° Roll: %.1f° Yaw: %.1f°",
		math32.RadToDeg(rot.X),
		math32.RadToDeg(rot.Z),
		math32.RadToDeg(rot.Y)))

	// Aggiornamento status
	var status []string
	if math.Abs(float64(s.SpaceShip.Speed)) < 0.001 {
		status = append(status, "HOVERING")
	} else if s.SpaceShip.Speed > 0 {
		status = append(status, "FORWARD")
	} else {
		status = append(status, "REVERSE")
	}
	if math.Abs(float64(rot.Z)) > 0.1 {
		if rot.Z > 0 {
			status = append(status, "ROLLING RIGHT")
		} else {
			status = append(status, "ROLLING LEFT")
		}
	}
	if math.Abs(float64(rot.X)) > 0.1 {
		if rot.X > 0 {
			status = append(status, "PITCHING UP")
		} else {
			status = append(status, "PITCHING DOWN")
		}
	}
	statusText := strings.Join(status, " | ")
	s.Hud.StatusLabel.SetText(fmt.Sprintf("Status: %s", statusText))
}
