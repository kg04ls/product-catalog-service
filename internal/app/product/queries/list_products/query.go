package list_products

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

func (q *Query) Execute(ctx context.Context, f contracts.ListProductsFilter) (contracts.ListProductsResult, error) {
	return q.readModel.ListProducts(ctx, f)
}
