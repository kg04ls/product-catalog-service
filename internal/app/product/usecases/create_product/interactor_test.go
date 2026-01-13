package create_product

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

type fakeProductRepo struct{ insertN int }

func (r *fakeProductRepo) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	return nil, nil
}
func (r *fakeProductRepo) InsertMut(p *domain.Product) contracts.Mutation {
	r.insertN++
	return fakeMut{}
}
func (r *fakeProductRepo) UpdateMut(p *domain.Product) contracts.Mutation { return fakeMut{} }

type fakeOutboxRepo struct{ n int }

func (o *fakeOutboxRepo) InsertMut(eventID, eventType, aggregateID string, payload []byte) contracts.Mutation {
	o.n++
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

func TestCreateProduct_Smoke(t *testing.T) {
	now := time.Date(2026, 1, 12, 12, 0, 0, 0, time.UTC)
	price, err := domain.NewMoneyFromRat(big.NewRat(1999, 100))
	if err != nil {
		t.Fatalf("setup money: %v", err)
	}

	pr := &fakeProductRepo{}
	or := &fakeOutboxRepo{}
	sc := &spyCommitter{}

	it := New(pr, or, sc, fakeClock{t: now})

	id, err := it.Execute(context.Background(), Request{
		ID:          "p1",
		Name:        "Name",
		Description: "Desc",
		Category:    "Cat",
		BasePrice:   price,
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if id != "p1" {
		t.Fatalf("expected id p1, got %s", id)
	}
	if pr.insertN != 1 {
		t.Fatalf("expected InsertMut called 1 time, got %d", pr.insertN)
	}
	if sc.applied != 1 {
		t.Fatalf("expected Apply called 1 time, got %d", sc.applied)
	}
	if sc.last == nil || len(sc.last.Mutations()) < 1 {
		t.Fatalf("expected plan with at least 1 mutation")
	}
}
