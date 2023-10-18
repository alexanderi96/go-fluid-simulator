package metrics

import (
	"runtime"
	"sync"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Metrics struct to hold performance metrics.
type Metrics struct {
	Mu            sync.Mutex // Mutex for goroutine safety
	Frametime     float32
	FPS           int32
	HeapSize      uint32
	CPUUsage      float32
	GPUUsage      float32
	DrawCalls     uint32
	Latency       float32
	ActiveThreads uint32
	DiskUsage     uint32
	NetworkUsage  uint32
}

// New creates a new instance of Metrics.
func New() *Metrics {
	return &Metrics{}
}

// UpdateMetrics updates all metrics at once.
func (m *Metrics) Update() {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	// Update memory usage
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	m.HeapSize = uint32(memStats.HeapAlloc / 1024) // Convert bytes to kilobytes

	// Update frame processing time
	m.Frametime = rl.GetFrameTime()
	m.FPS = rl.GetFPS()

	// ... rest of your code ...
}
