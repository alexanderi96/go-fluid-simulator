package gui

import (
	"github.com/alexanderi96/go-fluid-simulator/physics"
	"github.com/alexanderi96/go-fluid-simulator/utils"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func Init(s *physics.Simulation) {

	if s.Config.IsResizable {
		rl.SetConfigFlags(rl.FlagWindowResizable)
	}

	rl.InitWindow(s.Config.WindowWidth, s.Config.WindowHeight, "Go Fluid Simulator")
	rl.SetTargetFPS(s.Config.TargetFPS)

}

func Draw(s *physics.Simulation) {
	rl.BeginDrawing()
	rl.ClearBackground(rl.RayWhite)

	rl.BeginMode3D(s.Camera)

	// Disegno di un pavimento a griglia come punto di riferimento
	rl.DrawGrid(50, 5)

	drawFluid(s)

	// if s.Config.ShowOverlay {
	// 	for _, unit := range s.Fluid {
	// 		drawOverlay(unit)
	// 	}
	// }

	// Calcola i punti più vicini sul cubo di gioco
	xNear, yNear, zNear := calculateNearestCubePoints(s)
	if s.ControlMode == physics.UnitSpawnMode {
		// Disegna le linee tratteggiate
		drawDashedLine(xNear, s.SpawnPosition, 0.1, 0.1)
		drawDashedLine(yNear, s.SpawnPosition, 0.1, 0.1)
		drawDashedLine(zNear, s.SpawnPosition, 0.1, 0.1)
	}

	cubeColor := rl.Red
	if s.IsSpawnInRange() {
		cubeColor = rl.Green
	}
	rl.DrawCube(s.SpawnPosition, 1, 1, 1, cubeColor) // Modifica le dimensioni e il colore come preferisci

	if s.MouseButtonPressed && s.InitialMousePosition.X > 0 && s.InitialMousePosition.X < float32(s.Config.WindowWidth-s.Config.SidebarWidth) {
		rl.DrawLineEx(s.InitialMousePosition, s.CurrentMousePosition, 5, rl.Black)
	}

	if s.Config.ShowOctree {
		drawOctree(s.Octree)
	} else {
		rl.DrawBoundingBox(s.WorldBoundray, rl.Green)
	}

	rl.EndMode3D()
	drawSidebar(s)

	rl.DrawFPS(10, 10)
	rl.EndDrawing()

}

func drawFluid(s *physics.Simulation) {
	for _, unit := range s.Fluid {

		color := unit.Color

		if s.Config.ShowSpeedColor {
			if s.Config.UseExperimentalOctree {
				//color = utils.GetColorFromVelocity(unit.Velocity)
			} else {
				color = utils.GetColorFromVelocity(unit.GetVelocity(s.Metrics.Frametime))
			}
		} else if s.Config.ShowClusterColor {
			color = unit.BlendedColor()
		} else if s.Config.ShowMassColor {
			color = utils.GetColorFromMass(unit.Mass)
		}

		if s.Config.ShowVectors {
			drawVectors(unit, s.Metrics.Frametime)
		}

		rl.DrawSphere(rl.NewVector3(unit.Position.X, unit.Position.Y, unit.Position.Z), unit.Radius, color)
	}
}

func drawOctree(octree *physics.Octree) {
	if octree == nil {
		return // Ritorna se l'Octree è nil
	}

	// Disegna il BoundingBox dell'Octree corrente
	rl.DrawBoundingBox(octree.Bounds, rl.Red)

	// Disegna ricorsivamente i BoundingBox dei sotto-Octrees
	for _, child := range octree.Children {
		drawOctree(child)
	}
}

func drawVectors(u *physics.Unit, dt float32) {

	endVelocity := rl.Vector3Add(u.Position, rl.Vector3Scale(u.GetVelocity(dt), 0.1))

	rl.DrawLine3D(rl.NewVector3(u.Position.X, u.Position.Y, u.Position.Z), rl.NewVector3(endVelocity.X, endVelocity.Y, endVelocity.Z), rl.Blue)

	endAcceleration := rl.Vector3Add(u.Position, rl.Vector3Scale(u.Acceleration, 0.1))

	rl.DrawLine3D(rl.NewVector3(u.Position.X, u.Position.Y, u.Position.Z), rl.NewVector3(endAcceleration.X, endAcceleration.Y, endAcceleration.Z), rl.Orange)
}

func calculateNearestCubePoints(s *physics.Simulation) (xNear, yNear, zNear rl.Vector3) {
	if s.SpawnPosition.X < s.CubeCenter.X {
		xNear = rl.NewVector3(float32(s.WorldBoundray.Min.X), s.SpawnPosition.Y, s.SpawnPosition.Z)
	} else {
		xNear = rl.NewVector3(float32(s.WorldBoundray.Max.X), s.SpawnPosition.Y, s.SpawnPosition.Z)
	}
	if s.SpawnPosition.Y < s.CubeCenter.Y {
		yNear = rl.NewVector3(s.SpawnPosition.X, float32(s.WorldBoundray.Min.Y), s.SpawnPosition.Z)
	} else {
		yNear = rl.NewVector3(s.SpawnPosition.X, float32(s.WorldBoundray.Max.Y), s.SpawnPosition.Z)
	}
	if s.SpawnPosition.Z < s.CubeCenter.Z {
		zNear = rl.NewVector3(s.SpawnPosition.X, s.SpawnPosition.Y, float32(s.WorldBoundray.Min.Z))
	} else {
		zNear = rl.NewVector3(s.SpawnPosition.X, s.SpawnPosition.Y, float32(s.WorldBoundray.Max.Z))
	}
	return
}

func drawDashedLine(start, end rl.Vector3, dashesLength, spaceLength float32) {
	direction := rl.Vector3Subtract(end, start)
	totalLength := rl.Vector3Length(direction)
	direction = rl.Vector3Normalize(direction)

	for currentLength := float32(0); currentLength < totalLength; {
		nextDashEnd := currentLength + dashesLength
		if nextDashEnd > totalLength {
			nextDashEnd = totalLength
		}

		dashStart := rl.Vector3Add(start, rl.Vector3Scale(direction, currentLength))
		dashEnd := rl.Vector3Add(start, rl.Vector3Scale(direction, nextDashEnd))
		rl.DrawLine3D(dashStart, dashEnd, rl.Black) // Cambia il colore se necessario

		currentLength = nextDashEnd + spaceLength
	}
}
