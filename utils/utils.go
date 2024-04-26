package utils

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"sort"
	"strconv"

	"github.com/EliCDavis/vector/vector3"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	// rl "github.com/gen2brain/raylib-go/raylib"
)

// Mappatura dei valori Kelvin a colori RGBA.
// Questi valori sono esempi e possono essere adattati o dettagliati meglio.
var kelvinToRGB = map[float64]color.RGBA{
	1000:  {255, 56, 0, 255},
	2000:  {255, 137, 18, 255},
	3000:  {255, 180, 107, 255},
	4000:  {255, 209, 163, 255},
	5000:  {255, 228, 206, 255},
	6000:  {255, 242, 239, 255},
	7000:  {245, 243, 255, 255},
	8000:  {235, 238, 255, 255},
	9000:  {227, 233, 255, 255},
	10000: {220, 229, 255, 255},
}

func RandomColor() (r, g, b, a uint8) {
	r = uint8(rand.Intn(256))
	g = uint8(rand.Intn(256))
	b = uint8(rand.Intn(256))
	a = 255
	return
}

// func RandomRaylibColor() rl.Color {
// 	r, g, b, a := RandomColor()
// 	return rl.NewColor(r, g, b, a)
// }

// write a function that return a rl.Color from a ginev mass. the hevier the mass, the darker the color
func GetColorFromMass(mass float32) color.RGBA {
	return color.RGBA{
		R: uint8(mass * 255),
		G: uint8(mass * 255),
		B: uint8(mass * 255),
		A: 255,
	}
}

func GetColorFromVelocity(v vector3.Vector[float64]) color.RGBA {
	magnitude := math.Sqrt(float64(v.X()*v.X() + v.Y()*v.Y() + v.Z()*v.Z()))

	// Aggiungi 1 a magnitude per evitare log(0)
	// Utilizza una costante k per controllare la velocità della transizione verso il rosso
	k := 0.15
	colorFactor := math.Min(1, k*math.Log(magnitude+1))

	R := uint8(255 * colorFactor)
	G := uint8(0)
	B := uint8(255 * (1 - colorFactor))

	return color.RGBA{
		R: R,
		G: G,
		B: B,
		A: 255,
	}
}

func CheckTextFloat64(radMinText string) (float64, error) {
	floatValue, err := strconv.ParseFloat(radMinText, 32)
	if err == nil {
		return floatValue, nil
	} else {
		return 0, fmt.Errorf("error parsing float value: %v", err)
	}
}

// func ToRlVector3(v vector3.Vector[float64]) rl.Vector3 {
// 	return rl.Vector3{
// 		X: float32(v.X()),
// 		Y: float32(v.Y()),
// 		Z: float32(v.Z()),
// 	}
// }

// func ToVector3FromRlVector3(v rl.Vector3) vector3.Vector[float64] {
// 	return vector3.New(float64(v.X), float64(v.Y), float64(v.Z))
// }

// func ToRlBoundingBox(min, max vector3.Vector[float64]) rl.BoundingBox {
// 	return rl.BoundingBox{
// 		Min: ToRlVector3(min),
// 		Max: ToRlVector3(max),
// 	}
// }

// KelvinToRGBA converte un valore di temperatura Kelvin in un colore RGBA.
func KelvinToRGBA(kelvin float64) color.RGBA {

	// Trova i due valori Kelvin più vicini.
	keys := make([]float64, 0, len(kelvinToRGB))
	for k := range kelvinToRGB {
		keys = append(keys, k)
	}
	sort.Float64s(keys)

	var lower, upper float64
	for i, k := range keys {
		if kelvin < k {
			if i == 0 {
				return kelvinToRGB[keys[0]] // Se è minore del valore più basso, restituisci il valore più basso.
			}
			lower = keys[i-1]
			upper = k
			break
		}
	}
	if kelvin >= keys[len(keys)-1] {
		return kelvinToRGB[keys[len(keys)-1]] // Se è maggiore del valore più alto, restituisci il valore più alto.
	}

	// Calcola il fattore di interpolazione tra i due valori Kelvin.
	factor := (kelvin - lower) / (upper - lower)

	// Interpola linearmente i valori RGBA.
	lowerColor := kelvinToRGB[lower]
	upperColor := kelvinToRGB[upper]
	blend := func(a, b uint8) uint8 {
		return a + uint8(float64(b-a)*factor)
	}
	return color.RGBA{
		R: blend(lowerColor.R, upperColor.R),
		G: blend(lowerColor.G, upperColor.G),
		B: blend(lowerColor.B, upperColor.B),
		A: 255, // Alpha è sempre 255 (opaco).
	}
}

