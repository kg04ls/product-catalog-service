package domain

import (
	"math/big"
	"time"
)

type Discount struct {
	percent *big.Rat
	start   time.Time
	end     time.Time
}

func NewDiscount(percent *big.Rat, start, end time.Time) (*Discount, error) {
	if percent == nil || percent.Sign() <= 0 {
		return nil, ErrInvalidDiscountPercent
	}
	if percent.Cmp(big.NewRat(100, 1)) > 0 {
		return nil, ErrInvalidDiscountPercent
	}
	if start.IsZero() || end.IsZero() || !end.After(start) {
		return nil, ErrInvalidDiscountPeriod
	}

	return &Discount{
		percent: new(big.Rat).Set(percent),
		start:   start.UTC(),
		end:     end.UTC(),
	}, nil
}

func (d *Discount) Percent() *big.Rat {
	return new(big.Rat).Set(d.percent)
}

func (d *Discount) Start() time.Time {
	return d.start
}

func (d *Discount) End() time.Time {
	return d.end
}

func (d *Discount) IsValidAt(now time.Time) bool {
	t := now.UTC()
	return (t.Equal(d.start) || t.After(d.start)) && t.Before(d.end)
}

func (d *Discount) Overlaps(other *Discount) bool {
	if d == nil || other == nil {
		return false
	}
	return d.start.Before(other.end) && other.start.Before(d.end)
}
