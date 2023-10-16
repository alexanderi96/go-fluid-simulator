package physics

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	maxObjects = 4
	maxLevels  = 2
)

type Quadtree struct {
	level    int
	Bounds   rl.Rectangle
	objects  []*Unit
	Children [4]*Quadtree
}

// NewQuadtree crea un nuovo Quadtree.
func NewQuadtree(level int, bounds rl.Rectangle) *Quadtree {
	return &Quadtree{
		level:  level,
		Bounds: bounds,
	}
}

// Clear pulisce il quadtree.
func (qt *Quadtree) Clear() {
	qt.objects = qt.objects[:0]
	for i := 0; i < 4; i++ {
		if qt.Children[i] != nil {
			qt.Children[i].Clear()
			qt.Children[i] = nil
		}
	}
}

// Split divide il quadtree in quattro sotto-quadtrees.
func (qt *Quadtree) Split() {
	subWidth := qt.Bounds.Width / 2
	subHeight := qt.Bounds.Height / 2
	x := qt.Bounds.X
	y := qt.Bounds.Y

	qt.Children[0] = NewQuadtree(qt.level+1, rl.NewRectangle(x+subWidth, y, subWidth, subHeight))
	qt.Children[1] = NewQuadtree(qt.level+1, rl.NewRectangle(x, y, subWidth, subHeight))
	qt.Children[2] = NewQuadtree(qt.level+1, rl.NewRectangle(x, y+subHeight, subWidth, subHeight))
	qt.Children[3] = NewQuadtree(qt.level+1, rl.NewRectangle(x+subWidth, y+subHeight, subWidth, subHeight))
}

// Insert inserisce un oggetto nel quadtree.
func (qt *Quadtree) Insert(obj *Unit) {
	// Se i figli sono nil, dividere il quadtree
	if qt.Children[0] == nil && qt.Children[1] == nil && qt.Children[2] == nil && qt.Children[3] == nil {
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

// getIndex determina in quale sotto-quadtree un oggetto appartiene.
func (qt *Quadtree) getIndex(obj Unit) int {
	midHorizontal := qt.Bounds.X + qt.Bounds.Width/2
	midVertical := qt.Bounds.Y + qt.Bounds.Height/2

	topQuadrant := obj.Position.Y < midVertical && obj.Position.Y+obj.Radius*2 < midVertical
	bottomQuadrant := obj.Position.Y > midVertical

	if obj.Position.X < midHorizontal && obj.Position.X+obj.Radius*2 < midHorizontal {
		if topQuadrant {
			return 1
		} else if bottomQuadrant {
			return 2
		}
	} else if obj.Position.X > midHorizontal {
		if topQuadrant {
			return 0
		} else if bottomQuadrant {
			return 3
		}
	}

	return -1
}

// Retrieve restituisce tutti gli oggetti che potrebbero collidere con l'oggetto dato.
func (qt *Quadtree) Retrieve(returnObjects *[]*Unit, obj *Unit) {
	if qt.Children[0] != nil || qt.Children[1] != nil || qt.Children[2] != nil || qt.Children[3] != nil {
		index := qt.getIndex(*obj)
		if index != -1 && qt.Children[index] != nil { // Aggiunto controllo per nil
			qt.Children[index].Retrieve(returnObjects, obj)
		} else {
			for i := 0; i < 4; i++ {
				if qt.Children[i] != nil { // Aggiunto controllo per nil
					qt.Children[i].Retrieve(returnObjects, obj)
				}
			}
		}
	}

	*returnObjects = append(*returnObjects, qt.objects...)
}
