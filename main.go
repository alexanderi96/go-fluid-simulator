package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/AllenDang/giu"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
)

type Coordinates struct {
	X float64
	Y float64
}

type Ball struct {
	pos            Coordinates
	prevPos        Coordinates
	radius         *float32
	restitution    *float32
	density        *float32
	cohesionForce  *float32
	adhesionForce  *float32
	reactToGravity bool
	actionRadius   *float32

	color color.RGBA
}

var (
	sidebarWidth int = 200 // larghezza della barra lateralecheck
	winWidth     int = 1920
	winHeight    int = 1080
	//gameWidth    int = winWidth - sidebarWidth

	isPaused        = false
	isUnstable      = false
	enableSound     = false
	balls           []Ball
	lastTime                = time.Now()
	accumulatedTime float64 = 0.0

	numBalls          int32   = 250 // Nota che è int32
	ballRadius        float32 = 5
	ballRestitution   float32 = 0.9
	ballDensity       float32 = 10
	ballCohesionForce float32 = 200
	ballAdhesionForce float32 = 200
	ballActionRadius  float32 = 15
	wallRestitution   float32 = 0.8
	gravity           float32 = 9.81 // Assegnato a una variabile invece di essere una costante
	mouseForce        float32 = 50
	mouseForceRadius  float32 = 50

	collisionSound  beep.StreamSeekCloser
	collisionBuffer *beep.Buffer

	realDeltaTime float64
	lastFrameTime time.Time
	fps           float64

	mousePos     Coordinates
	isAttracting bool
	isRepulsing  bool
)

func (b *Ball) Mass() float32 {
	volume := (4.0 / 3.0) * math.Pi * math.Pow(float64(*b.radius), 3)
	return *b.density * float32(volume)
}

func (b *Ball) Velocity() Coordinates {
	// Calcola la velocità come differenza tra posizione attuale e precedente
	dx := (b.pos.X - b.prevPos.X) / realDeltaTime
	dy := (b.pos.Y - b.prevPos.Y) / realDeltaTime
	return Coordinates{X: dx, Y: dy}
}

func loadSound() {
	f, err := os.Open("collision.wav") // Sostituisci con il percorso al tuo file audio
	if err != nil {
		panic(err)
	}
	streamer, format, err := wav.Decode(f)
	if err != nil {
		panic(err)
	}
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	collisionBuffer = beep.NewBuffer(format)
	collisionBuffer.Append(streamer)
	streamer.Close()
}

func playCollisionSound() {
	if !enableSound {
		return
	}
	go func() {
		speaker.Play(beep.Seq(collisionBuffer.Streamer(0, collisionBuffer.Len()), beep.Callback(func() {
			log.Println("Collision!")
		})))
	}()
}

func resetBalls() {
	isUnstable = false
	balls = make([]Ball, numBalls)
	ballsPerRow := int(math.Sqrt(float64(numBalls)))
	ballsPerCol := (int(numBalls)-1)/ballsPerRow + 1
	spacing := int(ballRadius*2) + 20 // Nessun spazio extra, solo il doppio del raggio

	// Calcola le dimensioni totali della griglia
	totalGridWidth := ballsPerRow * spacing
	totalGridHeight := ballsPerCol * spacing

	// Verifica se la griglia si adatta nella finestra
	if totalGridWidth > winWidth || totalGridHeight > winHeight {
		// Qui puoi decidere come gestire questo caso: ridimensionare le palle, ridurre il loro numero, ecc.
		fmt.Println("Griglia troppo grande per la finestra!")
		return
	}

	for i := 0; i < int(numBalls); i++ {
		x := (i%ballsPerRow - ballsPerRow/2 + 1) * spacing
		y := (i/ballsPerRow - ballsPerCol/2 + 1) * spacing
		balls[i] = Ball{
			pos:            Coordinates{X: float64(winWidth+x) / 2, Y: float64(winHeight+y) / 2},
			prevPos:        Coordinates{X: float64(winWidth+x) / 2, Y: float64(winHeight+y) / 2},
			radius:         &ballRadius,
			restitution:    &ballRestitution,
			density:        &ballDensity,
			reactToGravity: true,
			cohesionForce:  &ballCohesionForce,
			adhesionForce:  &ballAdhesionForce,
			actionRadius:   &ballActionRadius,
		}
	}

	// Dopo aver inizializzato le palle in resetBalls(), aggiungi questo ciclo per controllare le sovrapposizioni
	for i, ball1 := range balls {
		for j, ball2 := range balls {
			if i >= j {
				continue // Salta il confronto della palla con se stessa o con palle già controllate
			}
			dx := ball2.pos.X - ball1.pos.X
			dy := ball2.pos.Y - ball1.pos.Y
			distance := math.Sqrt(dx*dx + dy*dy)

			// Calcola la velocità delle palline
			velocity1 := ball1.Velocity()
			velocity2 := ball2.Velocity()

			speed1 := math.Sqrt(velocity1.X*velocity1.X + velocity1.Y*velocity1.Y)
			speed2 := math.Sqrt(velocity2.X*velocity2.X + velocity2.Y*velocity2.Y)

			// Controlla la collisione e la velocità
			if distance < float64(*ball1.radius+*ball2.radius) {
				fmt.Println("Collisione rilevata tra palla", i, "e palla", j, "in posizione", ball1.pos, "e", ball2.pos)
				isUnstable = true
			}

			// Soglia di velocità elevata (ad esempio 100.0, modifica secondo necessità)
			highSpeedThreshold := 100.0
			if speed1 > highSpeedThreshold || speed2 > highSpeedThreshold {
				fmt.Println("Errore: Velocità elevata rilevata per palla", i, "o palla", j)
			}
		}
	}
}

