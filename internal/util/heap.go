package util

import (
	"bytes"
	"github.com/google/pprof/profile"
	"runtime/pprof"
	"sync"
)

var bufPool = sync.Pool{New: func() any { return new(bytes.Buffer) }}

func ParseHeap() (map[string]int64, error) {
	buf := bufPool.Get().(*bytes.Buffer)
	defer buf.Reset()

	err := pprof.WriteHeapProfile(buf)
	if err != nil {
		return nil, err
	}
	p, err := profile.Parse(buf)
	if err != nil {
		return nil, err
	}

	fileAlloc := map[string]int64{}
	for _, s := range p.Sample {
		value := s.Value[0] // alloc_space 或 inuse_space，取决于 profile.Type
		for _, loc := range s.Location {
			for _, line := range loc.Line {
				fn := line.Function
				if fn == nil {
					continue
				}
				fileAlloc[fn.Filename] += value
			}
		}
	}

	return fileAlloc, nil
}
