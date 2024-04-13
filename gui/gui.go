package gui

import (
	"github.com/EliCDavis/vector/vector3"
	"github.com/alexanderi96/go-fluid-simulator/physics"
	"github.com/alexanderi96/go-fluid-simulator/utils"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func Draw(s *physics.Simulation) {
	rl.BeginDrawing()
	rl.ClearBackground(rl.DarkGray)

	rl.BeginMode3D(s.Camera)

	// Disegno di un pavimento a griglia come punto di riferimento
	// rl.DrawGrid(20, 5)

	drawFluid(s)
	rl.DrawSphere(utils.ToRlVector3(s.Octree.CenterOfMass), 0.1, rl.Red)

	// if s.Config.ShowOverlay {
	// 	for _, unit := range s.Fluid {
	// 		drawOverlay(unit)
	// 	}
	// }

	// Calcola i punti più vicini sul cubo di gioco
	xNear, yNear, zNear := calculateNearestCubePoints(s)
	// Disegna le linee tratteggiate
	drawDashedLine(xNear, s.SpawnPosition, 0.1, 0.1)
	drawDashedLine(yNear, s.SpawnPosition, 0.1, 0.1)
	drawDashedLine(zNear, s.SpawnPosition, 0.1, 0.1)

	cubeColor := rl.Red
	if s.IsSpawnInRange() {
		cubeColor = rl.Green
	}
	rl.DrawCube(utils.ToRlVector3(s.SpawnPosition), 1, 1, 1, cubeColor) // Modifica le dimensioni e il colore come preferisci

	if s.MouseButtonPressed && s.InitialMousePosition.X > 0 && s.InitialMousePosition.X < float32(s.Config.WindowWidth-s.Config.SidebarWidth) {
		rl.DrawLineEx(s.InitialMousePosition, s.CurrentMousePosition, 5, rl.Black)
	}

	if s.Config.ShowOctree {
		drawOctree(s.Octree)
	} else {
		rl.DrawBoundingBox(utils.ToRlBoundingBox(s.WorldBoundray.Min, s.WorldBoundray.Max), rl.RayWhite)
	}

	rl.DrawCube(utils.ToRlVector3(s.WorldBoundray.Min), 1, 1, 1, rl.Blue)
	rl.DrawCube(utils.ToRlVector3(s.WorldBoundray.Max), 1, 1, 1, rl.Red)
	rl.DrawLine3D(utils.ToRlVector3(s.WorldBoundray.Min), utils.ToRlVector3(vector3.New(s.WorldBoundray.Max.X(), s.WorldBoundray.Min.Y(), s.WorldBoundray.Min.Z())), rl.Red)
	rl.DrawLine3D(utils.ToRlVector3(s.WorldBoundray.Min), utils.ToRlVector3(vector3.New(s.WorldBoundray.Min.X(), s.WorldBoundray.Max.Y(), s.WorldBoundray.Min.Z())), rl.Green)
	rl.DrawLine3D(utils.ToRlVector3(s.WorldBoundray.Min), utils.ToRlVector3(vector3.New(s.WorldBoundray.Min.X(), s.WorldBoundray.Min.Y(), s.WorldBoundray.Max.Z())), rl.Blue)

	rl.EndMode3D()
	drawSidebar(s)

	rl.DrawFPS(10, 10)
	rl.EndDrawing()

}

func drawFluid(s *physics.Simulation) {
	for _, unit := range s.Fluid {

		color := unit.Color

		if s.Config.ShowSpeedColor {
			color = utils.GetColorFromVelocity(unit.GetVelocity())
		}

		if s.Config.ShowVectors {
			drawVectors(unit)
		}

		rl.DrawSphere(utils.ToRlVector3(unit.Position), float32(unit.Radius), color)
	}
}

func drawOctree(octree *physics.Octree) {
	if octree == nil {
		return // Ritorna se l'Octree è nil
	}

	// Disegna il BoundingBox dell'Octree corrente
	rl.DrawBoundingBox(utils.ToRlBoundingBox(octree.Bounds.Min, octree.Bounds.Max), rl.Black)
	rl.DrawSphere(utils.ToRlVector3(octree.CenterOfMass), 0.2, rl.Red)

	// Disegna ricorsivamente i BoundingBox dei sotto-Octrees
	for _, child := range octree.Children {
		drawOctree(child)
	}
}

func drawVectors(u *physics.Unit) {

	endVelocity := u.Position.Add(u.GetVelocity().Scale(0.1))
	rl.DrawLine3D(utils.ToRlVector3(u.Position), utils.ToRlVector3(endVelocity), rl.Blue)

	endAcceleration := u.Position.Add(u.Acceleration.Scale(0.1))

	rl.DrawLine3D(utils.ToRlVector3(u.Position), utils.ToRlVector3(endAcceleration), rl.Orange)
}

func calculateNearestCubePoints(s *physics.Simulation) (xNear, yNear, zNear vector3.Vector[float64]) {
	if s.SpawnPosition.X() < s.WorldCenter.X() {
		xNear = vector3.New(s.WorldBoundray.Min.X(), s.SpawnPosition.Y(), s.SpawnPosition.Z())
	} else {
		xNear = vector3.New(s.WorldBoundray.Max.X(), s.SpawnPosition.Y(), s.SpawnPosition.Z())
	}
	if s.SpawnPosition.Y() < s.WorldCenter.Y() {
		yNear = vector3.New(s.SpawnPosition.X(), s.WorldBoundray.Min.Y(), s.SpawnPosition.Z())
	} else {
		yNear = vector3.New(s.SpawnPosition.X(), s.WorldBoundray.Max.Y(), s.SpawnPosition.Z())
	}
	if s.SpawnPosition.Z() < s.WorldCenter.Z() {
		zNear = vector3.New(s.SpawnPosition.X(), s.SpawnPosition.Y(), s.WorldBoundray.Min.Z())
	} else {
		zNear = vector3.New(s.SpawnPosition.X(), s.SpawnPosition.Y(), s.WorldBoundray.Max.Z())
	}
	return
}

func drawDashedLine(start, end vector3.Vector[float64], dashesLength, spaceLength float64) {
	direction := end.Sub(start)
	totalLength := direction.Length()
	direction = direction.Normalized()

	for currentLength := float64(0); currentLength < totalLength; {
		nextDashEnd := currentLength + dashesLength
		if nextDashEnd > totalLength {
			nextDashEnd = totalLength
		}

		dashStart := start.Add(direction.Scale(currentLength))
		dashEnd := start.Add(direction.Scale(nextDashEnd))
		rl.DrawLine3D(utils.ToRlVector3(dashStart), utils.ToRlVector3(dashEnd), rl.RayWhite) // Cambia il colore se necessario

		currentLength = nextDashEnd + spaceLength
	}
}
