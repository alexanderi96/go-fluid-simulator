package metrics

import (
	"runtime"
	"sync"
	// rl "github.com/gen2brain/raylib-go/raylib"
)

type Metrics struct {
	Mu            sync.Mutex
	RealFrametime float64
	SimDuration   float64
	FPS           int32
	HeapSize      uint32
	CPUUsage      float64
	GPUUsage      float64
	DrawCalls     uint32
	Latency       float64
	ActiveThreads uint32
	DiskUsage     uint32
	NetworkUsage  uint32
}

func New() *Metrics {
	return &Metrics{
		RealFrametime: 0,
		SimDuration:   0,
		FPS:           0,
		HeapSize:      0,
		CPUUsage:      0,
		GPUUsage:      0,
		DrawCalls:     0,
		Latency:       0,
		ActiveThreads: 0,
		DiskUsage:     0,
		NetworkUsage:  0,
	}
}

func (m *Metrics) Update(dt float64) {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	m.HeapSize = uint32(memStats.HeapAlloc / 1024)
	m.SimDuration += dt

	// m.RealFrametime = float64(rl.GetFrameTime())
	// m.FPS = rl.GetFPS()

}
