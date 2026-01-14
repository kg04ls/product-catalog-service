package repo

import (
	"context"
	"math/big"
	"time"

	"cloud.google.com/go/spanner"
	"google.golang.org/api/iterator"

	"product-catalog-service/internal/app/product/contracts"
	"product-catalog-service/internal/pkg/clock"
)

type SpannerReadModel struct {
	client *spanner.Client
	clock  clock.Clock
}

func NewSpannerReadModel(client *spanner.Client, clk clock.Clock) *SpannerReadModel {
	return &SpannerReadModel{client: client, clock: clk}
}

func (r *SpannerReadModel) GetProduct(ctx context.Context, id string) (contracts.ProductDTO, error) {
	st := spanner.NewStatement(`
		SELECT product_id, name, description, category,
		       base_price_numerator, base_price_denominator,
		       discount_percent, discount_start_date, discount_end_date,
		       status, created_at, updated_at, archived_at
		FROM products
		WHERE product_id = @id
	`)
	st.Params["id"] = id

	iter := r.client.Single().Query(ctx, st)
	defer iter.Stop()

	row, err := iter.Next()
	if err != nil {
		if err == iterator.Done {
			return contracts.ProductDTO{}, ErrProductNotFound
		}
		return contracts.ProductDTO{}, err
	}

	return r.mapRowToDTO(row)
}

func (r *SpannerReadModel) ListProducts(ctx context.Context, f contracts.ListProductsFilter) (contracts.ListProductsResult, error) {
	query := `
		SELECT product_id, name, description, category,
		       base_price_numerator, base_price_denominator,
		       discount_percent, discount_start_date, discount_end_date,
		       status, created_at, updated_at, archived_at
		FROM products
		WHERE archived_at IS NULL
	`
	params := make(map[string]interface{})

	if f.Category != "" {
		query += " AND category = @category"
		params["category"] = f.Category
	}

	if f.OnlyActive {
		query += " AND status = 'active'"
	}

	query += " ORDER BY created_at DESC"

	st := spanner.NewStatement(query)
	st.Params = params

	pageSize := int(f.Limit)
	if pageSize <= 0 {
		pageSize = 20
	}

	iter := r.client.Single().Query(ctx, st)
	defer iter.Stop()

	var items []contracts.ProductDTO
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return contracts.ListProductsResult{}, err
		}

		dto, err := r.mapRowToDTO(row)
		if err != nil {
			return contracts.ListProductsResult{}, err
		}
		items = append(items, dto)

		if len(items) >= pageSize {
			// Basic pagination: just stop here for now as Task doesn't require full sophisticated pagination
			break
		}
	}

	return contracts.ListProductsResult{
		Items: items,
	}, nil
}

func (r *SpannerReadModel) mapRowToDTO(row *spanner.Row) (contracts.ProductDTO, error) {
	var (
		id, name, category, status string
		baseNum, baseDen           int64
		discPercent                spanner.NullNumeric
		discStart, discEnd         spanner.NullTime
		createdAt, updatedAt       time.Time
		archivedAt                 spanner.NullTime
		description                spanner.NullString
	)

	if err := row.Columns(
		&id, &name, &description, &category,
		&baseNum, &baseDen,
		&discPercent, &discStart, &discEnd,
		&status, &createdAt, &updatedAt, &archivedAt,
	); err != nil {
		return contracts.ProductDTO{}, err
	}

	dto := contracts.ProductDTO{
		ID:           id,
		Name:         name,
		Description:  description.StringVal,
		Category:     category,
		Status:       status,
		BasePriceNum: baseNum,
		BasePriceDen: baseDen,
		CreatedAt:    createdAt.Format(time.RFC3339),
		UpdatedAt:    updatedAt.Format(time.RFC3339),
		EffectiveNum: baseNum,
		EffectiveDen: baseDen,
	}

	if archivedAt.Valid {
		dto.ArchivedAt = archivedAt.Time.Format(time.RFC3339)
	}

	if discPercent.Valid && discStart.Valid && discEnd.Valid {
		now := r.clock.Now().UTC()
		if now.After(discStart.Time) && now.Before(discEnd.Time) {
			pct := new(big.Rat).Set(&discPercent.Numeric)
			base := big.NewRat(baseNum, baseDen)

			one := big.NewRat(1, 1)
			hundred := big.NewRat(100, 1)
			pctRatio := new(big.Rat).Quo(pct, hundred)
			multiplier := new(big.Rat).Sub(one, pctRatio)
			effective := new(big.Rat).Mul(base, multiplier)

			dto.EffectiveNum = effective.Num().Int64()
			dto.EffectiveDen = effective.Denom().Int64()

			dto.DiscountPct = pct.FloatString(2)
			dto.DiscountStart = discStart.Time.Format(time.RFC3339)
			dto.DiscountEnd = discEnd.Time.Format(time.RFC3339)
		}
	}

	return dto, nil
}