// Aggiorna questa funzione per utilizzare l'integrazione di Verlet
func updateBallPositionAndCorrectOverlap(b *Ball) {
	// Calcolo delle accelerazioni (puoi cambiarle in base alle tue necessità)

	// Applica l'integrazione di Verlet per calcolare la nuova posizione
	newX := 2*b.pos.X - b.prevPos.X + realDeltaTime*realDeltaTime
	newY := 2*b.pos.Y - b.prevPos.Y + realDeltaTime*realDeltaTime

	// Aggiorna la posizione precedente e la posizione corrente
	b.prevPos.X = b.pos.X
	b.prevPos.Y = b.pos.Y
	b.pos.X = newX
	b.pos.Y = newY

	// Correzione della sovrapposizione
	for i := range balls {
		if b == &balls[i] {
			continue
		}

		other := &balls[i]
		dx := other.pos.X - b.pos.X
		dy := other.pos.Y - b.pos.Y
		distance := math.Sqrt(dx*dx + dy*dy)
		overlap := float64(*b.radius+*other.radius) - distance

		if overlap > 0 {
			nx := dx / distance
			ny := dy / distance
			correctionX := nx * overlap / 2
			correctionY := ny * overlap / 2

			b.pos.X -= correctionX
			b.pos.Y -= correctionY
			other.pos.X += correctionX
			other.pos.Y += correctionY
		}
	}
}

func (b *Ball) ApplyAdditionalForces() {
	if b.reactToGravity {
		b.pos.Y += 0.5 * float64(gravity) * math.Pow(realDeltaTime, 2)
	}
}

func checkWallCollisions(b *Ball, winWidth, winHeight int) {
	if int(b.pos.X-float64(*b.radius)) < sidebarWidth {
		b.pos.X = float64(sidebarWidth) + float64(*b.radius)
		b.prevPos.X = b.pos.X + (b.pos.X-b.prevPos.X)*float64(*b.restitution)
	} else if int(b.pos.X+float64(*b.radius)) > winWidth {
		b.pos.X = float64(winWidth) - float64(*b.radius)
		b.prevPos.X = b.pos.X + (b.pos.X-b.prevPos.X)*float64(*b.restitution)
	}

	if int(b.pos.Y-float64(*b.radius)) < 0 {
		b.pos.Y = float64(*b.radius)
		b.prevPos.Y = b.pos.Y + (b.pos.Y-b.prevPos.Y)*float64(*b.restitution)
	} else if int(b.pos.Y+float64(*b.radius)) > winHeight {
		b.pos.Y = float64(winHeight) - float64(*b.radius)
		b.prevPos.Y = b.pos.Y + (b.pos.Y-b.prevPos.Y)*float64(*b.restitution)
	}
}

func applyFluidForcesForPair(a, b *Ball) {
	dx := b.pos.X - a.pos.X
	dy := b.pos.Y - a.pos.Y

	distance := math.Sqrt(dx*dx + dy*dy)

	if distance < float64(*a.actionRadius) {
		// Calcola una forza attrattiva che aumenta con la distanza
		cohesion := -*a.cohesionForce * (float32(distance) / *a.actionRadius)

		// Calcola una forza repulsiva che aumenta quando le sfere si avvicinano troppo
		adhesion := *b.adhesionForce * (1 - float32(distance) / *a.actionRadius)

		// Calcola le componenti delle forze
		forceX := float32((dx / distance)) * (cohesion + adhesion)
		forceY := float32((dy / distance)) * (cohesion + adhesion)

		// Applica direttamente alla posizione
		a.pos.X += float64(forceX / a.Mass())
		a.pos.Y += float64(forceY / a.Mass())
		b.pos.X -= float64(forceX / b.Mass())
		b.pos.Y -= float64(forceY / b.Mass())
	}
}

