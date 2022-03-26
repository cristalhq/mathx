package mathx

import (
	"fmt"
	"math/big"
	"math/bits"
)

// Uint128 represents a uint128 using 2 uint64.
type Uint128 struct {
	hi, lo uint64
	_      struct{}
}

func NewUint128(hi, lo uint64) Uint128 {
	return Uint128{hi: hi, lo: lo}
}

func Uint128FromUint64(v uint64) Uint128 {
	return NewUint128(0, v)
}

func Uint128FromString(s string) (Uint128, error) {
	var u Uint128
	_, err := fmt.Sscan(s, &u)
	return u, err
}

func (u Uint128) Parts() (uint64, uint64) { return u.hi, u.lo }
func (u Uint128) IsZero() bool            { return u.hi|u.lo == 0 }
func (u Uint128) Equals(x Uint128) bool   { return u == x }

func (u Uint128) Cmp(x Uint128) int {
	switch {
	case u == x:
		return 0
	case u.hi < x.hi || (u.hi == x.hi && u.lo < x.lo):
		return -1
	default:
		return 1
	}
}

func (u Uint128) Inc() Uint128 {
	lo, carry := bits.Add64(u.lo, 1, 0)
	return Uint128{hi: u.hi + carry, lo: lo}
}

func (u Uint128) Dec() Uint128 {
	lo, borrow := bits.Sub64(u.lo, 1, 0)
	return Uint128{hi: u.hi - borrow, lo: lo}
}

func (u Uint128) Add(x Uint128) Uint128 {
	lo, carry := bits.Add64(u.lo, x.lo, 0)
	hi, _ := bits.Add64(u.hi, x.hi, carry)
	return Uint128{hi: hi, lo: lo}
}

func (u Uint128) AddCarry(x Uint128, carry uint64) (Uint128, uint64) {
	lo, c := bits.Add64(u.lo, x.lo, carry)
	hi, c := bits.Add64(u.hi, x.hi, c)
	return Uint128{hi: hi, lo: lo}, c
}

func (u Uint128) Sub(x Uint128) Uint128 {
	lo, borrow := bits.Sub64(u.lo, x.lo, 0)
	hi, _ := bits.Sub64(u.hi, x.hi, borrow)
	return Uint128{hi: hi, lo: lo}
}

func (u Uint128) SubBorrow(x Uint128, borrow uint64) (Uint128, uint64) {
	lo, b := bits.Sub64(u.lo, x.lo, borrow)
	hi, b := bits.Sub64(u.hi, x.hi, b)
	return Uint128{hi: hi, lo: lo}, b
}

func (u Uint128) Mul(x Uint128) Uint128 {
	hi, lo := bits.Mul64(u.lo, x.lo)
	hi += u.hi*x.lo + u.lo*x.hi
	return Uint128{hi: hi, lo: lo}
}

// multiply 128-bit unsigned integers and return high and lower product
func (a Uint128) MulFull(b Uint128) (Uint128, Uint128) {
	var lo, m1, m2, hi Uint128
	lo.hi, lo.lo = bits.Mul64(a.lo, b.lo)
	m1.hi, m1.lo = bits.Mul64(a.hi, b.lo)
	m2.hi, m2.lo = bits.Mul64(a.lo, b.hi)
	hi.hi, hi.lo = bits.Mul64(a.hi, b.hi)
	var carry uint64
	lo.hi, carry = bits.Add64(lo.hi, m1.lo, 0)
	hi.lo, carry = bits.Add64(hi.lo, m1.hi, carry)
	hi.hi, _ = bits.Add64(hi.hi, 0, carry)
	lo.hi, carry = bits.Add64(lo.hi, m2.lo, 0)
	hi.lo, carry = bits.Add64(hi.lo, m2.hi, carry)
	hi.hi, _ = bits.Add64(hi.hi, 0, carry)
	return hi, lo
}

func (u Uint128) And(x Uint128) Uint128 { return Uint128{hi: u.hi & x.hi, lo: u.lo & x.lo} }
func (u Uint128) Xor(x Uint128) Uint128 { return Uint128{hi: u.hi ^ x.hi, lo: u.lo ^ x.lo} }
func (u Uint128) Or(x Uint128) Uint128  { return Uint128{hi: u.hi | x.hi, lo: u.lo | x.lo} }
func (u Uint128) Not() Uint128          { return Uint128{hi: ^u.hi, lo: ^u.lo} }

func (u Uint128) Lsh(n uint) Uint128 {
	if n > 64 {
		return Uint128{hi: u.lo << (n - 64), lo: 0}
	}
	return Uint128{
		hi: u.hi<<n | u.lo>>(64-n),
		lo: u.lo << n,
	}
}

func (u Uint128) Rsh(n uint) Uint128 {
	if n > 64 {
		return Uint128{hi: 0, lo: u.hi >> (n - 64)}
	}
	return Uint128{
		hi: u.hi >> n,
		lo: u.lo>>n | u.hi<<(64-n),
	}
}

func (u Uint128) Big() *big.Int {
	i := new(big.Int).SetUint64(u.hi)
	i = i.Lsh(i, 64)
	i = i.Xor(i, new(big.Int).SetUint64(u.lo))
	return i
}

func (u Uint128) String() string {
	if u.IsZero() {
		return "0"
	}
	return u.Big().String()
}
