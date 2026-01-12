package domain

import (
	"math/big"
)

type Money struct {
	amount *big.Rat
}

func NewMoneyFromRat(r *big.Rat) (*Money, error) {
	if r == nil || r.Sign() < 0 {
		return nil, ErrInvalidMoney
	}
	return &Money{amount: new(big.Rat).Set(r)}, nil
}

func NewMoneyFromFraction(numerator, denominator int64) (*Money, error) {
	if denominator == 0 || numerator < 0 {
		return nil, ErrInvalidMoney
	}
	return NewMoneyFromRat(new(big.Rat).SetFrac64(numerator, denominator))
}

func (m *Money) Rat() *big.Rat {
	return new(big.Rat).Set(m.amount)
}

func (m *Money) Numerator() int64 {
	return m.amount.Num().Int64()
}

func (m *Money) Denominator() int64 {
	return m.amount.Denom().Int64()
}

func (m *Money) Sub(other *Money) *Money {
	return &Money{amount: new(big.Rat).Sub(m.amount, other.amount)}
}

func (m *Money) Mul(r *big.Rat) *Money {
	return &Money{amount: new(big.Rat).Mul(m.amount, r)}
}
