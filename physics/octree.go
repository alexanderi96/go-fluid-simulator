package physics

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	maxObjects = 4
	maxLevels  = 2
)

type Octree struct {
	level    int
	Bounds   rl.BoundingBox
	objects  []*Unit
	Children [8]*Octree
}

// Octree crea un nuovo Octree.
func NewOctree(level int, bounds rl.BoundingBox) *Octree {
	return &Octree{
		level:  level,
		Bounds: bounds,
	}
}

// Clear pulisce il Octree.
func (qt *Octree) Clear() {
	qt.objects = qt.objects[:0]
	for i := 0; i < 8; i++ {
		if qt.Children[i] != nil {
			qt.Children[i].Clear()
			qt.Children[i] = nil
		}
	}
}

// Split divide il Octree in quattro sotto-Octrees.
func (ot *Octree) Split() {
	subWidth := (ot.Bounds.Max.X - ot.Bounds.Min.X) / 2
	subHeight := (ot.Bounds.Max.Y - ot.Bounds.Min.Y) / 2
	subDepth := (ot.Bounds.Max.Z - ot.Bounds.Min.Z) / 2

	minX := ot.Bounds.Min.X
	minY := ot.Bounds.Min.Y
	minZ := ot.Bounds.Min.Z

	// Creazione dei nuovi otto sotto-octrees.
	ot.Children[0] = NewOctree(ot.level+1, rl.BoundingBox{Min: rl.Vector3{X: minX, Y: minY, Z: minZ}, Max: rl.Vector3{X: minX + subWidth, Y: minY + subHeight, Z: minZ + subDepth}})
	ot.Children[1] = NewOctree(ot.level+1, rl.BoundingBox{Min: rl.Vector3{X: minX + subWidth, Y: minY, Z: minZ}, Max: rl.Vector3{X: minX + 2*subWidth, Y: minY + subHeight, Z: minZ + subDepth}})
	ot.Children[2] = NewOctree(ot.level+1, rl.BoundingBox{Min: rl.Vector3{X: minX, Y: minY + subHeight, Z: minZ}, Max: rl.Vector3{X: minX + subWidth, Y: minY + 2*subHeight, Z: minZ + subDepth}})
	ot.Children[3] = NewOctree(ot.level+1, rl.BoundingBox{Min: rl.Vector3{X: minX + subWidth, Y: minY + subHeight, Z: minZ}, Max: rl.Vector3{X: minX + 2*subWidth, Y: minY + 2*subHeight, Z: minZ + subDepth}})
	ot.Children[4] = NewOctree(ot.level+1, rl.BoundingBox{Min: rl.Vector3{X: minX, Y: minY, Z: minZ + subDepth}, Max: rl.Vector3{X: minX + subWidth, Y: minY + subHeight, Z: minZ + 2*subDepth}})
	ot.Children[5] = NewOctree(ot.level+1, rl.BoundingBox{Min: rl.Vector3{X: minX + subWidth, Y: minY, Z: minZ + subDepth}, Max: rl.Vector3{X: minX + 2*subWidth, Y: minY + subHeight, Z: minZ + 2*subDepth}})
	ot.Children[6] = NewOctree(ot.level+1, rl.BoundingBox{Min: rl.Vector3{X: minX, Y: minY + subHeight, Z: minZ + subDepth}, Max: rl.Vector3{X: minX + subWidth, Y: minY + 2*subHeight, Z: minZ + 2*subDepth}})
	ot.Children[7] = NewOctree(ot.level+1, rl.BoundingBox{Min: rl.Vector3{X: minX + subWidth, Y: minY + subHeight, Z: minZ + subDepth}, Max: rl.Vector3{X: minX + 2*subWidth, Y: minY + 2*subHeight, Z: minZ + 2*subDepth}})
}

// Insert inserisce un oggetto nel Octree.
func (qt *Octree) Insert(obj *Unit) {
	// Se i figli sono nil, dividere il Octree
	if qt.Children[0] == nil && qt.Children[1] == nil && qt.Children[2] == nil && qt.Children[3] == nil && qt.Children[4] == nil && qt.Children[5] == nil && qt.Children[6] == nil && qt.Children[7] == nil {
		qt.Split()
	}

	index := qt.getIndex(*obj)
	if index != -1 {
		qt.Children[index].Insert(obj)
		return
	}

	if len(qt.objects) > maxObjects && qt.level < maxLevels {
		if len(qt.Children) == 0 {
			qt.Split()
		}

		i := 0
		for i < len(qt.objects) {
			index := qt.getIndex(*qt.objects[i])
			if index != -1 {
				qt.Children[index].Insert(qt.objects[i])
				qt.objects = append(qt.objects[:i], qt.objects[i+1:]...)
			} else {
				i++
			}
		}
	}
	qt.objects = append(qt.objects, obj)
}

// getIndex determina in quale sotto-Octree un oggetto appartiene.
func (ot *Octree) getIndex(obj Unit) int {
	midX := (ot.Bounds.Min.X + ot.Bounds.Max.X) / 2
	midY := (ot.Bounds.Min.Y + ot.Bounds.Max.Y) / 2
	midZ := (ot.Bounds.Min.Z + ot.Bounds.Max.Z) / 2

	// Determina in quale ottante dell'Octree l'oggetto si trova.
	inLeft := obj.Position.X < midX
	inRight := !inLeft
	inBottom := obj.Position.Y < midY
	inTop := !inBottom
	inBack := obj.Position.Z < midZ
	inFront := !inBack

	switch {
	case inTop && inRight && inFront:
		return 0
	case inTop && inLeft && inFront:
		return 1
	case inBottom && inLeft && inFront:
		return 2
	case inBottom && inRight && inFront:
		return 3
	case inTop && inRight && inBack:
		return 4
	case inTop && inLeft && inBack:
		return 5
	case inBottom && inLeft && inBack:
		return 6
	case inBottom && inRight && inBack:
		return 7
	}

	return -1 // Se l'oggetto non rientra in nessun ottante.
}

// Retrieve restituisce tutti gli oggetti che potrebbero collidere con l'oggetto dato.
func (ot *Octree) Retrieve(returnObjects *[]*Unit, obj *Unit) {
	index := ot.getIndex(*obj)

	// Se l'oggetto si trova in un ottante specifico e quel figlio esiste, cercare ricorsivamente in quel figlio.
	if index != -1 && ot.Children[index] != nil {
		ot.Children[index].Retrieve(returnObjects, obj)
	} else {
		// Se l'oggetto non si adatta a un singolo ottante (potrebbe sovrapporsi a più ottanti) o l'indice è -1,
		// dobbiamo controllare tutti i figli perché l'oggetto potrebbe intersecarsi con più ottanti.
		for i := 0; i < len(ot.Children); i++ {
			if ot.Children[i] != nil {
				ot.Children[i].Retrieve(returnObjects, obj)
			}
		}
	}

	// Aggiungere gli oggetti di questo nodo Octree alla lista di ritorno.
	*returnObjects = append(*returnObjects, ot.objects...)
}
