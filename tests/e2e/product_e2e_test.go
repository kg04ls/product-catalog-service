package e2e

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/iterator"

	"product-catalog-service/internal/app/product/domain"
	"product-catalog-service/internal/app/product/repo"
	"product-catalog-service/internal/app/product/usecases/activate_product"
	"product-catalog-service/internal/app/product/usecases/apply_discount"
	"product-catalog-service/internal/app/product/usecases/create_product"
	"product-catalog-service/internal/infra/spannerx"
	"product-catalog-service/internal/pkg/committer"
)

const (
	projectID  = "test-project"
	instanceID = "test-instance"
	databaseID = "test-db"
)

func getSpannerClient(ctx context.Context, t *testing.T) *spanner.Client {
	if os.Getenv("SPANNER_EMULATOR_HOST") == "" {
		os.Setenv("SPANNER_EMULATOR_HOST", "localhost:9010")
	}

	dbPath := fmt.Sprintf("projects/%s/instances/%s/databases/%s", projectID, instanceID, databaseID)
	client, err := spanner.NewClient(ctx, dbPath)
	require.NoError(t, err)
	return client
}

type testClock struct {
	now time.Time
}

func (c *testClock) Now() time.Time {
	return c.now
}

type spannerCommitter struct {
	client *spanner.Client
}

func (s *spannerCommitter) Apply(ctx context.Context, plan *committer.Plan) error {
	_, err := s.client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		var spannerMutations []*spanner.Mutation
		for _, m := range plan.Mutations() {
			if sm, ok := m.(spannerx.Mutation); ok {
				spannerMutations = append(spannerMutations, sm.M)
			}
		}
		return txn.BufferWrite(spannerMutations)
	})
	return err
}

func TestProductCreationFlow(t *testing.T) {
	ctx := context.Background()
	client := getSpannerClient(ctx, t)
	defer client.Close()

	// Dependencies
	clk := &testClock{now: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)}
	comm := &spannerCommitter{client: client}
	productRepo := repo.NewProductRepo(client, clk)
	outboxRepo := repo.NewOutboxRepo()

	// Usecase
	uc := create_product.New(productRepo, outboxRepo, comm, clk)

	// Data
	productID := uuid.NewString()
	name := "Test Product"
	desc := "A product for testing"
	category := "electronics"
	basePrice, err := domain.NewMoneyFromFraction(100, 1) // 100.00
	require.NoError(t, err)

	req := create_product.Request{
		ID:          productID,
		Name:        name,
		Description: desc,
		Category:    category,
		BasePrice:   basePrice,
	}

	// Exec
	id, err := uc.Execute(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, productID, id)

	// Verify Product in DB
	storedProduct, err := productRepo.GetByID(ctx, productID)
	require.NoError(t, err)
	assert.Equal(t, name, storedProduct.Name())
	assert.Equal(t, domain.ProductStatusInactive, storedProduct.Status()) // New products start inactive

	// Verify Outbox Event
	events := getOutboxEvents(ctx, t, client, productID)
	require.Len(t, events, 1)
	assert.Equal(t, "product.created", events[0].EventType)
}

func TestDiscountApplicationFlow(t *testing.T) {
	ctx := context.Background()
	client := getSpannerClient(ctx, t)
	defer client.Close()

	// Dependencies
	clk := &testClock{now: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)}
	comm := &spannerCommitter{client: client}
	productRepo := repo.NewProductRepo(client, clk)
	outboxRepo := repo.NewOutboxRepo()

	createUc := create_product.New(productRepo, outboxRepo, comm, clk)
	discountUc := apply_discount.New(productRepo, outboxRepo, comm, clk)

	// 1. Create Product
	productID := uuid.NewString()
	basePrice, _ := domain.NewMoneyFromFraction(200, 1) // 200.00
	_, err := createUc.Execute(ctx, create_product.Request{
		ID:        productID,
		Name:      "Discountable Product",
		Category:  "books",
		BasePrice: basePrice,
	})
	require.NoError(t, err)

	// 2. Activate Product
	activateUc := activate_product.New(productRepo, outboxRepo, comm, clk)
	err = activateUc.Execute(ctx, activate_product.Request{ProductID: productID})
	require.NoError(t, err)

	// 3. Apply Discount
	// 50% discount
	discountPercent := big.NewRat(50, 100)
	err = discountUc.Execute(ctx, apply_discount.Request{
		ProductID: productID,
		Percent:   discountPercent,
		Start:     clk.Now().Add(-1 * time.Hour), // Started 1 hour ago
		End:       clk.Now().Add(24 * time.Hour), // Ends tomorrow
	})
	require.NoError(t, err)

	p, err := productRepo.GetByID(ctx, productID)
	require.NoError(t, err)

	require.NotNil(t, p.Discount())
	assert.Equal(t, "1/2", p.Discount().Percent().String())
}

