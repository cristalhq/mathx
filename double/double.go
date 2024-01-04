package double

import (
	"math"
	"math/bits"
)

type D = Double

// Double constants.
var (
	Zero   = Double{hi: 0, lo: 0}
	One    = Double{hi: 1, lo: 0}
	Inf    = Double{hi: math.Inf(1), lo: math.Inf(1)}
	NegInf = Double{hi: -math.Inf(-1), lo: -math.Inf(-1)}
	NaN    = Double{hi: math.NaN(), lo: math.NaN()}
	Pi     = Double{hi: math.Pi, lo: 1.2246467991473532e-16}
	Tau    = Double{hi: 2 * math.Pi, lo: 2.4492935982947064e-16}
	E      = Double{hi: math.E, lo: 1.4456468917292502e-16}
	Log2   = Double{hi: math.Ln2, lo: 2.319046813846299e-17}
	Phi    = Double{hi: math.Phi, lo: -5.432115203682505e-17}
)

// Double float by T.J. Dekker.
//
// See: https://doi.org/10.1007/BF01397083
type Double struct {
	hi, lo float64
	_      struct{}
}

func FromFloat(x float64) Double  { return Double{hi: x} }
func FromSum(a, b float64) Double { return doubleTwoSum(a, b) }
func FromMul(a, b float64) Double { return twoProd(a, b) }
func FromSqr(x float64) Double    { return oneSqr(x) }

func (d Double) ToFloat64() float64  { return d.hi + d.lo }
func (d Double) Equal(x Double) bool { return d.hi == x.hi && d.lo == x.lo }
func (d Double) Add(x Double) Double { return add22(d, x) }
func (d Double) Sub(x Double) Double { return sub22(d, x) }
func (d Double) Mul(x Double) Double { return mul22(d, x) }
func (d Double) Div(x Double) Double { return div22(d, x) }
func (d Double) Abs() Double         { return doubleAbsD(d) }
func (d Double) Neg() Double         { return Double{hi: -d.hi, lo: -d.lo} }
func (d Double) Inv() Double         { return doubleInv2(d) }
func (d Double) Sqr() Double         { return doubleSqr2(d) }
func (d Double) Cmp(x Double) int {
	switch {
	case d.hi == x.hi && d.lo == x.lo:
		return 0
	case d.hi > x.hi || (d.hi == x.hi && d.lo > x.lo):
		return 1
	default:
		return -1
	}
}

func doubleEq21(x Double, f float64) bool { return x.hi == f && x.lo == 0. }
func doubleLe21(x Double, f float64) bool { return x.hi < f || (x.hi == f && x.lo <= 0) }

// x ** 2.
func doubleSqr2(x Double) Double {
	S := oneSqr(x.hi)
	c := x.hi * x.lo
	S.lo += c + c
	return Double{
		hi: S.hi + S.lo,
		lo: S.lo - (x.hi - S.hi),
	}
}

// x ** 0.5
func doubleSqrt2(x Double) Double {
	s := math.Sqrt(x.hi)
	t := oneSqr(s)
	e := (x.hi - t.hi - t.lo + x.lo) * 0.5 / s
	return Double{
		hi: s + e,
		lo: e - (x.hi - s),
	}
}

var padeCoef = []float64{
	1, 272, 36720, 3255840, 211629600, 10666131840, 430200650880, 14135164243200,
	381649434566400, 8481098545920000, 154355993535744030, 2273242813890047700, 26521166162050560000,
	236650405753681870000, 1.5213240369879552e+21, 6.288139352883548e+21, 1.2576278705767096e+22,
}

// e ** x
func doubleExp(x Double) Double {
	if doubleEq21(x, 0) {
		return One
	}
	if doubleEq21(x, 1) {
		return E
	}
	n := math.Floor(x.hi/Log2.hi + 0.5)
	x = sub22(x, doubleMulDF(Log2, n))
	U := One
	V := One
	for i, cLen := 0, len(padeCoef); i < cLen; i++ {
		U = addDF(mul22(U, x), padeCoef[i])
	}
	for i, cLen := 0, len(padeCoef); i < cLen; i++ {
		s := -1.0
		if i%2 == 0 {
			s = 1.0
		}
		V = addDF(mul22(V, x), padeCoef[i]*s)
	}
	x = doubleMulDFpow2(div22(U, V), int(n))
	return x
}

func doubleLn2(x Double) Double {
	switch {
	case doubleLe21(x, 0):
		return NegInf
	case doubleEq21(x, 1):
		return Zero
	default:
		Z := FromFloat(math.Log(x.hi))
		return doubleSubDF(add22(mul22(x, doubleExp(Z.Neg())), Z), 1)
	}
}

func doubleSinh2(x Double) Double {
	exp := doubleExp(x)
	return doubleMulDFpow2(sub22((exp), doubleInv2(exp)), -1)
}

