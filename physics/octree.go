package physics

import (
	"github.com/EliCDavis/vector/vector3"
	"github.com/alexanderi96/go-fluid-simulator/config"
	"github.com/alexanderi96/go-fluid-simulator/utils"

	"github.com/g3n/engine/core"
	"github.com/g3n/engine/graphic"
)

var (
	maxObjects, maxLevels int8
)

func InitOctree(config *config.Config) {
	maxLevels = config.OctreeMaxLevel
	maxObjects = config.MaxUnitNumberPerLevel
}

type Octree struct {
	level    int8
	Bounds   BoundingBox
	objects  []*Unit
	Children [8]*Octree
	divided  bool
	wf       *graphic.Lines

	CenterOfMass vector3.Vector[float64]
	TotalMass    float64
}

// Octree crea un nuovo Octree.
func NewOctree(level int8, bounds BoundingBox, scene *core.Node) *Octree {
	wireframe := &graphic.Lines{}
	if true {
		wireframe = utils.GetBoundsLine(bounds.Min, bounds.Max)
		wireframe.SetVisible(true)
		scene.Add(wireframe)
	}

	return &Octree{
		level:   level,
		Bounds:  bounds,
		divided: false,
		wf:      wireframe,

		CenterOfMass: vector3.Zero[float64](),
		TotalMass:    0,
	}
}

// Clear pulisce il Octree.
func (ot *Octree) Clear(scene *core.Node) {
	ot.objects = ot.objects[:0]
	ot.TotalMass = 0
	ot.CenterOfMass = vector3.Zero[float64]()
	scene.Remove(ot.wf)

	ot.wf = nil

	for i := 0; i < len(ot.Children); i++ {
		if ot.Children[i] != nil {
			ot.Children[i].Clear(scene)
			ot.Children[i] = nil
		}
	}
}

// Split divide il Octree in quattro sotto-Octrees.
func (ot *Octree) Split(scene *core.Node) {
	subWidth := (ot.Bounds.Max.X() - ot.Bounds.Min.X()) / 2
	subHeight := (ot.Bounds.Max.Y() - ot.Bounds.Min.Y()) / 2
	subDepth := (ot.Bounds.Max.Z() - ot.Bounds.Min.Z()) / 2

	minX := ot.Bounds.Min.X()
	minY := ot.Bounds.Min.Y()
	minZ := ot.Bounds.Min.Z()

	level := ot.level + 1

	// Creazione dei nuovi otto sotto-octrees.
	// children[0] = inBottom && inLeft && inBack (0)
	BLB := BoundingBox{Min: vector3.New(minX, minY, minZ), Max: vector3.New(minX+subWidth, minY+subHeight, minZ+subDepth)}
	ot.Children[0] = NewOctree(level, BLB, scene)
	// children[1] = inBottom && inRight && inBack (1)
	BRB := BoundingBox{Min: vector3.New(minX+subWidth, minY, minZ), Max: vector3.New(minX+2*subWidth, minY+subHeight, minZ+subDepth)}
	ot.Children[1] = NewOctree(level, BRB, scene)
	// children[5] = inBottom && inRight && inFront (2)
	BRF := BoundingBox{Min: vector3.New(minX+subWidth, minY, minZ+subDepth), Max: vector3.New(minX+2*subWidth, minY+subHeight, minZ+2*subDepth)}
	ot.Children[2] = NewOctree(level, BRF, scene)
	// children[4] = inBottom && inLeft && inFront (3)
	BLF := BoundingBox{Min: vector3.New(minX, minY, minZ+subDepth), Max: vector3.New(minX+subWidth, minY+subHeight, minZ+2*subDepth)}
	ot.Children[3] = NewOctree(level, BLF, scene)
	// children[2] = inTop && inLeft && inBack (4)
	TLB := BoundingBox{Min: vector3.New(minX, minY+subHeight, minZ), Max: vector3.New(minX+subWidth, minY+2*subHeight, minZ+subDepth)}
	ot.Children[4] = NewOctree(level, TLB, scene)
	// children[3] = inTop && inRight && inBack (5)
	TRB := BoundingBox{Min: vector3.New(minX+subWidth, minY+subHeight, minZ), Max: vector3.New(minX+2*subWidth, minY+2*subHeight, minZ+subDepth)}
	ot.Children[5] = NewOctree(level, TRB, scene)
	// children[7] = inTop && inRight && inFront (6)
	TRF := BoundingBox{Min: vector3.New(minX+subWidth, minY+subHeight, minZ+subDepth), Max: vector3.New(minX+2*subWidth, minY+2*subHeight, minZ+2*subDepth)}
	ot.Children[6] = NewOctree(level, TRF, scene)
	// children[6] = inTop && inLeft && inFront (7)
	TLF := BoundingBox{Min: vector3.New(minX, minY+subHeight, minZ+subDepth), Max: vector3.New(minX+subWidth, minY+2*subHeight, minZ+2*subDepth)}
	ot.Children[7] = NewOctree(level, TLF, scene)

	ot.divided = true
}