// Helper to query outbox events
type outboxEvent struct {
	EventID   string
	EventType string
	Payload   string
}

func getOutboxEvents(ctx context.Context, t *testing.T, client *spanner.Client, aggregateID string) []outboxEvent {
	stmt := spanner.NewStatement("SELECT event_id, event_type, payload FROM outbox_events WHERE aggregate_id = @id")
	stmt.Params["id"] = aggregateID
	iter := client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var events []outboxEvent
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		require.NoError(t, err)

		var e outboxEvent
		var payloadJSON spanner.NullJSON
		if err := row.Columns(&e.EventID, &e.EventType, &payloadJSON); err != nil {
			t.Fatal(err)
		}
		e.Payload = string(payloadJSON.String())
		events = append(events, e)
	}
	return events
}

func TestProductActivationDeactivation(t *testing.T) {
	ctx := context.Background()
	client := getSpannerClient(ctx, t)
	defer client.Close()

	// Dependencies
	clk := &testClock{now: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)}
	comm := &spannerCommitter{client: client}
	productRepo := repo.NewProductRepo(client, clk)
	outboxRepo := repo.NewOutboxRepo()

	createUc := create_product.New(productRepo, outboxRepo, comm, clk)
	activateUc := activate_product.New(productRepo, outboxRepo, comm, clk)

	// Create product
	productID := uuid.NewString()
	basePrice, _ := domain.NewMoneyFromFraction(100, 1)
	_, err := createUc.Execute(ctx, create_product.Request{
		ID:        productID,
		Name:      "Test Product",
		Category:  "test",
		BasePrice: basePrice,
	})
	require.NoError(t, err)

	// Verify initial status is inactive
	p, _ := productRepo.GetByID(ctx, productID)
	assert.Equal(t, domain.ProductStatusInactive, p.Status())

	// Activate product
	err = activateUc.Execute(ctx, activate_product.Request{ProductID: productID})
	require.NoError(t, err)

	// Verify status is now active
	p, _ = productRepo.GetByID(ctx, productID)
	assert.Equal(t, domain.ProductStatusActive, p.Status())

	// Verify activation event was created
	events := getOutboxEvents(ctx, t, client, productID)
	require.Len(t, events, 2) // created + activated
	assert.Equal(t, "product.activated", events[1].EventType)
}

func TestBusinessRuleValidation(t *testing.T) {
	ctx := context.Background()
	client := getSpannerClient(ctx, t)
	defer client.Close()

	// Dependencies
	clk := &testClock{now: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)}
	comm := &spannerCommitter{client: client}
	productRepo := repo.NewProductRepo(client, clk)
	outboxRepo := repo.NewOutboxRepo()

	createUc := create_product.New(productRepo, outboxRepo, comm, clk)
	discountUc := apply_discount.New(productRepo, outboxRepo, comm, clk)

	// Create inactive product
	productID := uuid.NewString()
	basePrice, _ := domain.NewMoneyFromFraction(100, 1)
	_, err := createUc.Execute(ctx, create_product.Request{
		ID:        productID,
		Name:      "Test Product",
		Category:  "test",
		BasePrice: basePrice,
	})
	require.NoError(t, err)

	// Try to apply discount to inactive product - should fail
	discountPercent := big.NewRat(10, 100)
	err = discountUc.Execute(ctx, apply_discount.Request{
		ProductID: productID,
		Percent:   discountPercent,
		Start:     clk.Now().Add(-1 * time.Hour),
		End:       clk.Now().Add(24 * time.Hour),
	})

	// Verify error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not active")
}
