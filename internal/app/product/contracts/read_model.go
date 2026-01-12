package contracts

import "context"

type ProductDTO struct {
	ID            string
	Name          string
	Description   string
	Category      string
	Status        string
	BasePriceNum  int64
	BasePriceDen  int64
	DiscountPct   string
	DiscountStart string
	DiscountEnd   string
	EffectiveNum  int64
	EffectiveDen  int64
	CreatedAt     string
	UpdatedAt     string
	ArchivedAt    string
}

type ListProductsFilter struct {
	Category   string
	OnlyActive bool
	Limit      int32
	PageToken  string
}

type ListProductsResult struct {
	Items         []ProductDTO
	NextPageToken string
}

type ProductReadModel interface {
	GetProduct(ctx context.Context, id string) (ProductDTO, error)
	ListProducts(ctx context.Context, f ListProductsFilter) (ListProductsResult, error)
}
