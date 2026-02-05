package app

import "sync"

type Gate struct {
	closing bool
	wg      sync.WaitGroup
	mu      sync.Mutex
}

func (g *Gate) Close() {
	g.mu.Lock()
	g.closing = true
	g.mu.Unlock()
}

func (g *Gate) IsClosing() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.closing
}

func (g *Gate) WaitFinish() {
	g.wg.Wait()
}

func (g *Gate) Open() {
	g.mu.Lock()
	g.closing = false
	g.mu.Unlock()
}

func (g *Gate) Add() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closing {
		return false
	}

	g.wg.Add(1)
	return true
}

func (g *Gate) Done() {
	g.wg.Done()
}