// Insert inserisce un oggetto nel Octree.
func (ot *Octree) Insert(obj *Unit, scene *core.Node) {
	if ot.Children[0] == nil {
		if len(ot.objects) < int(maxObjects) || ot.level >= maxLevels {
			ot.objects = append(ot.objects, obj)
			ot.updateMassAndCenterOfMass(obj)
			return
		}

		ot.Split(scene)
		for _, item := range ot.objects {
			ot.insertUnitIntoChildren(item, scene)
		}
		ot.objects = ot.objects[:0]
	} else {
		ot.insertUnitIntoChildren(obj, scene)
	}
}

func (ot *Octree) insertUnitIntoChildren(obj *Unit, scene *core.Node) {
	indices := ot.getIndices(*obj)
	inserted := false
	for _, index := range indices {
		if index != -1 {
			ot.Children[index].Insert(obj, scene)
			inserted = true
		}
	}

	if !inserted {
		ot.objects = append(ot.objects, obj)
	}
	// ot.updateMassAndCenterOfMass(obj)
}

func (ot *Octree) updateMassAndCenterOfMass(obj *Unit) {
	oldTotalMass := ot.TotalMass
	ot.TotalMass += obj.Mass
	if oldTotalMass == 0 {
		ot.CenterOfMass = obj.Position
	} else {
		ot.CenterOfMass = ot.CenterOfMass.Scale(oldTotalMass).Add(obj.Position.Scale(obj.Mass)).Scale(1 / ot.TotalMass)
	}
}

// getIndex determina in quale sotto-Octree un oggetto appartiene.
func (ot *Octree) getIndices(obj Unit) []int {
	indices := make([]int, 0, 8) // Preallocazione con la dimensione massima possibile

	// Confronta direttamente le coordinate dell'oggetto con i limiti dell'Octree
	for i := 0; i < 8; i++ {
		child := ot.Children[i]
		if child != nil && obj.Position.X() >= child.Bounds.Min.X() && obj.Position.X() <= child.Bounds.Max.X() &&
			obj.Position.Y() >= child.Bounds.Min.Y() && obj.Position.Y() <= child.Bounds.Max.Y() &&
			obj.Position.Z() >= child.Bounds.Min.Z() && obj.Position.Z() <= child.Bounds.Max.Z() {
			indices = append(indices, i)
		}
	}

	if len(indices) == 0 {
		return []int{-1}
	}

	return indices
}

// Retrieve restituisce tutti gli oggetti che potrebbero collidere con l'oggetto dato.
func (ot *Octree) Retrieve(returnObjects *[]*Unit, obj *Unit) {
	indices := ot.getIndices(*obj)
	for _, index := range indices {
		if index != -1 && ot.Children[index] != nil {
			ot.Children[index].Retrieve(returnObjects, obj)
		}
	}

	// Aggiungere gli oggetti di questo nodo Octree alla lista di ritorno.
	*returnObjects = append(*returnObjects, ot.objects...)
}
