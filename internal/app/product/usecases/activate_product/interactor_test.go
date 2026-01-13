package activate_product

import (
	"context"
	"math/big"
	"testing"
	"time"

	"product-catalog-service/internal/app/product/contracts"
	"product-catalog-service/internal/app/product/domain"
	"product-catalog-service/internal/pkg/committer"
)

type fakeClock struct{ t time.Time }

func (f fakeClock) Now() time.Time { return f.t }

type fakeMut struct{}

func (fakeMut) IsMutation() {}

type fakeProductRepo struct {
	p       *domain.Product
	updateN int
}

func (r *fakeProductRepo) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	return r.p, nil
}
func (r *fakeProductRepo) InsertMut(p *domain.Product) contracts.Mutation { return fakeMut{} }
func (r *fakeProductRepo) UpdateMut(p *domain.Product) contracts.Mutation {
	r.updateN++
	return fakeMut{}
}

type fakeOutboxRepo struct{}

func (fakeOutboxRepo) InsertMut(eventID, eventType, aggregateID string, payload []byte) contracts.Mutation {
	return fakeMut{}
}

type spyCommitter struct {
	applied int
	last    *committer.Plan
}

func (s *spyCommitter) Apply(ctx context.Context, plan *committer.Plan) error {
	s.applied++
	s.last = plan
	return nil
}

func TestActivateProduct_Smoke(t *testing.T) {
	now := time.Date(2026, 1, 12, 12, 0, 0, 0, time.UTC)
	price, err := domain.NewMoneyFromRat(big.NewRat(1000, 100))
	if err != nil {
		t.Fatalf("setup money: %v", err)
	}
	p, err := domain.NewProduct("p1", "Name", "Desc", "Cat", price, now)
	if err != nil {
		t.Fatalf("setup product: %v", err)
	}

	pr := &fakeProductRepo{p: p}
	sc := &spyCommitter{}

	it := New(pr, fakeOutboxRepo{}, sc, fakeClock{t: now})

	if err := it.Execute(context.Background(), Request{ProductID: "p1"}); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if pr.updateN != 1 {
		t.Fatalf("expected UpdateMut called 1 time, got %d", pr.updateN)
	}
	if sc.applied != 1 {
		t.Fatalf("expected Apply called 1 time, got %d", sc.applied)
	}
	if sc.last == nil || len(sc.last.Mutations()) < 1 {
		t.Fatalf("expected plan with at least 1 mutation")
	}
}
