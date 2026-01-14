package get_product

import (
	"context"

	"product-catalog-service/internal/app/product/contracts"
)

type Query struct {
	readModel contracts.ProductReadModel
}

func New(readModel contracts.ProductReadModel) *Query {
	return &Query{readModel: readModel}
}

func (q *Query) Execute(ctx context.Context, id string) (contracts.ProductDTO, error) {
	return q.readModel.GetProduct(ctx, id)
}
