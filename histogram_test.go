package mathx

import (
	"math"
	"sync"
	"testing"
)

func TestHistogramUnderflow(t *testing.T) {
	h := NewHistogram()

	q := h.Quantile(0.5)
	if !math.IsNaN(q) {
		t.Fatalf("unexpected quantile for empty histogram; got %v; want %v", q, NaN)
	}

	for i := 0; i < maxSamples; i++ {
		h.Update(float64(i))
	}
	qs := h.Quantiles(nil, []float64{0, 0.5, 1})
	if qs[0] != 0 {
		t.Fatalf("unexpected quantile value for phi=0; got %v; want %v", qs[0], 0)
	}
	if qs[1] != maxSamples/2 {
		t.Fatalf("unexpected quantile value for phi=0.5; got %v; want %v", qs[1], maxSamples/2)
	}
	if qs[2] != maxSamples-1 {
		t.Fatalf("unexpected quantile value for phi=1; got %v; want %v", qs[2], maxSamples-1)
	}
}

func TestHistogramOverflow(t *testing.T) {
	h := NewHistogram()

	for i := 0; i < maxSamples*10; i++ {
		h.Update(float64(i))
	}
	qs := h.Quantiles(nil, []float64{0, 0.5, 0.9999, 1})
	if qs[0] != 0 {
		t.Fatalf("unexpected quantile value for phi=0; got %v; want %v", qs[0], 0)
	}

	median := float64(maxSamples*10-1) / 2
	if qs[1] < median*0.9 || qs[1] > median*1.1 {
		t.Fatalf("unexpected quantile value for phi=0.5; got %v; want %v", qs[1], median)
	}
	if qs[2] < maxSamples*10*0.9 {
		t.Fatalf("unexpected quantile value for phi=0.9999; got %v; want %v", qs[2], maxSamples*10*0.9)
	}
	if qs[3] != maxSamples*10-1 {
		t.Fatalf("unexpected quantile value for phi=1; got %v; want %v", qs[3], maxSamples*10-1)
	}

	q := h.Quantile(NaN)
	if !math.IsNaN(q) {
		t.Fatalf("unexpected value for phi=NaN; got %v; want %v", q, NaN)
	}
}

func TestHistogramRepeatableResults(t *testing.T) {
	h := NewHistogram()

	for i := 0; i < maxSamples*10; i++ {
		h.Update(float64(i))
	}
	q1 := h.Quantile(0.95)

	for j := 0; j < 10; j++ {
		h.Reset()
		for i := 0; i < maxSamples*10; i++ {
			h.Update(float64(i))
		}
		q2 := h.Quantile(0.95)
		if q2 != q1 {
			t.Fatalf("unexpected quantile value; got %g; want %g", q2, q1)
		}
	}
}

var sink float64
var sinkLock sync.Mutex

func BenchmarkHistogramUpdate(b *testing.B) {
	b.ReportAllocs()
	b.SetBytes(1)
	b.RunParallel(func(pb *testing.PB) {
		h := NewHistogram()
		var v float64
		for pb.Next() {
			h.Update(v)
			v += 1.5
		}
		sinkLock.Lock()
		sink += h.Quantile(0.5)
		sinkLock.Unlock()
	})
}

func BenchmarkHistogramMerged(b *testing.B) {
	b.ReportAllocs()
	b.SetBytes(1)

	const size = 10
	var wg sync.WaitGroup
	wg.Add(size)

	hists := make([]*Histogram, size)
	for g := 0; g < size; g++ {
		g := g
		hists[g] = NewHistogram()

		go func() {
			defer wg.Done()

			v := 1.5
			for i := 0; i < b.N; i++ {
				hists[g].Update(v)
				v += 1.5
			}
		}()
	}

	wg.Wait()

	merged := MergeHistograms(hists)
	if quant := merged.Quantile(0.5); quant == 0 {
		b.Fatal("quant must be non-zero")
	}
}
