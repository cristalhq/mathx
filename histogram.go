package mathx

import (
	"math"
	"sort"

	"github.com/valyala/fastrand"
)

const maxSamples = 1000

// Histogram for floats.
// Based on https://github.com/valyala/histogram with small changes.
type Histogram struct {
	max   float64
	min   float64
	count uint64

	vals []float64
	tmp  []float64
	rng  fastrand.RNG
}

// NewHistogram returns new Histogram histogram.
func NewHistogram() *Histogram {
	h := &Histogram{}
	h.Reset()
	return h
}

// Reset resets the histogram.
func (h *Histogram) Reset() {
	h.max = InfNeg
	h.min = InfPos
	h.count = 0

	if len(h.vals) > 0 {
		h.vals = h.vals[:0]
		h.tmp = h.tmp[:0]
	} else {
		// Free up memory occupied by unused histogram.
		h.vals = nil
		h.tmp = nil
	}

	// Reset rng state in order to get repeatable results
	// for the same sequence of values passed to Histogram.Update.
	// See https://github.com/VictoriaMetrics/VictoriaMetrics/issues/1612
	h.rng.Seed(1)
}

// Update the histogram with v.
func (h *Histogram) Update(v float64) {
	if v > h.max {
		h.max = v
	}
	if v < h.min {
		h.min = v
	}

	h.count++
	if len(h.vals) < maxSamples {
		h.vals = append(h.vals, v)
	} else {
		n := int(h.rng.Uint32n(uint32(h.count)))
		if n < len(h.vals) {
			h.vals[n] = v
		}
	}
}

// Quantile returns the quantile value for the given phi.
func (h *Histogram) Quantile(phi float64) float64 {
	h.tmp = append(h.tmp[:0], h.vals...)
	sort.Float64s(h.tmp)
	return h.quantile(phi)
}

// Quantiles appends quantile values to dst for the given phis.
func (h *Histogram) Quantiles(dst, phis []float64) []float64 {
	h.tmp = append(h.tmp[:0], h.vals...)
	sort.Float64s(h.tmp)
	return h.quantiles(dst, phis)
}

func (h *Histogram) quantiles(dst, phis []float64) []float64 {
	for _, phi := range phis {
		q := h.quantile(phi)
		dst = append(dst, q)
	}
	return dst
}

func (h *Histogram) quantile(phi float64) float64 {
	switch {
	case len(h.tmp) == 0 || math.IsNaN(phi):
		return NaN
	case phi <= 0:
		return h.min
	case phi >= 1:
		return h.max
	default:
		idx := uint(phi*float64(len(h.tmp)-1) + 0.5)
		if idx >= uint(len(h.tmp)) {
			idx = uint(len(h.tmp) - 1)
		}
		return h.tmp[idx]
	}
}

// MergeHistograms returns 1 histogram built from the given.
func MergeHistograms(hs []*Histogram) *Histogram {
	n := 0
	for _, h := range hs {
		n += len(h.vals)
	}

	t := NewHistogram()
	t.vals = make([]float64, 0, n)

	for _, h := range hs {
		t.vals = append(t.vals, h.vals...)
		if t.max < h.max {
			t.max = h.max
		}
		if t.min > h.min {
			t.min = h.min
		}
	}
	return t
}
