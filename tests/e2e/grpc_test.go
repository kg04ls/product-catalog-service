package e2e

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"product-catalog-service/internal/app/product/queries/get_product"
	"product-catalog-service/internal/app/product/queries/list_products"
	"product-catalog-service/internal/app/product/repo"
	"product-catalog-service/internal/app/product/transport/grpc/product"
	"product-catalog-service/internal/app/product/usecases/activate_product"
	"product-catalog-service/internal/app/product/usecases/apply_discount"
	"product-catalog-service/internal/app/product/usecases/archive_product"
	"product-catalog-service/internal/app/product/usecases/create_product"
	"product-catalog-service/internal/app/product/usecases/deactivate_product"
	"product-catalog-service/internal/app/product/usecases/remove_discount"
	"product-catalog-service/internal/app/product/usecases/update_product"
	pb "product-catalog-service/proto/product/v1"
)

func TestGRPC_CreateProduct(t *testing.T) {
	ctx := context.Background()
	client := getSpannerClient(ctx, t)
	defer client.Close()

	// Dependencies
	clk := &testClock{now: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)}
	comm := &spannerCommitter{client: client}
	productRepo := repo.NewProductRepo(client, clk)
	outboxRepo := repo.NewOutboxRepo(clk)

	createUC := create_product.New(productRepo, outboxRepo, comm, clk)
	updateUC := update_product.New(productRepo, outboxRepo, comm, clk)
	activateUC := activate_product.New(productRepo, outboxRepo, comm, clk)
	deactivateUC := deactivate_product.New(productRepo, outboxRepo, comm, clk)
	archiveUC := archive_product.New(productRepo, outboxRepo, comm, clk)
	applyDiscUC := apply_discount.New(productRepo, outboxRepo, comm, clk)
	removeDiscUC := remove_discount.New(productRepo, outboxRepo, comm, clk)

	readModel := repo.NewSpannerReadModel(client, clk)
	getProdQ := get_product.New(readModel)
	listProdsQ := list_products.New(readModel)

	handler := product.NewHandler(
		createUC, updateUC, activateUC, deactivateUC, archiveUC,
		applyDiscUC, removeDiscUC, getProdQ, listProdsQ,
	)

	// Start gRPC server on random port
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	grpcServer := grpc.NewServer()
	pb.RegisterProductServiceServer(grpcServer, handler)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			fmt.Printf("gRPC server error: %v\n", err)
		}
	}()
	defer grpcServer.Stop()

	// gRPC Client
	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	grpcClient := pb.NewProductServiceClient(conn)

	// Test Create
	resp, err := grpcClient.CreateProduct(ctx, &pb.CreateProductRequest{
		Name:                 "GRPC Test Product",
		Description:          "Test Description",
		Category:             "electronics",
		BasePriceNumerator:   500,
		BasePriceDenominator: 1,
	})
	require.NoError(t, err)
	require.NotEmpty(t, resp.ProductId)

	// Verify via gRPC GetProduct
	getResp, err := grpcClient.GetProduct(ctx, &pb.GetProductRequest{ProductId: resp.ProductId})
	require.NoError(t, err)
	require.Equal(t, "GRPC Test Product", getResp.Name)
	require.Equal(t, "inactive", getResp.Status)
}
