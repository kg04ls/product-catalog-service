package apply_discount

import (
	"context"
	"math/big"
	"testing"
	"time"

	"product-catalog-service/internal/app/product/contracts"
	"product-catalog-service/internal/app/product/domain"
	"product-catalog-service/internal/app/product/usecases/activate_product"

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

func TestApplyDiscount_Smoke(t *testing.T) {
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

	actpr := activate_product.New(pr, fakeOutboxRepo{}, sc, fakeClock{t: now})
	err = actpr.Execute(context.Background(), activate_product.Request{
		ProductID: "p1",
	})

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	it := New(pr, fakeOutboxRepo{}, sc, fakeClock{t: now})

	err = it.Execute(context.Background(), Request{
		ProductID: "p1",
		Percent:   big.NewRat(10, 1),
		Start:     now.Truncate(time.Hour),
		End:       now.Add(2 * time.Hour),
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if pr.updateN != 2 {
		t.Fatalf("expected UpdateMut called 1 time, got %d", pr.updateN)
	}
	if sc.applied != 2 {
		t.Fatalf("expected Apply called 1 time, got %d", sc.applied)
	}
	if sc.last == nil || len(sc.last.Mutations()) < 1 {
		t.Fatalf("expected plan with at least 1 mutation")
	}
}

func TestApplyDiscount_InactiveProduct_ReturnsErr(t *testing.T) {
	now := time.Date(2026, 1, 12, 12, 0, 0, 0, time.UTC)
	price, err := domain.NewMoneyFromRat(big.NewRat(1000, 100))
	if err != nil {
		t.Fatalf("setup money: %v", err)
	}

	p, _ := domain.NewProduct("p1", "Name", "Desc", "Cat", price, now)

	pr := &fakeProductRepo{p: p}
	sc := &spyCommitter{}
	it := New(pr, fakeOutboxRepo{}, sc, fakeClock{t: now})

	err = it.Execute(context.Background(), Request{
		ProductID: "p1",
		Percent:   big.NewRat(10, 1),
		Start:     now.Add(time.Hour),
		End:       now.Add(2 * time.Hour),
	})

	if err == nil {
		t.Fatalf("expected error")
	}
	if sc.applied != 0 {
		t.Fatalf("expected Apply not called")
	}
}
