package services

import (
	"math/big"
	"time"

	"product-catalog-service/internal/app/product/domain"
)

type PricingCalculator struct{}

func NewPricingCalculator() *PricingCalculator {
	return &PricingCalculator{}
}

func (pc *PricingCalculator) EffectivePrice(p *domain.Product, now time.Time) *domain.Money {
	base := p.BasePrice()
	d := p.Discount()
	if d == nil || !d.IsValidAt(now) {
		return base
	}

	percent := d.Percent()
	rate := new(big.Rat).Quo(percent, big.NewRat(100, 1))
	discountAmount := base.Mul(rate)
	return base.Sub(discountAmount)
}
