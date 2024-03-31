package physics

import (
    "github.com/gen2brain/raylib-go/raylib"
)

// Punto 3D
type Vector3 struct {
    X, Y, Z float32
}

// Nodo dell'Octree
type OctreeNode struct {
    Boundary rl.BoundingBox // Il volume dello spazio rappresentato dal nodo
    Children [8]*OctreeNode // I figli dell'octree (ottanti)
    Units    []*Unit        // Le unità di fluido contenute nel nodo (solo per i nodi foglia)
    IsLeaf   bool           // Indica se il nodo è una foglia
}

// Unità di fluido
type Unit struct {
    Position Vector3 // Posizione dell'unità di fluido
    // Altri attributi dell'unità di fluido, se necessario
}

// Metodo per creare un nuovo nodo dell'Octree
func NewOctreeNode(boundary rl.BoundingBox) *OctreeNode {
    return &OctreeNode{
        Boundary: boundary,
        IsLeaf:   true,
    }
}

// Funzione per creare un nuovo Octree
func BuildOctree(boundary rl.BoundingBox, units []*Unit, depth int) *OctreeNode {
    if depth == 0 || len(units) <= 1 {
        // Crea un nodo foglia
        return &OctreeNode{Boundary: boundary, Units: units, IsLeaf: true}
    }
    // Suddividi lo spazio e distribuisci le unità nei corrispondenti ottanti
    // Implementa la suddivisione dello spazio e la distribuzione delle unità nei figli
    // ...
    return nil
}

// Metodo per aggiornare l'Octree
func (node *OctreeNode) UpdateOctree() {
    if node.IsLeaf {
        // Verifica se le unità sono ancora all'interno del volume del nodo
        // Implementa il controllo per le unità all'interno del nodo foglia
        // ...
    } else {
        // Aggiorna ricorsivamente gli ottanti
        for _, child := range node.Children {
            if child != nil {
                child.UpdateOctree()
            }
        }
    }
}

// Metodo per controllare le collisioni all'interno dell'Octree
func (node *OctreeNode) CheckCollisions() {
    if node.IsLeaf {
        // Controlla le collisioni tra le unità all'interno del nodo foglia
        // Implementa la logica per il controllo delle collisioni tra le unità all'interno del nodo foglia
        // ...
    } else {
        // Controlla le collisioni nei sottoalberi
        for _, child := range node.Children {
            if child != nil {
                child.CheckCollisions()
            }
        }
    }
}

// Metodo per inserire un oggetto nel nodo dell'Octree
func (node *OctreeNode) Insert(object *Unit) {
    // Se il nodo è una foglia, aggiungi l'oggetto ai suoi contenuti
    if node.IsLeaf {
        node.Units = append(node.Units, object)
        // Controllo se il nodo deve essere diviso
        if len(node.Units) > MaxUnitsPerNode {
            node.Split()
        }
    } else {
        // Se il nodo non è una foglia, inserisci l'oggetto nei figli appropriati
        for _, child := range node.Children {
            if child.Boundary.CheckCollisionSphere(object.Position, 1) {
                child.Insert(object)
                break
            }
        }
    }
}

// Metodo per dividere il nodo dell'Octree in 8 figli
func (node *OctreeNode) Split() {
    // Implementa la suddivisione del nodo in 8 figli
    // ...
}

// Costante per il numero massimo di unità per nodo prima della suddivisione
const MaxUnitsPerNode = 8
