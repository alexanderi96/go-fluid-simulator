package draw

// import (
// 	"github.com/EliCDavis/vector/vector3"
// 	"github.com/alexanderi96/go-fluid-simulator/physics"
// 	"github.com/alexanderi96/go-fluid-simulator/utils"

// 	rl "github.com/gen2brain/raylib-go/raylib"
// )

// // TODO: https://github.com/g3n/engine

// func Draw(s *physics.Simulation) {
// 	rl.BeginDrawing()
// 	rl.ClearBackground(rl.NewColor(10, 10, 10, 255))

// 	cubeColor := rl.Red
// 	// Calcola i punti più vicini sul cubo di gioco
// 	xNear, yNear, zNear := calculateNearestCubePoints(s)

// 	rl.BeginMode3D(s.Camera)

// 	// Disegno di un pavimento a griglia come punto di riferimento
// 	// rl.DrawGrid(20, 5)

// 	drawFluid(s)

// 	if s.ControlMode == physics.UnitSpawnMode {

// 		if s.IsSpawnInRange() {
// 			cubeColor = rl.Green

// 			// Disegna le linee tratteggiate
// 			drawDashedLine(xNear, s.FinalSpawnPosition, 0.1, 0.1)
// 			drawDashedLine(yNear, s.FinalSpawnPosition, 0.1, 0.1)
// 			drawDashedLine(zNear, s.FinalSpawnPosition, 0.1, 0.1)

// 		}
// 		rl.DrawCube(utils.ToRlVector3(s.FinalSpawnPosition), 1, 1, 1, cubeColor) // Modifica le dimensioni e il colore come preferisci

// 		if s.MouseButtonPressed && s.InitialMousePosition.X > 0 && s.InitialMousePosition.X < float32(s.Config.WindowWidth-s.Config.SidebarWidth) {
// 			rl.DrawLine3D(utils.ToRlVector3(s.InitialSpawnPosition), utils.ToRlVector3(s.FinalSpawnPosition), rl.RayWhite)
// 		}
// 	}

// 	if s.Config.ShowOctree {
// 		drawOctree(s.Octree)
// 	} else {
// 		rl.DrawBoundingBox(utils.ToRlBoundingBox(s.WorldBoundray.Min, s.WorldBoundray.Max), rl.RayWhite)
// 	}

// 	rl.DrawLine3D(utils.ToRlVector3(s.WorldBoundray.Min), utils.ToRlVector3(vector3.New(s.WorldBoundray.Max.X(), s.WorldBoundray.Min.Y(), s.WorldBoundray.Min.Z())), rl.Red)
// 	rl.DrawLine3D(utils.ToRlVector3(s.WorldBoundray.Min), utils.ToRlVector3(vector3.New(s.WorldBoundray.Min.X(), s.WorldBoundray.Max.Y(), s.WorldBoundray.Min.Z())), rl.Green)
// 	rl.DrawLine3D(utils.ToRlVector3(s.WorldBoundray.Min), utils.ToRlVector3(vector3.New(s.WorldBoundray.Min.X(), s.WorldBoundray.Min.Y(), s.WorldBoundray.Max.Z())), rl.Blue)

// 	rl.EndMode3D()
// 	drawSidebar(s)

// 	rl.DrawFPS(10, 10)
// 	rl.EndDrawing()

// }

// func drawFluid(s *physics.Simulation) {
// 	for _, unit := range s.Fluid {

// 		color := utils.KelvinToRGBA(unit.Heat)

// 		if s.Config.ShowSpeedColor {
// 			color = utils.GetColorFromVelocity(unit.Velocity)
// 		}

// 		if s.Config.ShowVectors {
// 			drawVectors(unit)
// 		}

// 		// Disegna la sfera che rappresenta l'unità
// 		rl.DrawSphere(utils.ToRlVector3(unit.Position), float32(unit.Radius), color)
// 	}
// }

// func drawOctree(octree *physics.Octree) {
// 	if octree == nil {
// 		return // Ritorna se l'Octree è nil
// 	}

// 	// Disegna il BoundingBox dell'Octree corrente
// 	rl.DrawBoundingBox(utils.ToRlBoundingBox(octree.Bounds.Min, octree.Bounds.Max), rl.RayWhite)
// 	//rl.DrawSphere(utils.ToRlVector3(octree.CenterOfMass), float32(octree.TotalMass/(4*math.Pi)), rl.Red)

// 	// Disegna ricorsivamente i BoundingBox dei sotto-Octrees
// 	for _, child := range octree.Children {
// 		drawOctree(child)
// 	}
// }

// func drawVectors(u *physics.Unit) {

// 	endVelocity := u.Position.Add(u.Velocity.Scale(1000))
// 	rl.DrawLine3D(utils.ToRlVector3(u.Position), utils.ToRlVector3(endVelocity), rl.Blue)

// 	endAcceleration := u.Position.Add(u.Acceleration.Scale(1))

// 	rl.DrawLine3D(utils.ToRlVector3(u.Position), utils.ToRlVector3(endAcceleration), rl.Orange)
// }

// func calculateNearestCubePoints(s *physics.Simulation) (xNear, yNear, zNear vector3.Vector[float64]) {
// 	if s.FinalSpawnPosition.X() < s.WorldCenter.X() {
// 		xNear = vector3.New(s.WorldBoundray.Min.X(), s.FinalSpawnPosition.Y(), s.FinalSpawnPosition.Z())
// 	} else {
// 		xNear = vector3.New(s.WorldBoundray.Max.X(), s.FinalSpawnPosition.Y(), s.FinalSpawnPosition.Z())
// 	}
// 	if s.FinalSpawnPosition.Y() < s.WorldCenter.Y() {
// 		yNear = vector3.New(s.FinalSpawnPosition.X(), s.WorldBoundray.Min.Y(), s.FinalSpawnPosition.Z())
// 	} else {
// 		yNear = vector3.New(s.FinalSpawnPosition.X(), s.WorldBoundray.Max.Y(), s.FinalSpawnPosition.Z())
// 	}
// 	if s.FinalSpawnPosition.Z() < s.WorldCenter.Z() {
// 		zNear = vector3.New(s.FinalSpawnPosition.X(), s.FinalSpawnPosition.Y(), s.WorldBoundray.Min.Z())
// 	} else {
// 		zNear = vector3.New(s.FinalSpawnPosition.X(), s.FinalSpawnPosition.Y(), s.WorldBoundray.Max.Z())
// 	}
// 	return
// }

// func drawDashedLine(start, end vector3.Vector[float64], dashesLength, spaceLength float64) {
// 	direction := end.Sub(start)
// 	totalLength := direction.Length()
// 	direction = direction.Normalized()

// 	for currentLength := float64(0); currentLength < totalLength; {
// 		nextDashEnd := currentLength + dashesLength
// 		if nextDashEnd > totalLength {
// 			nextDashEnd = totalLength
// 		}

// 		dashStart := start.Add(direction.Scale(currentLength))
// 		dashEnd := start.Add(direction.Scale(nextDashEnd))
// 		rl.DrawLine3D(utils.ToRlVector3(dashStart), utils.ToRlVector3(dashEnd), rl.RayWhite) // Cambia il colore se necessario

// 		currentLength = nextDashEnd + spaceLength
// 	}
// }
