package contracts

import (
	"context"

	"product-catalog-service/internal/app/product/domain"
)

type ProductRepo interface {
	GetByID(ctx context.Context, id string) (*domain.Product, error)
	InsertMut(p *domain.Product) Mutation
	UpdateMut(p *domain.Product) Mutation
}
