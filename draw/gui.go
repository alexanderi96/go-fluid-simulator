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
	// Create a panel for mode indicators (top right)
	modePanel := gui.NewPanel(200, 80)
	modePanel.SetBorders(1, 1, 1, 1)
	modePanel.SetBordersColor(math32.NewColor("darkgray"))
	modePanel.SetColor4(&math32.Color4{R: 0.2, G: 0.2, B: 0.2, A: 0.7})
	s.Scene.Add(modePanel)
	s.Hud.ModePanel = modePanel

	// Navigation mode label
	s.Hud.NavigationLabel = gui.NewLabel("MODE: CAMERA")
	s.Hud.NavigationLabel.SetPosition(10, 10)
	s.Hud.NavigationLabel.SetColor(math32.NewColor("lightgreen"))
	modePanel.Add(s.Hud.NavigationLabel)

	// Spaceship status
	s.Hud.ShipStatusLabel = gui.NewLabel("SHIP: NOT PRESENT")
	s.Hud.ShipStatusLabel.SetPosition(10, 30)
	s.Hud.ShipStatusLabel.SetColor(math32.NewColor("yellow"))
	modePanel.Add(s.Hud.ShipStatusLabel)

	// Controls panel (top left)
	controlsPanel := gui.NewPanel(180, 140)
	controlsPanel.SetBorders(1, 1, 1, 1)
	controlsPanel.SetBordersColor(math32.NewColor("darkgray"))
	controlsPanel.SetColor4(&math32.Color4{R: 0.2, G: 0.2, B: 0.2, A: 0.7})
	s.Scene.Add(controlsPanel)
	s.Hud.ControlsPanel = controlsPanel

	// Debug info
	s.Hud.FpsLabel = gui.NewLabel("FPS: 0")
	s.Hud.FpsLabel.SetPosition(10, 10)
	s.Hud.FpsLabel.SetColor(math32.NewColor("white"))
	controlsPanel.Add(s.Hud.FpsLabel)

	s.Hud.FtLabel = gui.NewLabel("FrameTime: 0")
	s.Hud.FtLabel.SetPosition(10, 30)
	s.Hud.FtLabel.SetColor(math32.NewColor("white"))
	controlsPanel.Add(s.Hud.FtLabel)

	s.Hud.UnitLabel = gui.NewLabel("Units: 0")
	s.Hud.UnitLabel.SetPosition(10, 50)
	s.Hud.UnitLabel.SetColor(math32.NewColor("white"))
	controlsPanel.Add(s.Hud.UnitLabel)

	s.Hud.SimDurationLabel = gui.NewLabel("Sim duration: 0")
	s.Hud.SimDurationLabel.SetPosition(10, 70)
	s.Hud.SimDurationLabel.SetColor(math32.NewColor("white"))
	controlsPanel.Add(s.Hud.SimDurationLabel)

	s.Hud.RealDurationLabel = gui.NewLabel("Real duration: 0")
	s.Hud.RealDurationLabel.SetPosition(10, 90)
	s.Hud.RealDurationLabel.SetColor(math32.NewColor("white"))
	controlsPanel.Add(s.Hud.RealDurationLabel)

	// Key controls info (bottom left)
	keysPanel := gui.NewPanel(200, 160)
	keysPanel.SetBorders(1, 1, 1, 1)
	keysPanel.SetBordersColor(math32.NewColor("darkgray"))
	keysPanel.SetColor4(&math32.Color4{R: 0.2, G: 0.2, B: 0.2, A: 0.7})
	s.Scene.Add(keysPanel)
	s.Hud.KeysPanel = keysPanel

	keyTitle := gui.NewLabel("CONTROLS")
	keyTitle.SetPosition(10, 10)
	keyTitle.SetColor(math32.NewColor("orange"))
	keysPanel.Add(keyTitle)

	keyControls := []string{
		"F - Toggle Flight Mode",
		"W/S - Thrust Control",
		"A/D - Roll Left/Right",
		"Q/E - Yaw Left/Right",
		"K/M - Pitch Up/Down",
		"SPACE - Brake",
	}

	for i, control := range keyControls {
		label := gui.NewLabel(control)
		label.SetPosition(10, float32(35+i*20))
		label.SetColor(math32.NewColor("lightblue"))
		keysPanel.Add(label)
	}

	// Ship info panel (bottom center)
	shipPanel := gui.NewPanel(300, 140)
	shipPanel.SetBorders(1, 1, 1, 1)
	shipPanel.SetBordersColor(math32.NewColor("darkgray"))
	shipPanel.SetColor4(&math32.Color4{R: 0.2, G: 0.2, B: 0.2, A: 0.7})
	s.Scene.Add(shipPanel)
	s.Hud.ShipPanel = shipPanel

	s.Hud.PositionLabel = gui.NewLabel("")
	s.Hud.PositionLabel.SetPosition(10, 10)
	s.Hud.PositionLabel.SetColor(math32.NewColor("white"))
	shipPanel.Add(s.Hud.PositionLabel)

	s.Hud.SpeedLabel = gui.NewLabel("")
	s.Hud.SpeedLabel.SetPosition(10, 30)
	s.Hud.SpeedLabel.SetColor(math32.NewColor("white"))
	shipPanel.Add(s.Hud.SpeedLabel)

	s.Hud.DirectionLabel = gui.NewLabel("")
	s.Hud.DirectionLabel.SetPosition(10, 50)
	s.Hud.DirectionLabel.SetColor(math32.NewColor("white"))
	shipPanel.Add(s.Hud.DirectionLabel)

	s.Hud.OrientationLabel = gui.NewLabel("")
	s.Hud.OrientationLabel.SetPosition(10, 70)
	s.Hud.OrientationLabel.SetColor(math32.NewColor("white"))
	shipPanel.Add(s.Hud.OrientationLabel)

	s.Hud.StatusLabel = gui.NewLabel("")
	s.Hud.StatusLabel.SetPosition(10, 90)
	s.Hud.StatusLabel.SetColor(math32.NewColor("white"))
	shipPanel.Add(s.Hud.StatusLabel)
}

