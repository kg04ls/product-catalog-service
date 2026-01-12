package contracts

import (
	"context"
	"product-catalog-service/internal/app/product/domain"
)

type ProductRepo interface {
	InsertMut(p *domain.Product) any
	UpdateMut(p *domain.Product) any
	GetByID(ctx context.Context, id string) (*domain.Product, error)
}
