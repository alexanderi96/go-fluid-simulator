package physics

import (
	"github.com/alexanderi96/go-fluid-simulator/config"
	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	maxObjects, maxLevels int
)

func InitOctree(config *config.Config) {
	maxLevels = int(config.OctreeMaxLevel)
	maxObjects = int(config.MaxUnitNumberPerLevel)
}

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
func (ot *Octree) Clear() {
	ot.objects = ot.objects[:0]
	for i := 0; i < 8; i++ {
		if ot.Children[i] != nil {
			ot.Children[i].Clear()
			ot.Children[i] = nil
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

	level := ot.level + 1

	// Creazione dei nuovi otto sotto-octrees.
	// children[0] = inBottom && inLeft && inBack
	ot.Children[0] = NewOctree(level, rl.BoundingBox{Min: rl.Vector3{X: minX, Y: minY, Z: minZ}, Max: rl.Vector3{X: minX + subWidth, Y: minY + subHeight, Z: minZ + subDepth}})
	// children[1] = inBottom && inRight && inBack
	ot.Children[1] = NewOctree(level, rl.BoundingBox{Min: rl.Vector3{X: minX + subWidth, Y: minY, Z: minZ}, Max: rl.Vector3{X: minX + 2*subWidth, Y: minY + subHeight, Z: minZ + subDepth}})
	// children[2] = inTop && inLeft && inBack
	ot.Children[2] = NewOctree(level, rl.BoundingBox{Min: rl.Vector3{X: minX, Y: minY + subHeight, Z: minZ}, Max: rl.Vector3{X: minX + subWidth, Y: minY + 2*subHeight, Z: minZ + subDepth}})
	// children[3] = inTop && inRight && inBack
	ot.Children[3] = NewOctree(level, rl.BoundingBox{Min: rl.Vector3{X: minX + subWidth, Y: minY + subHeight, Z: minZ}, Max: rl.Vector3{X: minX + 2*subWidth, Y: minY + 2*subHeight, Z: minZ + subDepth}})
	// children[4] = inBottom && inLeft && inFront
	ot.Children[4] = NewOctree(level, rl.BoundingBox{Min: rl.Vector3{X: minX, Y: minY, Z: minZ + subDepth}, Max: rl.Vector3{X: minX + subWidth, Y: minY + subHeight, Z: minZ + 2*subDepth}})
	// children[5] = inBottom && inRight && inFront
	ot.Children[5] = NewOctree(level, rl.BoundingBox{Min: rl.Vector3{X: minX + subWidth, Y: minY, Z: minZ + subDepth}, Max: rl.Vector3{X: minX + 2*subWidth, Y: minY + subHeight, Z: minZ + 2*subDepth}})
	// children[6] = inTop && inLeft && inFront
	ot.Children[6] = NewOctree(level, rl.BoundingBox{Min: rl.Vector3{X: minX, Y: minY + subHeight, Z: minZ + subDepth}, Max: rl.Vector3{X: minX + subWidth, Y: minY + 2*subHeight, Z: minZ + 2*subDepth}})
	// children[7] = inTop && inRight && inFront
	ot.Children[7] = NewOctree(level, rl.BoundingBox{Min: rl.Vector3{X: minX + subWidth, Y: minY + subHeight, Z: minZ + subDepth}, Max: rl.Vector3{X: minX + 2*subWidth, Y: minY + 2*subHeight, Z: minZ + 2*subDepth}})
}

// Insert inserisce un oggetto nel Octree.
// Insert inserisce un oggetto nel Octree.
func (ot *Octree) Insert(obj *Unit) {

	if ot.Children[0] == nil {
		if len(ot.objects) < maxObjects || ot.level >= maxLevels {
			// Se il nodo corrente ha spazio o abbiamo raggiunto il livello massimo, aggiungi qui.
			ot.objects = append(ot.objects, obj)
			return
		}

		// Se il nodo corrente è pieno e non al livello massimo, dividi.
		ot.Split()
	}

	// Prova ad inserire l'oggetto nei figli.
	indices := ot.getIndices(*obj)
	inserted := false
	for _, index := range indices {
		if index != -1 {
			ot.Children[index].Insert(obj)
			inserted = true
			break // L'oggetto va inserito in un solo figlio, quindi interrompiamo il ciclo.
		}
	}

	// Se l'oggetto non è stato inserito in nessun figlio, aggiungilo a questo nodo.
	if !inserted {
		ot.objects = append(ot.objects, obj)
	}
}

// getIndex determina in quale sotto-Octree un oggetto appartiene.
func (ot *Octree) getIndices(obj Unit) []int {
	// Calcola il punto medio dell'Octree per le tre dimensioni.
	midX := (ot.Bounds.Min.X + ot.Bounds.Max.X) / 2
	midY := (ot.Bounds.Min.Y + ot.Bounds.Max.Y) / 2
	midZ := (ot.Bounds.Min.Z + ot.Bounds.Max.Z) / 2

	// Calcola gli estremi dell'oggetto considerando il suo raggio.
	minX := obj.Position.X - obj.Radius
	maxX := obj.Position.X + obj.Radius
	minY := obj.Position.Y - obj.Radius
	maxY := obj.Position.Y + obj.Radius
	minZ := obj.Position.Z - obj.Radius
	maxZ := obj.Position.Z + obj.Radius

	// Determina la posizione dell'oggetto rispetto al punto medio per ogni dimensione.
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
