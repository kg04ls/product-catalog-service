package spannerx

import (
	"context"
	"fmt"

	"cloud.google.com/go/spanner"

	"product-catalog-service/internal/pkg/committer"
)

type Committer struct {
	client *spanner.Client
}

func NewCommitter(client *spanner.Client) *Committer {
	return &Committer{client: client}
}

func (c *Committer) Apply(ctx context.Context, plan *committer.Plan) error {
	if plan == nil || len(plan.Mutations()) == 0 {
		return nil
	}

	muts := make([]*spanner.Mutation, 0, len(plan.Mutations()))
	for _, m := range plan.Mutations() {
		sm, ok := m.(Mutation)
		if !ok {
			return fmt.Errorf("unsupported mutation type: %T", m)
		}
		if sm.M != nil {
			muts = append(muts, sm.M)
		}
	}

	if len(muts) == 0 {
		return nil
	}

	_, err := c.client.ReadWriteTransaction(ctx, func(ctx context.Context, tx *spanner.ReadWriteTransaction) error {
		return tx.BufferWrite(muts)
	})
	return err
}

var _ committer.Committer = (*Committer)(nil)