func RgbaToMath32(rgba color.RGBA) *math32.Color {
	return &math32.Color{
		R: float32(rgba.R) / 255.0,
		G: float32(rgba.G) / 255.0,
		B: float32(rgba.B) / 255.0,
	}
}

func GetBoundsLine(min, max vector3.Vector[float64]) *graphic.Lines {

	// Crea il wireframe per il nodo corrente
	vertices := math32.NewArrayF32(0, 16)
	vertices.Append(
		min.ToFloat32().X(), min.ToFloat32().Y(), min.ToFloat32().Z(), max.ToFloat32().X(), min.ToFloat32().Y(), min.ToFloat32().Z(),
		max.ToFloat32().X(), min.ToFloat32().Y(), min.ToFloat32().Z(), max.ToFloat32().X(), max.ToFloat32().Y(), min.ToFloat32().Z(),
		max.ToFloat32().X(), max.ToFloat32().Y(), min.ToFloat32().Z(), min.ToFloat32().X(), max.ToFloat32().Y(), min.ToFloat32().Z(),
		min.ToFloat32().X(), max.ToFloat32().Y(), min.ToFloat32().Z(), min.ToFloat32().X(), min.ToFloat32().Y(), min.ToFloat32().Z(),
		min.ToFloat32().X(), min.ToFloat32().Y(), max.ToFloat32().Z(), max.ToFloat32().X(), min.ToFloat32().Y(), max.ToFloat32().Z(),
		max.ToFloat32().X(), min.ToFloat32().Y(), max.ToFloat32().Z(), max.ToFloat32().X(), max.ToFloat32().Y(), max.ToFloat32().Z(),
		max.ToFloat32().X(), max.ToFloat32().Y(), max.ToFloat32().Z(), min.ToFloat32().X(), max.ToFloat32().Y(), max.ToFloat32().Z(),
		min.ToFloat32().X(), max.ToFloat32().Y(), max.ToFloat32().Z(), min.ToFloat32().X(), min.ToFloat32().Y(), max.ToFloat32().Z(),
		min.ToFloat32().X(), min.ToFloat32().Y(), min.ToFloat32().Z(), min.ToFloat32().X(), min.ToFloat32().Y(), max.ToFloat32().Z(),
		max.ToFloat32().X(), min.ToFloat32().Y(), min.ToFloat32().Z(), max.ToFloat32().X(), min.ToFloat32().Y(), max.ToFloat32().Z(),
		max.ToFloat32().X(), max.ToFloat32().Y(), min.ToFloat32().Z(), max.ToFloat32().X(), max.ToFloat32().Y(), max.ToFloat32().Z(),
		min.ToFloat32().X(), max.ToFloat32().Y(), min.ToFloat32().Z(), min.ToFloat32().X(), max.ToFloat32().Y(), max.ToFloat32().Z(),
	)
	vbo := gls.NewVBO(vertices).AddAttrib(gls.VertexPosition)

	geom := geometry.NewGeometry()
	geom.AddVBO(vbo)
	mat := material.NewStandard(math32.NewColor("White"))
	mat.SetLineWidth(1.5)
	mat.SetSide(material.SideDouble)
	return graphic.NewLines(geom, mat)
}
