package decimal

import (
	"math"
	"math/bits"
)

// Double constants
var (
	DoubleZero   = Double{hi: 0., lo: 0.}
	DoubleOne    = Double{hi: 1., lo: 0.}
	DoubleInf    = Double{hi: math.Inf(1), lo: math.Inf(1)}
	DoubleNegInf = Double{hi: -math.Inf(-1), lo: -math.Inf(-1)}
	DoubleNaN    = Double{hi: math.NaN(), lo: math.NaN()}
	DoublePi     = Double{hi: math.Pi, lo: 1.2246467991473532e-16}
	DoubleTau    = Double{hi: 2 * math.Pi, lo: 2.4492935982947064e-16}
	DoubleE      = Double{hi: math.E, lo: 1.4456468917292502e-16}
	DoubleLog2   = Double{hi: math.Ln2, lo: 2.319046813846299e-17}
	DoublePhi    = Double{hi: math.Phi, lo: -5.432115203682505e-17}
)

// Double float by T.J. Dekker.
//
// See: T. J. Dekker, "A floating-point technique for extending the available precision", Numerische Mathematik 18 (3), 1971, 224-242. doi:10.1007/BF01397083
type Double struct{ hi, lo float64 }

func DoubleFromFloat(x float64) Double  { return Double{hi: x, lo: 0.} }
func DoubleFromSum(a, b float64) Double { return twoSum(a, b) }
func DoubleFromMul(a, b float64) Double { return twoProd(a, b) }
func DoubleFromSqr(x float64) Double    { return oneSqr(x) }

func (d Double) ToFloat64() float64  { return d.hi + d.lo }
func (d Double) Equal(x Double) bool { return (d.hi == x.hi && d.lo == x.lo) }
func (d Double) Add(x Double) Double { return add22(d, x) }
func (d Double) Sub(x Double) Double { return sub22(d, x) }
func (d Double) Mul(x Double) Double { return mul22(d, x) }
func (d Double) Div(x Double) Double { return div22(d, x) }
func (d Double) Abs() Double         { return absD(d) }
func (d Double) Neg() Double         { return negD(d) }
func (d Double) Inv() Double         { return inv2(d) }
func (d Double) Sqr() Double         { return Sqr2(d) }
func (d Double) GT(x Double) bool    { return (d.hi > x.hi || (d.hi == x.hi && d.lo > x.lo)) }
func (d Double) LT(x Double) bool    { return (d.hi < x.hi || (d.hi == x.hi && d.lo < x.lo)) }
func (d Double) GE(x Double) bool    { return (d.hi > x.hi || (d.hi == x.hi && d.lo >= x.lo)) }
func (d Double) LE(x Double) bool    { return (d.hi < x.hi || (d.hi == x.hi && d.lo <= x.lo)) }

func eq21(x Double, f float64) bool { return (x.hi == f && x.lo == 0.) }
func le21(x Double, f float64) bool { return (x.hi < f || (x.hi == f && x.lo <= 0.)) }

// x ** 2
func Sqr2(x Double) Double {
	S := oneSqr(x.hi)
	c := x.hi * x.lo
	S.lo += c + c
	return Double{
		hi: S.hi + S.lo,
		lo: S.lo - (x.hi - S.hi),
	}
}

// x ** 0.5
func Sqrt2(x Double) Double {
	s := math.Sqrt(x.hi)
	T := oneSqr(s)
	e := (x.hi - T.hi - T.lo + x.lo) * 0.5 / s
	return Double{
		hi: s + e,
		lo: e - (x.hi - s),
	}
}

var padeCoef = []float64{1, 272, 36720, 3255840, 211629600, 10666131840, 430200650880, 14135164243200,
	381649434566400, 8481098545920000, 154355993535744030, 2273242813890047700, 26521166162050560000,
	236650405753681870000, 1.5213240369879552e+21, 6.288139352883548e+21, 1.2576278705767096e+22}

// e ** x
func Exp(x Double) Double {
	if eq21(x, 0.) {
		return DoubleOne
	}
	if eq21(x, 1.) {
		return DoubleE
	}
	n := math.Floor(x.hi/DoubleLog2.hi + 0.5)
	x = sub22(x, mulDF(DoubleLog2, n))
	U := DoubleOne
	V := DoubleOne
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
	x = mulDFpow2(div22(U, V), int(n))
	return x
}

func Ln2(x Double) Double {
	if le21(x, 0) {
		return DoubleNegInf
	}
	if eq21(x, 1) {
		return Double{}
	}
	Z := DoubleFromFloat(math.Log(x.hi))
	return subDF(add22(mul22(x, Exp(negD((Z)))), Z), 1.)
}

func Sinh2(x Double) Double {
	exp := Exp(x)
	return mulDFpow2(sub22((exp), inv2(exp)), -1.)
}

func Cosh2(x Double) Double {
	exp := Exp(x)
	return mulDFpow2(add22((exp), inv2(exp)), -1.)
}

func Pow22(base Double, exp Double) Double {
	return Exp(mul22(Ln2(base), exp))
}

const splitter = 1<<27 + 1 // Veltkamp’s splitter

func twoSum(a, b float64) Double {
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
	s := twoSum(x.hi, y.hi)
	e := twoSum(x.lo, y.lo)
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
	s := twoSum(x.hi, -y.hi)
	e := twoSum(x.lo, -y.lo)
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
	s := twoSum(x.hi, f)
	s.lo += x.lo
	return Double{
		hi: s.hi + s.lo,
		lo: s.lo - (x.hi - s.hi),
	}
}

// x - f
func subDF(x Double, f float64) Double {
	s := twoSum(x.hi, -f)
	s.lo += x.lo
	return Double{
		hi: s.hi + s.lo,
		lo: s.lo - (x.hi - s.hi),
	}
}

// x * f
func mulDF(x Double, f float64) Double {
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
func divDF(x Double, f float64) Double {
	th := x.hi / f
	p := twoProd(th, f)
	d := twoSum(x.hi, -p.hi)
	tl := (d.hi + (d.lo + (x.lo - p.lo))) / f
	return Double{
		hi: th + tl,
		lo: tl - (x.hi - th),
	}
}

// |x|
func absD(x Double) Double {
	if x.hi < 0. {
		return negD(x)
	}
	return x
}

// -x
func negD(x Double) Double { return Double{hi: -x.hi, lo: -x.lo} }

// 1 / x
func inv2(x Double) Double {
	xh := x.hi
	s := 1. / xh
	x = mulDF(x, s)
	zl := (1. - x.hi - x.lo) / xh
	return Double{
		hi: s + zl,
		lo: zl - (x.hi - s),
	}
}

// x * 2 ** n
func mulDFpow2(x Double, n int) Double {
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
		return DoubleOne
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
		x = Sqr2(x)
	}
	for ; j != 0; j-- {
		x = mul22(x, x0)
	}
	if isPositive {
		return x
	}
	return inv2(x)
}