func UpdateHUD(s *physics.Simulation, deltaTime time.Duration) {
	// Update panel positions based on current window size
	w, h := s.App.GetSize()
	width, height := float32(w), float32(h)

	// Update panel positions
	s.Hud.ModePanel.SetPosition(width-210, 10)
	s.Hud.ControlsPanel.SetPosition(10, 10)
	s.Hud.KeysPanel.SetPosition(10, height-170)
	s.Hud.ShipPanel.SetPosition(width/2-150, height-150)

	fps := 1.0 / float64(deltaTime.Seconds())
	s.Hud.FpsLabel.SetText(fmt.Sprintf("FPS: %.0f", fps))
	s.Hud.FtLabel.SetText(fmt.Sprintf("FrameTime: %.2f", s.Config.Frametime))
	s.Hud.UnitLabel.SetText(fmt.Sprintf("Units: %d", len(s.Fluid)))
	s.Hud.SimDurationLabel.SetText(fmt.Sprintf("Sim duration: %.1f", s.Metrics.SimDuration))
	s.Hud.RealDurationLabel.SetText(fmt.Sprintf("Real duration: %.1f", -time.Until(s.AppStartTime).Seconds()))

	// Update navigation mode and ship status
	if s.Fly {
		s.Hud.NavigationLabel.SetText("MODE: FLIGHT")
		s.Hud.NavigationLabel.SetColor(math32.NewColor("lightgreen"))
	} else {
		s.Hud.NavigationLabel.SetText("MODE: CAMERA")
		s.Hud.NavigationLabel.SetColor(math32.NewColor("skyblue"))
	}

	if s.SpaceShip != nil {
		s.Hud.ShipStatusLabel.SetText("SHIP: ACTIVE")
		s.Hud.ShipStatusLabel.SetColor(math32.NewColor("lightgreen"))

		// Update ship information
		pos := s.SpaceShip.Ship.Position()
		s.Hud.PositionLabel.SetText(fmt.Sprintf("Position: X:%.1f Y:%.1f Z:%.1f", pos.X, pos.Y, pos.Z))
		s.Hud.SpeedLabel.SetText(fmt.Sprintf("Speed: %.1f units/s", s.SpaceShip.Speed))

		// Calculate and update direction
		forward := math32.NewVector3(0, 0, 1)
		matrix := s.SpaceShip.Ship.Matrix()
		forward.ApplyMatrix4(&matrix)
		forward.Normalize()

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

		// Update orientation
		rot := s.SpaceShip.Ship.Rotation()
		s.Hud.OrientationLabel.SetText(fmt.Sprintf("Orientation - P:%.1f° R:%.1f° Y:%.1f°",
			math32.RadToDeg(rot.X),
			math32.RadToDeg(rot.Z),
			math32.RadToDeg(rot.Y)))

		// Update status
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
	} else {
		s.Hud.ShipStatusLabel.SetText("SHIP: NOT PRESENT")
		s.Hud.ShipStatusLabel.SetColor(math32.NewColor("yellow"))

		// Clear ship information when no ship is present
		s.Hud.PositionLabel.SetText("")
		s.Hud.SpeedLabel.SetText("")
		s.Hud.DirectionLabel.SetText("")
		s.Hud.OrientationLabel.SetText("")
		s.Hud.StatusLabel.SetText("")
	}
}
