package metric

import (
	"container/heap"
	"github.com/jhue58/latency/buckets"
	"github.com/jhue58/latency/duration"
	"github.com/jhue58/latency/recorder"
	"sync"
	"time"
)

type trace struct {
	id      string
	command string
	t       time.Time
}

type Metric struct {
	Name         string
	StartTime    time.Time
	BootDuration duration.Duration
	rmp          map[string]recorder.Recorder
	traces       *traceMinHeap
	mu           sync.Mutex
}

func NewMetric(name string) *Metric {
	return &Metric{
		Name:      name,
		StartTime: time.Now(),
		rmp:       make(map[string]recorder.Recorder),
		traces:    &traceMinHeap{},
		mu:        sync.Mutex{},
	}
}

func (m *Metric) SetBootDuration(dur time.Duration) {
	m.BootDuration = duration.NewDuration(dur)
}

func (m *Metric) CommandInit(commands ...string) {
	for _, command := range commands {
		m.rmp[command] = buckets.NewBucketsRecorder()
	}

}

func (m *Metric) Cleanup() {
	for _, r := range m.rmp {
		r.Cleanup()
	}
	m.traces = &traceMinHeap{}
}

func (m *Metric) Start() {
	m.StartTime = time.Now()
}

func (m *Metric) CommandStart(id, command string) {
	_, ok := m.rmp[command]
	if !ok {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	heap.Push(m.traces, &trace{
		id:      id,
		command: command,
		t:       time.Now(),
	})
}

func (m *Metric) CommandEnd(id string) {

	now := time.Now()
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, tr := range m.traces.Range() {
		if tr.id != id {
			continue
		}
		dur := now.Sub(tr.t)
		if r, ok := m.rmp[tr.command]; ok {
			r.Record(duration.NewDuration(dur))
		}
		heap.Remove(m.traces, i)
		break
	}

}

func (m *Metric) Snapshot() map[string]recorder.RecordedSnapshot {
	res := make(map[string]recorder.RecordedSnapshot, len(m.rmp))

	for cmd, r := range m.rmp {
		snap := recorder.RecordedSnapshot{}
		r.Snapshot(snap)
		res[cmd] = snap
	}
	return res
}
