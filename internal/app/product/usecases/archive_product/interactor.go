package archive_product

import (
	"context"
	"encoding/json"
	"errors"

	"product-catalog-service/internal/app/product/contracts"
	"product-catalog-service/internal/pkg/clock"
	"product-catalog-service/internal/pkg/committer"

	"github.com/google/uuid"
)

type Request struct {
	ProductID string
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

func (it *Interactor) Execute(ctx context.Context, req Request) error {
	p, err := it.products.GetByID(ctx, req.ProductID)
	if err != nil {
		return err
	}

	err = p.Archive(it.clock.Now())
	if err != nil {
		return err
	}

	plan := committer.NewPlan()

	plan.Add(it.products.UpdateMut(p))

	for _, e := range p.DomainEvents() {
		b, err := json.Marshal(e)
		if err != nil {
			return err
		}

		m := it.outbox.InsertMut(uuid.NewString(), e.EventType(), req.ProductID, b)
		if m == nil {
			return errors.New("outbox mutation is nil")
		}
		plan.Add(m)
	}
	return it.comm.Apply(ctx, plan)
}
