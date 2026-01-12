package spannerx

import (
	"cloud.google.com/go/spanner"

	"product-catalog-service/internal/app/product/contracts"
)

type Mutation struct {
	M *spanner.Mutation
}

func (Mutation) IsMutation() {}

func Wrap(m *spanner.Mutation) contracts.Mutation {
	if m == nil {
		return nil
	}
	return Mutation{M: m}
}
