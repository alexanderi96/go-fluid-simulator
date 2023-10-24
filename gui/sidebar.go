package gui

import (
	"fmt"
	"strconv"

	"github.com/alexanderi96/go-fluid-simulator/physics"
	"github.com/alexanderi96/go-fluid-simulator/utils"

	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type SidebarSection int

const (
	DebugSection SidebarSection = iota
	DisplayConfigSection
	GameConfigSection
)

var (
	border             = int32(10)
	xStart             int32
	yStartTop          int32
	sliderLength       float32
	sliderThickness                   = float32(20)
	currentSection     SidebarSection = DebugSection // Inizializza a un valore o recupera da qualche parte
	sidebarOpen                       = true
	sideMenuTotalWidth int32

	sideMenuContentWidth = int32(200)
	menuButtonWidth      = int32(10)
)

func drawSidebar(s *physics.Simulation) error {

	if sidebarOpen {
		sideMenuTotalWidth = menuButtonWidth + sideMenuContentWidth
	} else {
		sideMenuTotalWidth = menuButtonWidth // o qualsiasi valore per la visualizzazione minimizzata
	}

	yStartTop = border                                                                // or whatever value you have for 'border'
	xStart = s.Config.WindowWidth - (sideMenuContentWidth - menuButtonWidth) + border // or whatever value you have for 'border'
	sliderLength = float32(sideMenuTotalWidth - 2*10)                                 // or whatever value you have for 'border'

	rl.DrawRectangle(s.Config.WindowWidth-sideMenuTotalWidth, 0, s.Config.WindowWidth, s.Config.WindowHeight, rl.RayWhite)

	// Chiamata alla funzione per disegnare il selettore di sezione
	drawSectionSelector(xStart, yStartTop, &currentSection)

	yStartTop += 40 // Vai alla riga successiva

	// Ora disegna la sezione corrispondente
	switch currentSection {
	case DebugSection:
		drawDebugSection(s)
	case DisplayConfigSection:
		drawDisplayConfigSection(s)
	case GameConfigSection:
		drawGameConfigSection(s)
	}

	return nil

}

func drawSectionSelector(xStart, yStartTop int32, currentSection *SidebarSection) {
	sections := "Debug;Display;Game"

	selected := gui.ComboBox(rl.Rectangle{X: float32(xStart), Y: float32(yStartTop), Width: 200, Height: 30}, sections, int32(*currentSection))
	*currentSection = SidebarSection(selected)
}

func drawDebugSection(s *physics.Simulation) {
	s.Config.ShowVectors = gui.CheckBox(rl.Rectangle{X: float32(xStart), Y: float32(yStartTop), Width: 20, Height: 20}, "Show Vectors", s.Config.ShowVectors)
	yStartTop += 20 + 5

	s.Config.ShowQuadtree = gui.CheckBox(rl.Rectangle{X: float32(xStart), Y: float32(yStartTop), Width: 20, Height: 20}, "Show Quadtree", s.Config.ShowQuadtree)
	yStartTop += 20 + 5

	s.Config.ShowTrail = gui.CheckBox(rl.Rectangle{X: float32(xStart), Y: float32(yStartTop), Width: 20, Height: 20}, "Show Trail", s.Config.ShowTrail)
	yStartTop += 20 + 5

	s.Config.ShouldBeProfiled = gui.CheckBox(rl.Rectangle{X: float32(xStart), Y: float32(yStartTop), Width: 20, Height: 20}, "Should be profiled", s.Config.ShouldBeProfiled)
	yStartTop += 20 + 5

	quadtree := fmt.Sprintf("Using qTree: %t", s.Config.UseExperimentalQuadtree)
	rl.DrawText(quadtree, xStart, yStartTop, 20, rl.Black)
	yStartTop += 40 + 5

	frametime := fmt.Sprintf("Frametime: %f", s.Metrics.Frametime)
	rl.DrawText(frametime, xStart, yStartTop, 20, rl.Black)
	yStartTop += 20 + 5

	fps := fmt.Sprintf("FPS: %d", s.Metrics.FPS)
	rl.DrawText(fps, xStart, yStartTop, 20, rl.Black)
	yStartTop += 20 + 5

	heapSize := fmt.Sprintf("Heap Size: %d kb", s.Metrics.HeapSize)
	rl.DrawText(heapSize, xStart, yStartTop, 20, rl.Black)
	yStartTop += 20 + 5
}

func drawDisplayConfigSection(s *physics.Simulation) {
	s.Config.FullScreen = gui.CheckBox(rl.Rectangle{X: float32(xStart), Y: float32(yStartTop), Width: 20, Height: 20}, "Fullscreen", s.Config.FullScreen)
	yStartTop += 20 + 5

	windowSizes := fmt.Sprintf("Window size: %dx%d", s.Config.WindowWidth, s.Config.WindowHeight)
	rl.DrawText(windowSizes, xStart, yStartTop, 20, rl.Black)
	yStartTop += 40 + 5

	sidebarWidth := fmt.Sprintf("Window size: %d", sideMenuTotalWidth)
	rl.DrawText(sidebarWidth, xStart, yStartTop, 20, rl.Black)
	yStartTop += 40 + 5

	viewportSizes := fmt.Sprintf("Viewport size: %dx%d", s.Config.ViewportX, s.Config.ViewportY)
	rl.DrawText(viewportSizes, xStart, yStartTop, 20, rl.Black)
	yStartTop += 40 + 5

	TargetFPS := fmt.Sprintf("Target FPS: %dx%d", s.Config.ViewportX, s.Config.TargetFPS)
	rl.DrawText(TargetFPS, xStart, yStartTop, 20, rl.Black)
	yStartTop += 40 + 5

	s.Config.IsResizable = gui.CheckBox(rl.Rectangle{X: float32(xStart), Y: float32(yStartTop), Width: 20, Height: 20}, "Is Resizable", s.Config.IsResizable)
	yStartTop += 20 + 5

	s.Config.ShowOverlay = gui.CheckBox(rl.Rectangle{X: float32(xStart), Y: float32(yStartTop), Width: 20, Height: 20}, "Show Overlay", s.Config.ShowOverlay)
	yStartTop += 20 + 5
}

func drawGameConfigSection(s *physics.Simulation) {
	gameSizes := fmt.Sprintf("Game size: %dx%dx%d", s.Config.GameX, s.Config.GameY, s.Config.GameZ)
	rl.DrawText(gameSizes, xStart, yStartTop, 20, rl.Black)
	yStartTop += 40 + 5

	unitNumbers := fmt.Sprintf("Spawned Units: %d", len(s.Fluid))
	rl.DrawText(unitNumbers, xStart, yStartTop, 20, rl.Black)
	yStartTop += 20 + 5

	unitsToBeSpawned := fmt.Sprintf("Units  To Be Spawned: %d", s.Config.UnitNumber)
	rl.DrawText(unitsToBeSpawned, xStart, yStartTop, 20, rl.Black)
	yStartTop += 20 + 5

	s.Config.UnitNumber = int32(gui.Slider(rl.Rectangle{X: float32(xStart), Y: float32(yStartTop), Width: sliderLength, Height: sliderThickness}, "", "", float32(s.Config.UnitNumber), 1, 1000))
	yStartTop += 20 + 5

	s.Config.ApplyGravity = gui.CheckBox(rl.Rectangle{X: float32(xStart), Y: float32(yStartTop), Width: 20, Height: 20}, "Apply Gravity", s.Config.ApplyGravity)
	yStartTop += 20 + 5

	s.Config.SetRandomColor = gui.CheckBox(rl.Rectangle{X: float32(xStart), Y: float32(yStartTop), Width: 20, Height: 20}, "Set Random Color", s.Config.SetRandomColor)
	yStartTop += 20 + 5

	s.Config.ShowSpeedColor = gui.CheckBox(rl.Rectangle{X: float32(xStart), Y: float32(yStartTop), Width: 20, Height: 20}, "Show Speed Color", s.Config.ShowSpeedColor)
	yStartTop += 20 + 5

	s.Config.UnitsEmitGravity = gui.CheckBox(rl.Rectangle{X: float32(xStart), Y: float32(yStartTop), Width: 20, Height: 20}, "Unit Emit Gravity", s.Config.UnitsEmitGravity)
	yStartTop += 20 + 5

	s.Config.SetRandomRadius = gui.CheckBox(rl.Rectangle{X: float32(xStart), Y: float32(yStartTop), Width: 20, Height: 20}, "Set Random Radius", s.Config.SetRandomRadius)
	yStartTop += 20 + 5

	if s.Config.SetRandomRadius {
		radMinText := strconv.FormatFloat(float64(s.Config.RadiusMin), 'f', 2, 64)

		radMinInput := rl.Rectangle{X: float32(xStart), Y: float32(yStartTop), Width: 200, Height: 30}

		radMinEdited := gui.TextBox(radMinInput, &radMinText, 10, true)
		yStartTop += 30 + 5

		if radMinEdited {

			if radMin, err := utils.CheckTextFloat32(radMinText); err == nil {
				s.Config.RadiusMin = radMin
			}
		}

		radMaxText := strconv.FormatFloat(float64(s.Config.RadiusMax), 'f', 2, 64)

		radMaxInput := rl.Rectangle{X: float32(xStart), Y: float32(yStartTop), Width: 200, Height: 30}

		radMaxEdited := gui.TextBox(radMaxInput, &radMaxText, 10, true)
		yStartTop += 30 + 5

		if radMaxEdited {

			if radMax, err := utils.CheckTextFloat32(radMaxText); err == nil {
				s.Config.RadiusMax = radMax
			}
		}

	}

	s.Config.SetRandomElasticity = gui.CheckBox(rl.Rectangle{X: float32(xStart), Y: float32(yStartTop), Width: 20, Height: 20}, "Set Random Elasticity", s.Config.SetRandomElasticity)
	yStartTop += 20 + 5

	if s.Config.SetRandomElasticity {
		elasticityMinText := strconv.FormatFloat(float64(s.Config.ElasticityMin), 'f', 2, 64)

		elasticityMinInput := rl.Rectangle{X: float32(xStart), Y: float32(yStartTop), Width: 200, Height: 30}

		elasticityMinEdited := gui.TextBox(elasticityMinInput, &elasticityMinText, 10, true)
		yStartTop += 30 + 5

		if elasticityMinEdited {

			if elasticityMin, err := utils.CheckTextFloat32(elasticityMinText); err == nil {
				s.Config.ElasticityMin = elasticityMin
			}
		}

		elasticityMaxText := strconv.FormatFloat(float64(s.Config.ElasticityMax), 'f', 2, 64)

		elasticityMaxInput := rl.Rectangle{X: float32(xStart), Y: float32(yStartTop), Width: 200, Height: 30}

		elasticityMaxEdited := gui.TextBox(elasticityMaxInput, &elasticityMaxText, 10, true)
		yStartTop += 30 + 5

		if elasticityMaxEdited {

			if elasticityMax, err := utils.CheckTextFloat32(elasticityMaxText); err == nil {
				s.Config.ElasticityMax = elasticityMax
			}
		}

	}

	s.Config.SetRandomMassMultiplier = gui.CheckBox(rl.Rectangle{X: float32(xStart), Y: float32(yStartTop), Width: 20, Height: 20}, "Set Random MassMultiplier", s.Config.SetRandomMassMultiplier)
	yStartTop += 20 + 5

	if s.Config.SetRandomMassMultiplier {
		massMinText := strconv.FormatFloat(float64(s.Config.MassMultiplierMin), 'f', 2, 64)

		MassMultiplierMinInput := rl.Rectangle{X: float32(xStart), Y: float32(yStartTop), Width: 200, Height: 30}

		MassMultiplierMinEdited := gui.TextBox(MassMultiplierMinInput, &massMinText, 10, true)
		yStartTop += 30 + 5

		if MassMultiplierMinEdited {
			if MassMultiplierMin, err := utils.CheckTextFloat32(massMinText); err == nil {
				s.Config.MassMultiplierMin = MassMultiplierMin
			}
		}

		massMaxText := strconv.FormatFloat(float64(s.Config.MassMultiplierMax), 'f', 2, 64)

		MassMultiplierMaxInput := rl.Rectangle{X: float32(xStart), Y: float32(yStartTop), Width: 200, Height: 30}

		MassMultiplierMaxEdited := gui.TextBox(MassMultiplierMaxInput, &massMaxText, 10, true)
		yStartTop += 30 + 5

		if MassMultiplierMaxEdited {
			if MassMultiplierMax, err := utils.CheckTextFloat32(massMaxText); err == nil {
				s.Config.MassMultiplierMax = MassMultiplierMax
			}
		}

	}

}
