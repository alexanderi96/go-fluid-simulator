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

func GetColorFromVelocity(v rl.Vector2) color.RGBA {

	magnitude := math.Sqrt(float64(v.X*v.X + v.Y*v.Y))
	colorFactor := math.Min(1, math.Pow(magnitude, 0.5))

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
