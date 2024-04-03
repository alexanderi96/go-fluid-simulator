package physics

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/google/uuid"
)

func newCluster(unitA, unitB *Unit) *Cluster {
	return &Cluster{
		Id:    uuid.New(),
		Color: blendedColorBasedOnMasses(unitA, unitB),
	}
}

func blendedColorBasedOnMasses(unitA, unitB *Unit) color.RGBA {
	// Assicurati che t sia compreso tra 0 e 1
	t := unitA.Mass / unitB.Mass

	unitAColor := unitA.BlendedColor()
	unitBColor := unitB.BlendedColor()
	if t <= 0 {
		return unitAColor
	} else if t > 1 {
		return unitBColor
	}

	// Calcola i componenti del nuovo colore interpolando linearmente tra color1 e clusterColor
	r := float32(unitAColor.R) + (float32(unitBColor.R)-float32(unitAColor.R))*t
	g := float32(unitAColor.G) + (float32(unitBColor.G)-float32(unitAColor.G))*t
	b := float32(unitAColor.B) + (float32(unitBColor.B)-float32(unitAColor.B))*t
	a := float32(unitAColor.A) + (float32(unitBColor.A)-float32(unitAColor.A))*t

	return rl.NewColor(uint8(r), uint8(g), uint8(b), uint8(a))
}

func updateClusters(unitA, unitB *Unit, clusterMasses map[uuid.UUID]float32) {
	if unitA.Cluster != nil && unitB.Cluster != nil && unitA.Cluster.Id != unitB.Cluster.Id {
		if clusterMasses[unitA.Cluster.Id] < clusterMasses[unitB.Cluster.Id] {
			unitA.TransitionTimer = 0
			unitA.OldCluster = unitA.Cluster
			unitA.Cluster = unitB.Cluster
		} else {
			unitB.TransitionTimer = 0
			unitB.OldCluster = unitB.Cluster
			unitB.Cluster = unitA.Cluster
		}
	} else if unitA.Cluster == nil && unitB.Cluster != nil {
		unitA.TransitionTimer = 0
		unitA.Cluster = unitB.Cluster
	} else if unitB.Cluster == nil && unitA.Cluster != nil {
		unitB.TransitionTimer = 0
		unitB.Cluster = unitA.Cluster
	} else if unitA.Cluster == nil && unitB.Cluster == nil {
		c := *newCluster(unitA, unitB)
		unitA.Cluster = &c
		unitB.Cluster = &c
	}
}

func (s *Simulation) GetClusterMass(clusterId uuid.UUID) float32 {
	mass := float32(0)
	for _, unit := range s.Fluid {
		if unit.Cluster == nil {
			continue
		}
		if unit.Cluster.Id != clusterId {
			continue
		}
		mass += unit.Mass
	}
	return mass
}
