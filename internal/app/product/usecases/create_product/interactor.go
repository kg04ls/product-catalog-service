package create_product

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"

	"product-catalog-service/internal/app/product/contracts"
	"product-catalog-service/internal/app/product/domain"
	"product-catalog-service/internal/pkg/clock"
	"product-catalog-service/internal/pkg/committer"
)

type Request struct {
	ID          string
	Name        string
	Description string
	Category    string
	BasePrice   *domain.Money
}

type Interactor struct {
	products contracts.ProductRepo
	outbox   contracts.OutboxRepo
	comm     committer.Committer
	clock    clock.Clock
}

func New(products contracts.ProductRepo, outbox contracts.OutboxRepo, comm committer.Committer, clk clock.Clock) *Interactor {
	return &Interactor{products: products, outbox: outbox, comm: comm, clock: clk}
}

func (it *Interactor) Execute(ctx context.Context, req Request) (string, error) {
	if req.BasePrice == nil {
		return "", errors.New("base price is required")
	}

	now := it.clock.Now()

	p, err := domain.NewProduct(req.ID, req.Name, req.Description, req.Category, req.BasePrice, now)
	if err != nil {
		return "", err
	}

	plan := committer.NewPlan()

	plan.Add(it.products.InsertMut(p))

	for _, e := range p.DomainEvents() {
		b, err := json.Marshal(e)
		if err != nil {
			return "", err
		}
		m := it.outbox.InsertMut(uuid.NewString(), e.EventType(), req.ID, b)
		if m == nil {
			return "", errors.New("outbox mutation is nil")
		}
		plan.Add(m)
	}

	if err := it.comm.Apply(ctx, plan); err != nil {
		return "", err
	}

	return req.ID, nil
}
