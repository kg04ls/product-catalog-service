package domain_test

import (
	"math/big"
	"testing"
	"time"

	"product-catalog-service/internal/app/product/domain"
)

func TestNewProduct_CreatesInactiveAndEmitsCreatedEvent(t *testing.T) {
	now := time.Date(2026, 1, 12, 10, 0, 0, 0, time.UTC)
	price, err := domain.NewMoneyFromFraction(100, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p, err := domain.NewProduct("p1", "name", "desc", "cat", price, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.Status() != domain.ProductStatusInactive {
		t.Fatalf("expected inactive, got %s", p.Status())
	}

	ev := p.DomainEvents()
	if len(ev) != 1 {
		t.Fatalf("expected 1 event, got %d", len(ev))
	}
	if ev[0].EventType() != "product.created" {
		t.Fatalf("expected product.created, got %s", ev[0].EventType())
	}
}

func TestDiscount_IsValidAt_StartInclusive_EndExclusive(t *testing.T) {
	start := time.Date(2026, 1, 12, 10, 0, 0, 0, time.UTC)
	end := start.Add(2 * time.Hour)

	d, err := domain.NewDiscount(big.NewRat(10, 1), start, end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !d.IsValidAt(start) {
		t.Fatalf("expected valid at start")
	}
	if d.IsValidAt(end) {
		t.Fatalf("expected invalid at end")
	}
}

func TestApplyDiscount_InactiveProduct_ReturnsErrProductNotActive(t *testing.T) {
	now := time.Date(2026, 1, 12, 10, 0, 0, 0, time.UTC)
	price, err := domain.NewMoneyFromFraction(100, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p, err := domain.NewProduct("p1", "name", "desc", "cat", price, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	d, err := domain.NewDiscount(big.NewRat(10, 1), now.Add(-time.Minute), now.Add(time.Hour))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := p.ApplyDiscount(d, now); err != domain.ErrProductNotActive {
		t.Fatalf("expected ErrProductNotActive, got %v", err)
	}
}

func TestMoney_MulAndSub(t *testing.T) {
	base, err := domain.NewMoneyFromFraction(200, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	disc := base.Mul(big.NewRat(25, 100))
	if disc.Rat().Cmp(big.NewRat(50, 1)) != 0 {
		t.Fatalf("expected 50, got %s", disc.Rat().String())
	}

	final := base.Sub(disc)
	if final.Rat().Cmp(big.NewRat(150, 1)) != 0 {
		t.Fatalf("expected 150, got %s", final.Rat().String())
	}
}
