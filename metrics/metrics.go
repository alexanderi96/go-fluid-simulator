package metrics

import (
	"runtime"
	"sync"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Metrics struct {
	Mu            sync.Mutex
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

func New() *Metrics {
	return &Metrics{}
}

func (m *Metrics) Update() {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	m.HeapSize = uint32(memStats.HeapAlloc / 1024)

	m.Frametime = rl.GetFrameTime()
	m.FPS = rl.GetFPS()

}
