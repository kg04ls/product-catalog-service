package repo

import (
	"context"
	"errors"
	"math/big"

	"cloud.google.com/go/spanner"
	"google.golang.org/api/iterator"

	"product-catalog-service/internal/app/product/contracts"
	"product-catalog-service/internal/app/product/domain"
	"product-catalog-service/internal/infra/spannerx"
	"product-catalog-service/internal/models/m_product"
	"product-catalog-service/internal/pkg/clock"
)

var ErrProductNotFound = errors.New("product not found")

type ProductRepo struct {
	client *spanner.Client
	model  m_product.Model
	clock  clock.Clock
}

func NewProductRepo(client *spanner.Client, clk clock.Clock) *ProductRepo {
	return &ProductRepo{
		client: client,
		model:  m_product.Model{},
		clock:  clk,
	}
}

func (r *ProductRepo) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	st := spanner.NewStatement(`
		SELECT product_id, name, description, category,
		       base_price_numerator, base_price_denominator,
		       discount_percent, discount_start_date, discount_end_date,
		       status
		FROM products
		WHERE product_id = @id
	`)
	st.Params["id"] = id

	iter := r.client.Single().Query(ctx, st)
	defer iter.Stop()

	row, err := iter.Next()
	if err != nil {
		if err == iterator.Done {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	var (
		productID string
		name      string
		desc      spanner.NullString
		category  string

		baseNum int64
		baseDen int64

		discPercent spanner.NullNumeric
		discStart   spanner.NullTime
		discEnd     spanner.NullTime

		statusStr string
	)

	if err := row.Columns(
		&productID, &name, &desc, &category,
		&baseNum, &baseDen,
		&discPercent, &discStart, &discEnd,
		&statusStr,
	); err != nil {
		return nil, err
	}

	base, err := domain.NewMoneyFromFraction(baseNum, baseDen)
	if err != nil {
		return nil, err
	}

	var discount *domain.Discount
	if discPercent.Valid && discStart.Valid && discEnd.Valid {
		dp := new(big.Rat).Set(&discPercent.Numeric)
		d, derr := domain.NewDiscount(dp, discStart.Time, discEnd.Time)

		if derr != nil {
			return nil, derr
		}
		discount = d
	}

	status := parseStatus(statusStr)

	p := domain.HydrateProduct(
		productID,
		name,
		nullString(desc),
		category,
		base,
		discount,
		status,
	)

	return p, nil
}

func (r *ProductRepo) InsertMut(p *domain.Product) contracts.Mutation {
	now := r.clock.Now()

	row := map[string]interface{}{
		m_product.ProductID:            p.ID(),
		m_product.Name:                 p.Name(),
		m_product.Description:          p.Description(),
		m_product.Category:             p.Category(),
		m_product.BasePriceNumerator:   p.BasePrice().Numerator(),
		m_product.BasePriceDenominator: p.BasePrice().Denominator(),
		m_product.Status:               string(p.Status()),
		m_product.CreatedAt:            now,
		m_product.UpdatedAt:            now,
	}

	if d := p.Discount(); d != nil {
		row[m_product.DiscountPercent] = spanner.NullNumeric{Numeric: *d.Percent(), Valid: true}
		row[m_product.DiscountStartDate] = d.Start()
		row[m_product.DiscountEndDate] = d.End()
	}

	return spannerx.Wrap(r.model.InsertMut(row))
}

func (r *ProductRepo) UpdateMut(p *domain.Product) contracts.Mutation {
	ch := p.Changes()

	updates := map[string]interface{}{
		m_product.ProductID: p.ID(),
	}

	if ch.Dirty(domain.FieldName) {
		updates[m_product.Name] = p.Name()
	}
	if ch.Dirty(domain.FieldDescription) {
		updates[m_product.Description] = p.Description()
	}
	if ch.Dirty(domain.FieldCategory) {
		updates[m_product.Category] = p.Category()
	}
	if ch.Dirty(domain.FieldStatus) {
		updates[m_product.Status] = string(p.Status())
	}
	if ch.Dirty(domain.FieldDiscount) {
		if d := p.Discount(); d != nil {
			updates[m_product.DiscountPercent] = spanner.NullNumeric{Numeric: *d.Percent(), Valid: true}
			updates[m_product.DiscountStartDate] = d.Start()
			updates[m_product.DiscountEndDate] = d.End()
		} else {
			updates[m_product.DiscountPercent] = spanner.NullNumeric{Valid: false}
			updates[m_product.DiscountStartDate] = spanner.NullTime{Valid: false}
			updates[m_product.DiscountEndDate] = spanner.NullTime{Valid: false}
		}
	}

	if len(updates) == 1 {
		return nil
	}
	updates[m_product.UpdatedAt] = r.clock.Now()

	return spannerx.Wrap(r.model.UpdateMut(updates))
}

func nullString(s spanner.NullString) string {
	if s.Valid {
		return s.StringVal
	}
	return ""
}

func parseStatus(s string) domain.ProductStatus {
	switch s {
	case string(domain.ProductStatusActive):
		return domain.ProductStatusActive
	case string(domain.ProductStatusInactive):
		return domain.ProductStatusInactive
	default:
		return domain.ProductStatusInactive
	}
}
