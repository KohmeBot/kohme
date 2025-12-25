package metric

import (
	"iter"
	"slices"
)

type traceMinHeap []*trace

func (h *traceMinHeap) Len() int {
	vh := *h

	return len(vh)
}

func (h *traceMinHeap) Less(i, j int) bool {
	vh := *h
	// 时间越早，优先级越高（最小堆）
	return vh[i].t.Before(vh[j].t)
}

func (h *traceMinHeap) Swap(i, j int) {
	vh := *h
	vh[i], vh[j] = vh[j], vh[i]
}

func (h *traceMinHeap) Push(x any) {
	*h = append(*h, x.(*trace))
}

func (h *traceMinHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[:n-1]
	return item
}

func (h *traceMinHeap) Clip() {
	*h = slices.Clip(*h)
}
func (h *traceMinHeap) Range() iter.Seq2[int, *trace] {
	return func(yield func(int, *trace) bool) {
		for i, v := range *h {
			if !yield(i, v) {
				return
			}
		}
	}
}
