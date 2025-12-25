package metric

import (
	"container/heap"
	"strconv"
	"testing"
	"time"
)

func TestHeap(t *testing.T) {
	h := &traceMinHeap{}
	heap.Init(h)

	for i := 0; i < 1000; i++ {
		time.Sleep(time.Millisecond)
		tr := &trace{
			id:      strconv.Itoa(i),
			command: strconv.Itoa(i),
			t:       time.Now(),
		}
		heap.Push(h, tr)
	}
	l := h.Len()
	for i := 0; i < l; i++ {
		tr := heap.Pop(h).(*trace)
		t.Log(tr.id)
	}

	h.Clip()

}