func doubleCosh2(x Double) Double {
	exp := doubleExp(x)
	return doubleMulDFpow2(add22((exp), doubleInv2(exp)), -1)
}

func doublePow22(base Double, exp Double) Double {
	return doubleExp(mul22(doubleLn2(base), exp))
}

const splitter = 1<<27 + 1 // Veltkamp’s splitter

func doubleTwoSum(a, b float64) Double {
	s := a + b
	a1 := s - b
	return Double{
		hi: s,
		lo: (a - a1) + (b - (s - a1)),
	}
}

func twoProd(a, b float64) Double {
	t := splitter * a
	ah := t + (a - t)
	al := a - ah
	t = splitter * b
	bh := t + (b - t)
	bl := b - bh
	t = a * b
	return Double{
		hi: t,
		lo: ((ah*bh - t) + ah*bl + al*bh) + al*bl,
	}
}

func oneSqr(a float64) Double {
	t := splitter * a
	ah := t + (a - t)
	al := a - ah
	t = a * a
	hl := al * ah
	return Double{
		hi: t,
		lo: ((ah*ah - t) + hl + hl) + al*al,
	}
}

func add22(x, y Double) Double {
	s := doubleTwoSum(x.hi, y.hi)
	e := doubleTwoSum(x.lo, y.lo)
	c := s.lo + e.hi
	vh := s.hi + c
	vl := c - (vh - s.hi)
	c = vl + e.lo
	return Double{
		hi: vh + c,
		lo: c - (x.hi - vh),
	}
}

func sub22(x, y Double) Double {
	s := doubleTwoSum(x.hi, -y.hi)
	e := doubleTwoSum(x.lo, -y.lo)
	c := s.lo + e.hi
	vh := s.hi + c
	vl := c - (vh - s.hi)
	c = vl + e.lo
	return Double{
		hi: vh + c,
		lo: c - (x.hi - vh),
	}
}

func mul22(x, y Double) Double {
	s := twoProd(x.hi, y.hi)
	s.lo += x.hi*y.lo + x.lo*y.hi
	return Double{
		hi: s.hi + s.lo,
		lo: s.lo - (x.hi - s.hi),
	}
}

func div22(x, y Double) Double {
	s := x.hi / y.hi
	t := twoProd(s, y.hi)
	e := ((((x.hi - t.hi) - t.lo) + x.lo) - s*y.lo) / y.hi
	return Double{
		hi: s + e,
		lo: e - (x.hi - s),
	}
}

// x + f
func addDF(x Double, f float64) Double {
	s := doubleTwoSum(x.hi, f)
	s.lo += x.lo
	return Double{
		hi: s.hi + s.lo,
		lo: s.lo - (x.hi - s.hi),
	}
}

// x - f
func doubleSubDF(x Double, f float64) Double {
	s := doubleTwoSum(x.hi, -f)
	s.lo += x.lo
	return Double{
		hi: s.hi + s.lo,
		lo: s.lo - (x.hi - s.hi),
	}
}

// x * f
func doubleMulDF(x Double, f float64) Double {
	c := twoProd(x.hi, f)
	cl := x.lo * f
	th := c.hi + cl
	x.lo = cl - (th - c.hi)
	cl = x.lo + c.lo
	return Double{
		hi: th + cl,
		lo: cl - (x.hi - th),
	}
}

// x / f
func doubleDivDF(x Double, f float64) Double {
	th := x.hi / f
	p := twoProd(th, f)
	d := doubleTwoSum(x.hi, -p.hi)
	tl := (d.hi + (d.lo + (x.lo - p.lo))) / f
	return Double{
		hi: th + tl,
		lo: tl - (x.hi - th),
	}
}

// |x|
func doubleAbsD(x Double) Double {
	if x.hi < 0 {
		return x.Neg()
	}
	return x
}

// 1 / x
func doubleInv2(x Double) Double {
	xh := x.hi
	s := 1 / xh
	x = doubleMulDF(x, s)
	zl := (1 - x.hi - x.lo) / xh
	return Double{
		hi: s + zl,
		lo: zl - (x.hi - s),
	}
}

// x * 2 ** n
func doubleMulDFpow2(x Double, n int) Double {
	if n < 0 {
		n = -n
	}
	c := float64(int(1) << n)
	if n < 0 {
		c = 1 / c
	}
	x.hi = x.hi * c
	x.lo = x.lo * c
	return x
}

// x ** n
func pow2n(x Double, n int) Double {
	if n == 0 {
		return One
	}
	if n == 1 {
		return x
	}
	isPositive := n > 0
	if !isPositive {
		n = -n
	}
	i := 31 - bits.LeadingZeros32(uint32(n|1))
	j := math.Floor(float64(n - (1 << i)))
	x0 := x
	for ; i != 0; i-- {
		x = doubleSqr2(x)
	}
	for ; j != 0; j-- {
		x = mul22(x, x0)
	}
	if isPositive {
		return x
	}
	return doubleInv2(x)
}
