package services_test

import (
	"math/big"
	"testing"
	"time"

	"product-catalog-service/internal/app/product/domain"
	"product-catalog-service/internal/app/product/domain/services"
)

func TestPricingCalculator_NoDiscount_ReturnsBasePrice(t *testing.T) {
	price, err := domain.NewMoneyFromFraction(100, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	now := time.Date(2026, 1, 12, 10, 0, 0, 0, time.UTC)

	p, err := domain.NewProduct("p1", "n", "d", "c", price, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pc := services.NewPricingCalculator()
	eff := pc.EffectivePrice(p, now)

	if eff.Rat().Cmp(price.Rat()) != 0 {
		t.Fatalf("expected base price, got %s", eff.Rat().String())
	}
}

func TestPricingCalculator_ExpiredDiscount_Ignored(t *testing.T) {
	price, err := domain.NewMoneyFromFraction(200, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	now := time.Date(2026, 1, 12, 10, 0, 0, 0, time.UTC)

	p, err := domain.NewProduct("p1", "n", "d", "c", price, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := p.Activate(now); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	d, err := domain.NewDiscount(
		big.NewRat(50, 1),
		now.Add(-2*time.Hour),
		now.Add(-time.Hour),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := p.ApplyDiscount(d, now.Add(-90*time.Minute)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pc := services.NewPricingCalculator()
	eff := pc.EffectivePrice(p, now)

	if eff.Rat().Cmp(price.Rat()) != 0 {
		t.Fatalf("expected base price, got %s", eff.Rat().String())
	}
}
