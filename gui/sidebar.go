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
	xContentStart      int32
	xMenuStart         int32
	yContentStartTop   int32
	yMenuStartTop      int32
	sliderLength       float32
	sliderThickness                   = float32(20)
	currentSection     SidebarSection = DebugSection // Inizializza a un valore o recupera da qualche parte
	sidebarOpen                       = false
	sideMenuTotalWidth int32

	sideMenuContentWidth = int32(250)
	menuButtonWidth      = int32(50)
)

func drawSidebar(s *physics.Simulation) error {

	if sidebarOpen {
		sideMenuTotalWidth = menuButtonWidth + sideMenuContentWidth + 3*border
	} else {
		sideMenuTotalWidth = menuButtonWidth + 2*border
	}

	s.Config.ResizeViewport(-sideMenuTotalWidth, 0)

	yContentStartTop = border // or whatever value you have for 'border'
	yMenuStartTop = border
	xMenuStart = s.Config.ViewportX + border // or whatever value you have for 'border'
	xContentStart = xMenuStart + menuButtonWidth + 2*border

	sliderLength = float32(sideMenuContentWidth - 2*10) // or whatever value you have for 'border'

	rl.DrawRectangle(s.Config.WindowWidth-sideMenuTotalWidth, 0, s.Config.WindowWidth, s.Config.WindowHeight, rl.RayWhite)

	// Chiamata alla funzione per disegnare il selettore di sezione
	drawSectionSelector(xMenuStart, yMenuStartTop, &currentSection)

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

func drawSectionSelector(xMenuStart, yMenuStartTop int32, currentSection *SidebarSection) {
	// Creiamo un elenco dei nomi delle sezioni per i pulsanti
	sectionNames := []string{"Debug", "Display", "Game"}

	if sidebarOpen {
		sectionNames = append(sectionNames, ">>")
	} else {
		sectionNames = append(sectionNames, "<<")
	}

	for i, sectionName := range sectionNames {
		// Posizione per disegnare il pulsante
		buttonRect := rl.Rectangle{X: float32(xMenuStart), Y: float32(yMenuStartTop + int32(i)*35), Width: float32(menuButtonWidth), Height: 30}

		// Controlla se il pulsante viene premuto
		if gui.Button(buttonRect, sectionName) {
			if sectionName == ">>" || sectionName == "<<" {
				// Cambia lo stato della sidebar quando il pulsante "Toggle Sidebar" viene premuto
				sidebarOpen = !sidebarOpen
			} else {
				// Altrimenti, imposta la currentSection in base al pulsante premuto
				*currentSection = SidebarSection(i)
				sidebarOpen = true
			}
		}
	}
}

func drawDebugSection(s *physics.Simulation) {
	s.Config.ShowVectors = gui.CheckBox(rl.Rectangle{X: float32(xContentStart), Y: float32(yContentStartTop), Width: 20, Height: 20}, "Show Vectors", s.Config.ShowVectors)
	yContentStartTop += 20 + 5

	s.Config.ShowQuadtree = gui.CheckBox(rl.Rectangle{X: float32(xContentStart), Y: float32(yContentStartTop), Width: 20, Height: 20}, "Show Quadtree", s.Config.ShowQuadtree)
	yContentStartTop += 20 + 5

	s.Config.ShowTrail = gui.CheckBox(rl.Rectangle{X: float32(xContentStart), Y: float32(yContentStartTop), Width: 20, Height: 20}, "Show Trail", s.Config.ShowTrail)
	yContentStartTop += 20 + 5

	s.Config.ShouldBeProfiled = gui.CheckBox(rl.Rectangle{X: float32(xContentStart), Y: float32(yContentStartTop), Width: 20, Height: 20}, "Should be profiled", s.Config.ShouldBeProfiled)
	yContentStartTop += 20 + 5

	quadtree := fmt.Sprintf("Using qTree: %t", s.Config.UseExperimentalQuadtree)
	rl.DrawText(quadtree, xContentStart, yContentStartTop, 20, rl.Black)
	yContentStartTop += 40 + 5

	frametime := fmt.Sprintf("Frametime: %f", s.Metrics.Frametime)
	rl.DrawText(frametime, xContentStart, yContentStartTop, 20, rl.Black)
	yContentStartTop += 20 + 5

	fps := fmt.Sprintf("FPS: %d", s.Metrics.FPS)
	rl.DrawText(fps, xContentStart, yContentStartTop, 20, rl.Black)
	yContentStartTop += 20 + 5

	heapSize := fmt.Sprintf("Heap Size: %d kb", s.Metrics.HeapSize)
	rl.DrawText(heapSize, xContentStart, yContentStartTop, 20, rl.Black)
	yContentStartTop += 20 + 5
}

func drawDisplayConfigSection(s *physics.Simulation) {
	s.Config.FullScreen = gui.CheckBox(rl.Rectangle{X: float32(xContentStart), Y: float32(yContentStartTop), Width: 20, Height: 20}, "Fullscreen", s.Config.FullScreen)
	yContentStartTop += 20 + 5

	windowSizes := fmt.Sprintf("Window size: %dx%d", s.Config.WindowWidth, s.Config.WindowHeight)
	rl.DrawText(windowSizes, xContentStart, yContentStartTop, 20, rl.Black)
	yContentStartTop += 40 + 5

	sidebarWidth := fmt.Sprintf("Window size: %d", sideMenuTotalWidth)
	rl.DrawText(sidebarWidth, xContentStart, yContentStartTop, 20, rl.Black)
	yContentStartTop += 40 + 5

	viewportSizes := fmt.Sprintf("Viewport size: %dx%d", s.Config.ViewportX, s.Config.ViewportY)
	rl.DrawText(viewportSizes, xContentStart, yContentStartTop, 20, rl.Black)
	yContentStartTop += 40 + 5

	TargetFPS := fmt.Sprintf("Target FPS: %dx%d", s.Config.ViewportX, s.Config.TargetFPS)
	rl.DrawText(TargetFPS, xContentStart, yContentStartTop, 20, rl.Black)
	yContentStartTop += 40 + 5

	s.Config.IsResizable = gui.CheckBox(rl.Rectangle{X: float32(xContentStart), Y: float32(yContentStartTop), Width: 20, Height: 20}, "Is Resizable", s.Config.IsResizable)
	yContentStartTop += 20 + 5

	s.Config.ShowOverlay = gui.CheckBox(rl.Rectangle{X: float32(xContentStart), Y: float32(yContentStartTop), Width: 20, Height: 20}, "Show Overlay", s.Config.ShowOverlay)
	yContentStartTop += 20 + 5
}

func drawGameConfigSection(s *physics.Simulation) {
	gameSizes := fmt.Sprintf("Game size: %.1fx%.1fx%.1f", s.Config.GameX, s.Config.GameY, s.Config.GameZ)
	rl.DrawText(gameSizes, xContentStart, yContentStartTop, 20, rl.Black)
	yContentStartTop += 40 + 5

	unitNumbers := fmt.Sprintf("Spawned Units: %d", len(s.Fluid))
	rl.DrawText(unitNumbers, xContentStart, yContentStartTop, 20, rl.Black)
	yContentStartTop += 20 + 5

	unitsToBeSpawned := fmt.Sprintf("Units  To Be Spawned: %d", s.Config.UnitNumber)
	rl.DrawText(unitsToBeSpawned, xContentStart, yContentStartTop, 20, rl.Black)
	yContentStartTop += 20 + 5

	s.Config.UnitNumber = int32(gui.Slider(rl.Rectangle{X: float32(xContentStart), Y: float32(yContentStartTop), Width: sliderLength, Height: sliderThickness}, "", "", float32(s.Config.UnitNumber), 1, 1000))
	yContentStartTop += 20 + 5

	s.Config.ApplyGravity = gui.CheckBox(rl.Rectangle{X: float32(xContentStart), Y: float32(yContentStartTop), Width: 20, Height: 20}, "Apply Gravity", s.Config.ApplyGravity)
	yContentStartTop += 20 + 5

	s.Config.SetRandomColor = gui.CheckBox(rl.Rectangle{X: float32(xContentStart), Y: float32(yContentStartTop), Width: 20, Height: 20}, "Set Random Color", s.Config.SetRandomColor)
	yContentStartTop += 20 + 5

	s.Config.ShowSpeedColor = gui.CheckBox(rl.Rectangle{X: float32(xContentStart), Y: float32(yContentStartTop), Width: 20, Height: 20}, "Show Speed Color", s.Config.ShowSpeedColor)
	yContentStartTop += 20 + 5

	s.Config.UnitsEmitGravity = gui.CheckBox(rl.Rectangle{X: float32(xContentStart), Y: float32(yContentStartTop), Width: 20, Height: 20}, "Unit Emit Gravity", s.Config.UnitsEmitGravity)
	yContentStartTop += 20 + 5

	s.Config.SetRandomRadius = gui.CheckBox(rl.Rectangle{X: float32(xContentStart), Y: float32(yContentStartTop), Width: 20, Height: 20}, "Set Random Radius", s.Config.SetRandomRadius)
	yContentStartTop += 20 + 5

	if s.Config.SetRandomRadius {
		radMinText := strconv.FormatFloat(float64(s.Config.RadiusMin), 'f', 2, 64)

		radMinInput := rl.Rectangle{X: float32(xContentStart), Y: float32(yContentStartTop), Width: 200, Height: 30}

		radMinEdited := gui.TextBox(radMinInput, &radMinText, 10, true)
		yContentStartTop += 30 + 5

		if radMinEdited {

			if radMin, err := utils.CheckTextFloat32(radMinText); err == nil {
				s.Config.RadiusMin = radMin
			}
		}

		radMaxText := strconv.FormatFloat(float64(s.Config.RadiusMax), 'f', 2, 64)

		radMaxInput := rl.Rectangle{X: float32(xContentStart), Y: float32(yContentStartTop), Width: 200, Height: 30}

		radMaxEdited := gui.TextBox(radMaxInput, &radMaxText, 10, true)
		yContentStartTop += 30 + 5

		if radMaxEdited {

			if radMax, err := utils.CheckTextFloat32(radMaxText); err == nil {
				s.Config.RadiusMax = radMax
			}
		}

	}

	s.Config.SetRandomElasticity = gui.CheckBox(rl.Rectangle{X: float32(xContentStart), Y: float32(yContentStartTop), Width: 20, Height: 20}, "Set Random Elasticity", s.Config.SetRandomElasticity)
	yContentStartTop += 20 + 5

	if s.Config.SetRandomElasticity {
		elasticityMinText := strconv.FormatFloat(float64(s.Config.ElasticityMin), 'f', 2, 64)

		elasticityMinInput := rl.Rectangle{X: float32(xContentStart), Y: float32(yContentStartTop), Width: 200, Height: 30}

		elasticityMinEdited := gui.TextBox(elasticityMinInput, &elasticityMinText, 10, true)
		yContentStartTop += 30 + 5

		if elasticityMinEdited {

			if elasticityMin, err := utils.CheckTextFloat32(elasticityMinText); err == nil {
				s.Config.ElasticityMin = elasticityMin
			}
		}

		elasticityMaxText := strconv.FormatFloat(float64(s.Config.ElasticityMax), 'f', 2, 64)

		elasticityMaxInput := rl.Rectangle{X: float32(xContentStart), Y: float32(yContentStartTop), Width: 200, Height: 30}

		elasticityMaxEdited := gui.TextBox(elasticityMaxInput, &elasticityMaxText, 10, true)
		yContentStartTop += 30 + 5

		if elasticityMaxEdited {

			if elasticityMax, err := utils.CheckTextFloat32(elasticityMaxText); err == nil {
				s.Config.ElasticityMax = elasticityMax
			}
		}

	}

	s.Config.SetRandomMassMultiplier = gui.CheckBox(rl.Rectangle{X: float32(xContentStart), Y: float32(yContentStartTop), Width: 20, Height: 20}, "Set Random MassMultiplier", s.Config.SetRandomMassMultiplier)
	yContentStartTop += 20 + 5

	if s.Config.SetRandomMassMultiplier {
		massMinText := strconv.FormatFloat(float64(s.Config.MassMultiplierMin), 'f', 2, 64)

		MassMultiplierMinInput := rl.Rectangle{X: float32(xContentStart), Y: float32(yContentStartTop), Width: 200, Height: 30}

		MassMultiplierMinEdited := gui.TextBox(MassMultiplierMinInput, &massMinText, 10, true)
		yContentStartTop += 30 + 5

		if MassMultiplierMinEdited {
			if MassMultiplierMin, err := utils.CheckTextFloat32(massMinText); err == nil {
				s.Config.MassMultiplierMin = MassMultiplierMin
			}
		}

		massMaxText := strconv.FormatFloat(float64(s.Config.MassMultiplierMax), 'f', 2, 64)

		MassMultiplierMaxInput := rl.Rectangle{X: float32(xContentStart), Y: float32(yContentStartTop), Width: 200, Height: 30}

		MassMultiplierMaxEdited := gui.TextBox(MassMultiplierMaxInput, &massMaxText, 10, true)
		yContentStartTop += 30 + 5

		if MassMultiplierMaxEdited {
			if MassMultiplierMax, err := utils.CheckTextFloat32(massMaxText); err == nil {
				s.Config.MassMultiplierMax = MassMultiplierMax
			}
		}

	}

}
