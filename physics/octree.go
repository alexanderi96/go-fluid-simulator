package physics

import (
	"github.com/EliCDavis/vector/vector3"
	"github.com/alexanderi96/go-fluid-simulator/config"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
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

// NewWireframeOctree crea un oggetto grafico che rappresenta il wireframe dell'octree.
func getOctreeWireframe(min, max vector3.Vector[float64]) *graphic.Lines {

	// Crea il wireframe per il nodo corrente
	vertices := math32.NewArrayF32(0, 16)
	vertices.Append(
		float32(min.X()), float32(min.Y()), float32(min.Z()), float32(max.X()), float32(min.Y()), float32(min.Z()),
		float32(max.X()), float32(min.Y()), float32(min.Z()), float32(max.X()), float32(max.Y()), float32(min.Z()),
		float32(max.X()), float32(max.Y()), float32(min.Z()), float32(min.X()), float32(max.Y()), float32(min.Z()),
		float32(min.X()), float32(max.Y()), float32(min.Z()), float32(min.X()), float32(min.Y()), float32(min.Z()),
		float32(min.X()), float32(min.Y()), float32(max.Z()), float32(max.X()), float32(min.Y()), float32(max.Z()),
		float32(max.X()), float32(min.Y()), float32(max.Z()), float32(max.X()), float32(max.Y()), float32(max.Z()),
		float32(max.X()), float32(max.Y()), float32(max.Z()), float32(min.X()), float32(max.Y()), float32(max.Z()),
		float32(min.X()), float32(max.Y()), float32(max.Z()), float32(min.X()), float32(min.Y()), float32(max.Z()),
		float32(min.X()), float32(min.Y()), float32(min.Z()), float32(min.X()), float32(min.Y()), float32(max.Z()),
		float32(max.X()), float32(min.Y()), float32(min.Z()), float32(max.X()), float32(min.Y()), float32(max.Z()),
		float32(max.X()), float32(max.Y()), float32(min.Z()), float32(max.X()), float32(max.Y()), float32(max.Z()),
		float32(min.X()), float32(max.Y()), float32(min.Z()), float32(min.X()), float32(max.Y()), float32(max.Z()),
	)
	vbo := gls.NewVBO(vertices).AddAttrib(gls.VertexPosition)

	geom := geometry.NewGeometry()
	geom.AddVBO(vbo)
	mat := material.NewBasic()
	mat.SetLineWidth(1)
	mat.SetSide(material.SideDouble)
	return graphic.NewLines(geom, mat)
}

// Octree crea un nuovo Octree.
func NewOctree(level int8, bounds BoundingBox, scene *core.Node) *Octree {

	wireframe := getOctreeWireframe(bounds.Min, bounds.Max)
	wireframe.SetVisible(false)
	scene.Add(wireframe)
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
	// children[0] = inBottom && inLeft && inBack
	swb := BoundingBox{Min: vector3.New(minX, minY, minZ), Max: vector3.New(minX+subWidth, minY+subHeight, minZ+subDepth)}
	ot.Children[0] = NewOctree(level, swb, scene)
	// children[1] = inBottom && inRight && inBack
	seb := BoundingBox{Min: vector3.New(minX+subWidth, minY, minZ), Max: vector3.New(minX+2*subWidth, minY+subHeight, minZ+subDepth)}
	ot.Children[1] = NewOctree(level, seb, scene)
	// children[2] = inTop && inLeft && inBack
	nwb := BoundingBox{Min: vector3.New(minX, minY+subHeight, minZ), Max: vector3.New(minX+subWidth, minY+2*subHeight, minZ+subDepth)}
	ot.Children[2] = NewOctree(level, nwb, scene)
	// children[3] = inTop && inRight && inBack
	neb := BoundingBox{Min: vector3.New(minX+subWidth, minY+subHeight, minZ), Max: vector3.New(minX+2*subWidth, minY+2*subHeight, minZ+subDepth)}
	ot.Children[3] = NewOctree(level, neb, scene)
	// children[4] = inBottom && inLeft && inFront
	swf := BoundingBox{Min: vector3.New(minX, minY, minZ+subDepth), Max: vector3.New(minX+subWidth, minY+subHeight, minZ+2*subDepth)}
	ot.Children[4] = NewOctree(level, swf, scene)
	// children[5] = inBottom && inRight && inFront
	sef := BoundingBox{Min: vector3.New(minX+subWidth, minY, minZ+subDepth), Max: vector3.New(minX+2*subWidth, minY+subHeight, minZ+2*subDepth)}
	ot.Children[5] = NewOctree(level, sef, scene)
	// children[6] = inTop && inLeft && inFront
	nwf := BoundingBox{Min: vector3.New(minX, minY+subHeight, minZ+subDepth), Max: vector3.New(minX+subWidth, minY+2*subHeight, minZ+2*subDepth)}
	ot.Children[6] = NewOctree(level, nwf, scene)
	// children[7] = inTop && inRight && inFront
	nef := BoundingBox{Min: vector3.New(minX+subWidth, minY+subHeight, minZ+subDepth), Max: vector3.New(minX+2*subWidth, minY+2*subHeight, minZ+2*subDepth)}
	ot.Children[7] = NewOctree(level, nef, scene)

	ot.divided = true
}

// Insert inserisce un oggetto nel Octree.
func (ot *Octree) Insert(obj *Unit, scene *core.Node) {

	// Calcola il nuovo centro di massa come media ponderata
	oldTotalMass := ot.TotalMass
	ot.TotalMass += obj.Mass
	if oldTotalMass == 0 {
		ot.CenterOfMass = obj.Position
	} else {
		ot.CenterOfMass = ot.CenterOfMass.Scale(oldTotalMass).Add(obj.Position.Scale(obj.Mass)).Scale(1 / ot.TotalMass)
	}

	if ot.Children[0] == nil {
		if len(ot.objects) < int(maxObjects) || ot.level >= maxLevels {
			// Se il nodo corrente ha spazio o abbiamo raggiunto il livello massimo, aggiungi qui.
			ot.objects = append(ot.objects, obj)
			return
		}

		// Se il nodo corrente è pieno e non al livello massimo, dividi.
		ot.Split(scene)

		// Reinserisci gli oggetti negli octree figli.
		for _, item := range ot.objects {
			ot.insertUnitIntoChildren(item, scene)
		}
		ot.objects = ot.objects[:0] // Svuota l'elenco degli oggetti nel nodo corrente dopo la suddivisione
	} else {
		// Prova ad inserire l'oggetto nei figli.
		ot.insertUnitIntoChildren(obj, scene)
	}
}

// insertUnitIntoChildren inserisce un'unità nei figli dell'octree.
func (ot *Octree) insertUnitIntoChildren(obj *Unit, scene *core.Node) {
	indices := ot.getIndices(*obj)
	inserted := false
	for _, index := range indices {
		if index != -1 {
			ot.Children[index].Insert(obj, scene)
			inserted = true
		}
	}

	// Se l'oggetto non è stato inserito in nessun figlio, aggiungilo a questo nodo.
	if !inserted {
		ot.objects = append(ot.objects, obj)
	} else {
		// Aggiorna la massa totale e il centro di massa dell'Octree padre.
		ot.TotalMass += obj.Mass
		ot.CenterOfMass = ot.CenterOfMass.Scale(ot.TotalMass - obj.Mass).Add(obj.Position.Scale(obj.Mass)).Scale(1 / ot.TotalMass)
	}
}

// getIndex determina in quale sotto-Octree un oggetto appartiene.
func (ot *Octree) getIndices(obj Unit) []int {
	// Calcola il punto medio dell'Octree per le tre dimensioni.
	midX := (ot.Bounds.Min.X() + ot.Bounds.Max.X()) / 2
	midY := (ot.Bounds.Min.Y() + ot.Bounds.Max.Y()) / 2
	midZ := (ot.Bounds.Min.Z() + ot.Bounds.Max.Z()) / 2

	// Calcola gli estremi dell'oggetto considerando il suo raggio.
	minX := obj.Position.X() + obj.Radius
	maxX := obj.Position.X() - obj.Radius
	minY := obj.Position.Y() + obj.Radius
	maxY := obj.Position.Y() - obj.Radius
	minZ := obj.Position.Z() + obj.Radius
	maxZ := obj.Position.Z() - obj.Radius

	// Determina la posizione dell'oggetto rispetto al punto medio per ogni dimensione,
	// tenendo conto dei raggi delle unità.
	inLeft := minX <= midX
	inRight := maxX >= midX
	inBottom := minY <= midY
	inTop := maxY >= midY
	inBack := minZ <= midZ
	inFront := maxZ >= midZ

	// Inizializza un array vuoto per gli indici.
	var indices []int

	// Assegna un indice basato sulla posizione dell'oggetto.
	// Utilizza una struttura condizionale compatta per verificare tutte le possibili combinazioni.
	if inTop {
		if inRight {
			if inFront {
				indices = append(indices, 7)
			}
			if inBack {
				indices = append(indices, 3)
			}
		}
		if inLeft {
			if inFront {
				indices = append(indices, 6)
			}
			if inBack {
				indices = append(indices, 2)
			}
		}
	}
	if inBottom {
		if inRight {
			if inFront {
				indices = append(indices, 5)
			}
			if inBack {
				indices = append(indices, 1)
			}
		}
		if inLeft {
			if inFront {
				indices = append(indices, 4)
			}
			if inBack {
				indices = append(indices, 0)
			}
		}
	}

	// Se l'array degli indici è vuoto, significa che l'oggetto non appartiene a nessun ottante.
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
