package utils

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"strconv"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func RandomColor() (r, g, b, a uint8) {
	r = uint8(rand.Intn(256))
	g = uint8(rand.Intn(256))
	b = uint8(rand.Intn(256))
	a = 255
	return
}

func RandomRaylibColor() rl.Color {
	r, g, b, a := RandomColor()
	return rl.NewColor(r, g, b, a)
}

// write a function that return a rl.Color from a ginev mass. the hevier the mass, the darker the color
func GetColorFromMass(mass float32) color.RGBA {
	return color.RGBA{
		R: uint8(mass * 255),
		G: uint8(mass * 255),
		B: uint8(mass * 255),
		A: 255,
	}
}

func GetColorFromVelocity(v rl.Vector3) color.RGBA {
	magnitude := math.Sqrt(float64(v.X*v.X + v.Y*v.Y + v.Z*v.Z))

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

func CheckTextFloat32(radMinText string) (float32, error) {
	floatValue, err := strconv.ParseFloat(radMinText, 32)
	if err == nil {
		return float32(floatValue), nil
	} else {
		return 0, fmt.Errorf("error parsing float value: %v", err)
	}
}