func updateColors() {
	for i := range balls {
		ball := &balls[i]

		velocity := ball.Velocity()                                           // Ottieni le componenti della velocità
		speed := math.Sqrt(math.Pow(velocity.X, 2) + math.Pow(velocity.Y, 2)) // Calcola la magnitudine della velocità
		colorFactor := math.Min(1, math.Pow(speed/200, 0.3))                  // Calcola il colorFactor come prima

		// // Calcola il colore RGB
		// R := uint8(0 + colorFactor*255)         // Da 0 (blu mare) a 255 (bianco)
		// G := uint8(128 + colorFactor*(255-128)) // Da 128 (blu mare) a 255 (bianco)
		// B := uint8(255 + colorFactor*(255-255)) // Da 255 (blu mare) a 255 (bianco)

		// Calcola una scala di colori da blu (freddo, lento) a rosso (caldo, veloce)
		R := uint8(255 * colorFactor)
		G := uint8(0)
		B := uint8(255 * (1 - colorFactor))

		ball.color = color.RGBA{
			R: R,
			G: G,
			B: B,
			A: 255,
		}

	}
}

func Update(winWidth, winHeight int) {
	updateColors()
	var wg sync.WaitGroup
	n := len(balls)
	numGoroutines := runtime.NumCPU()
	chunkSize := n / numGoroutines

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			start := i * chunkSize
			end := start + chunkSize
			for j := start; j < end; j++ {
				ball := &balls[j]
				ball.ApplyAdditionalForces()
				checkWallCollisions(ball, winWidth, winHeight)
				updateBallPositionAndCorrectOverlap(ball)
				for k := j + 1; k < n; k++ {
					//resolveCollision(ball, &balls[k])
					applyFluidForcesForPair(ball, &balls[k])
				}
			}
		}(i)
	}
	wg.Wait()

	// Applica la forza del mouse a tutte le palline
	if isAttracting || isRepulsing {
		for i := 0; i < n; i++ {
			applyMouseForce(&balls[i])
		}
	}
}

// Funzione per generare una nuova pallina in una posizione casuale
func addRandomBall() {
	randX := rand.Intn(winWidth)  // Genera una coordinata X casuale
	randY := rand.Intn(winHeight) // Genera una coordinata Y casuale

	newBall := Ball{
		pos:            Coordinates{X: float64(randX), Y: float64(randY)},
		prevPos:        Coordinates{X: float64(randX), Y: float64(randY)},
		radius:         &ballRadius,
		restitution:    &ballRestitution,
		density:        &ballDensity,
		reactToGravity: true,
		cohesionForce:  &ballCohesionForce,
		adhesionForce:  &ballAdhesionForce,
		actionRadius:   &ballActionRadius,
	}

	balls = append(balls, newBall) // Aggiunge la nuova pallina all'array

}

func removeRandomBall() {

	if len(balls) == 0 {
		return
	}

	randIndex := rand.Intn(len(balls))
	balls = append(balls[:randIndex], balls[randIndex+1:]...)

}

func applyMouseForce(b *Ball) {
	dx := b.pos.X - mousePos.X
	dy := b.pos.Y - mousePos.Y
	distance := math.Sqrt(dx*dx + dy*dy)
	if distance == 0 || float32(distance) > mouseForceRadius**b.radius {
		return
	}

	normalizedX := dx / distance
	normalizedY := dy / distance

	force := mouseForce / float32(math.Pow(distance, 2))

	// Calcola il termine di accelerazione per il realDeltaTime corrente
	ax := float64(force) * normalizedX / float64(b.Mass())
	ay := float64(force) * normalizedY / float64(b.Mass())

	// Aggiorna la posizione della palla in base alla forza del mouse
	if isAttracting {
		b.pos.X -= ax
		b.pos.Y -= ay
	} else if isRepulsing {
		b.pos.X += ax
		b.pos.Y += ay
	}
}

