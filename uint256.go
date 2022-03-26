package mathx

import (
	"math/big"
)

// Uint256 represents a uint256 using 2 uint64.
type Uint256 struct {
	hi, lo Uint128
	_      struct{}
}

func NewUint256(hi, lo Uint128) Uint256  { return Uint256{hi: hi, lo: lo} }
func Uint256FromUint64(v uint64) Uint256 { return NewUint256(Uint128{}, NewUint128(0, v)) }

func (u Uint256) Parts() (Uint128, Uint128) { return u.hi, u.lo }
func (u Uint256) IsZero() bool              { return u.hi.IsZero() && u.lo.IsZero() }
func (u Uint256) Equals(x Uint256) bool     { return u.hi.Equals(x.hi) && u.lo.Equals(x.lo) }

func (u Uint256) Cmp(x Uint256) int {
	if h := u.hi.Cmp(x.hi); h != 0 {
		return h
	}
	return u.lo.Cmp(x.lo)
}

func (u Uint256) Inc() Uint256 { return u.Add(Uint256{lo: Uint128{lo: 1}}) }

func (u Uint256) Dec() Uint256 { return u.Sub(Uint256{lo: Uint128{lo: 1}}) }

func (u Uint256) Add(x Uint256) Uint256 {
	s, _ := u.AddCarry(x, 0)
	return s
}

func (u Uint256) AddCarry(x Uint256, carry uint64) (Uint256, uint64) {
	lo, c := u.lo.AddCarry(x.lo, carry)
	hi, c := u.hi.AddCarry(x.hi, c)
	return Uint256{hi: hi, lo: lo}, c
}

func (u Uint256) Sub(x Uint256) Uint256 {
	d, _ := u.SubBorrow(x, 0)
	return d
}

func (u Uint256) SubBorrow(x Uint256, borrow uint64) (Uint256, uint64) {
	lo, b := u.lo.SubBorrow(x.lo, borrow)
	hi, b := u.hi.SubBorrow(x.hi, b)
	return Uint256{hi: hi, lo: lo}, b
}

func (u Uint256) Mul(v Uint256) Uint256 {
	hi, lo := u.lo.MulFull(v.lo)
	hi = hi.Add(u.hi.Mul(v.lo))
	hi = hi.Add(u.lo.Mul(v.hi))
	return Uint256{lo: lo, hi: hi}
}

func (u Uint256) MulFull(x Uint256) (Uint256, Uint256) {
	var lo, hi Uint256
	lo.hi, lo.lo = u.lo.MulFull(x.lo)
	hi.hi, hi.lo = u.hi.MulFull(x.hi)
	t0, t1 := u.lo.MulFull(x.hi)
	t2, t3 := u.hi.MulFull(x.lo)

	var c0, c1 uint64
	lo.hi, c0 = lo.hi.AddCarry(t1, 0)
	lo.hi, c1 = lo.hi.AddCarry(t3, 0)
	hi.lo, c0 = hi.lo.AddCarry(t0, c0)
	hi.lo, c1 = hi.lo.AddCarry(t2, c1)
	hi.hi = hi.hi.Add(Uint128{lo: c0 + c1})
	return hi, lo
}

func (u Uint256) And(x Uint256) Uint256 { return Uint256{hi: u.hi.And(x.hi), lo: u.lo.And(x.lo)} }
func (u Uint256) Xor(x Uint256) Uint256 { return Uint256{hi: u.hi.Xor(x.hi), lo: u.lo.Xor(x.lo)} }
func (u Uint256) Or(x Uint256) Uint256  { return Uint256{hi: u.hi.Or(x.hi), lo: u.lo.Or(x.lo)} }
func (u Uint256) Not() Uint256          { return Uint256{hi: u.hi.Not(), lo: u.lo.Not()} }

func (u Uint256) Lsh(n uint) Uint256 {
	if n > 128 {
		return Uint256{hi: u.lo.Lsh(n - 128), lo: Uint128{}}
	}
	return Uint256{
		hi: u.hi.Lsh(n).Or(u.lo.Rsh(128 - n)),
		lo: u.lo.Lsh(n),
	}
}

func (u Uint256) Rsh(n uint) Uint256 {
	if n > 128 {
		return Uint256{hi: Uint128{}, lo: u.hi.Rsh(n - 128)}
	}
	return Uint256{
		hi: u.hi.Rsh(n),
		lo: u.lo.Rsh(n).Or(u.hi.Lsh(128 - n)),
	}
}

func (u Uint256) Big() *big.Int {
	i := u.hi.Big()
	i = i.Lsh(i, 128)
	i = i.Xor(i, u.lo.Big())
	return i
}

func (u Uint256) String() string {
	if u.IsZero() {
		return "0"
	}
	return u.Big().String()
}