func loop() {
	giu.Update()

	giu.SingleWindow().Layout(
		giu.Custom(func() {

			if int(numBalls) > len(balls) {
				addRandomBall()

			} else if int(numBalls) < len(balls) {
				removeRandomBall()
			}

			if giu.IsMouseDown(giu.MouseButtonLeft) {
				isAttracting = true
				isRepulsing = false
				mousePosition := giu.GetMousePos()
				mousePos = Coordinates{X: float64(mousePosition.X), Y: float64(mousePosition.Y)}
			} else if giu.IsMouseDown(giu.MouseButtonRight) {
				isAttracting = false
				isRepulsing = true
				mousePosition := giu.GetMousePos()
				mousePos = Coordinates{X: float64(mousePosition.X), Y: float64(mousePosition.Y)}
			} else {
				isAttracting = false
				isRepulsing = false
			}

			if giu.IsKeyPressed(giu.KeySpace) {
				isPaused = !isPaused
				if !isPaused {
					lastTime = time.Now() // Reset lastTime when resuming
					accumulatedTime = 0.0 // Reset accumulated time
				} else if isUnstable {
					resetBalls()
				}
			}
			if giu.IsKeyPressed(giu.KeyR) {
				resetBalls()
			}
			if giu.IsKeyPressed(giu.KeyUp) {
				numBalls++
			}
			if giu.IsKeyPressed(giu.KeyDown) && numBalls > 0 {
				numBalls--
			}

			w, h := giu.GetAvailableRegion()
			winWidth = int(w) // sottrai la larghezza della barra laterale
			winHeight = int(h)
			// gameWidth = winWidth - sidebarWidth

			if !isPaused && !isUnstable {
				now := time.Now()
				realDeltaTime = now.Sub(lastTime).Seconds()
				lastTime = now

				Update(winWidth, winHeight)
			}

			canvas := giu.GetCanvas()
			for _, ball := range balls {
				canvas.AddCircleFilled(image.Pt(int(ball.pos.X), int(ball.pos.Y)), float32(*ball.radius), ball.color)
			}

			// Calcola gli FPS
			currentTime := time.Now()
			deltaTime := currentTime.Sub(lastFrameTime).Seconds()
			fps = 1.0 / deltaTime

			// Aggiorna lastFrameTime
			lastFrameTime = currentTime
		}),
		giu.Child().Size(float32(sidebarWidth), float32(winHeight)).Layout(
			// giu.MenuBar().Layout(
			// // giu.Menu("Options").Layout(
			// // 	giu.MenuItem("Show Sidebar").Selected(showSidebar),
			// // ),
			// ),
			giu.Label("Press 'Space' to pause,\n'R' to reset,\n'Up' to add ball,\n'Down' to remove ball"),
			giu.Label(""),
			giu.Label("Settings:"),
			giu.Label("Gravity:"),
			giu.SliderFloat(&gravity, 0, 20),
			giu.Label("Number of balls:"),
			giu.SliderInt(&numBalls, 1, 10000),
			giu.Label("Ball Radius:"),
			giu.SliderFloat(&ballRadius, 1, 50),
			giu.Label(""),
			giu.Label("Wall Restitution:"),
			giu.SliderFloat(&wallRestitution, 0, 2),
			giu.Label("Ball Restitution:"),
			giu.SliderFloat(&ballRestitution, 0, 2),
			giu.Label(""),
			giu.Label("Ball Density:"),
			giu.SliderFloat(&ballDensity, 1, 100),
			giu.Label("Ball Cohesion Force:"),
			giu.SliderFloat(&ballCohesionForce, 0, 1000),
			giu.Label("Ball Adhesion Force:"),
			giu.SliderFloat(&ballAdhesionForce, 0, 1000),
			giu.Label("Ball Action Radius:"),
			giu.SliderFloat(&ballActionRadius, 1, 100),
			giu.Label(""),
			giu.Label("Mouse Force:"),
			giu.SliderFloat(&mouseForce, 1, 1000),
			giu.Label("Mouse Force Radius:"),
			giu.SliderFloat(&mouseForceRadius, 1, 1000),
			giu.Checkbox("Enable Collision Sound:", &enableSound),
			giu.Label(fmt.Sprintf("FPS: %.2f", fps)),
		),
	)

}

func main() {
	f, err := os.Create("mem.out")
	if err != nil {
		log.Fatal("could not create memory profile: ", err)
	}
	defer f.Close() // chiude il file alla fine della funzione main

	runtime.GC() // esegue il garbage collector per ottenere statistiche più accurate
	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Fatal("could not write memory profile: ", err)
	}
	lastFrameTime = time.Now()

	loadSound()
	wnd := giu.NewMasterWindow("Bouncing Ball", int(winWidth), int(winHeight), 0)
	resetBalls()
	wnd.Run(loop)
}
