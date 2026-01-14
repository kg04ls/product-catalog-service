package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"cloud.google.com/go/spanner"
	"google.golang.org/grpc"

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
	"product-catalog-service/internal/infra/spannerx"
	"product-catalog-service/internal/pkg/clock"
	"product-catalog-service/internal/pkg/committer"
	pb "product-catalog-service/proto/product/v1"
)

type comm struct{ c *spanner.Client }

func (s *comm) Apply(ctx context.Context, p *committer.Plan) error {
	_, err := s.c.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		var ms []*spanner.Mutation
		for _, m := range p.Mutations() {
			if sm, ok := m.(spannerx.Mutation); ok {
				ms = append(ms, sm.M)
			}
		}
		return txn.BufferWrite(ms)
	})
	return err
}

func main() {
	p, i, d := os.Getenv("SPANNER_PROJECT_ID"), os.Getenv("SPANNER_INSTANCE_ID"), os.Getenv("SPANNER_DATABASE_ID")
	if p == "" || i == "" || d == "" {
		log.Fatal("Env vars missing")
	}

	c, err := spanner.NewClient(context.Background(), fmt.Sprintf("projects/%s/instances/%s/databases/%s", p, i, d))
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	ck, cm := clock.System{}, &comm{c}
	pr, or := repo.NewProductRepo(c, ck), repo.NewOutboxRepo(ck)
	rm := repo.NewSpannerReadModel(c, ck)

	h := product.NewHandler(
		create_product.New(pr, or, cm, ck), update_product.New(pr, or, cm, ck),
		activate_product.New(pr, or, cm, ck), deactivate_product.New(pr, or, cm, ck),
		archive_product.New(pr, or, cm, ck), apply_discount.New(pr, or, cm, ck),
		remove_discount.New(pr, or, cm, ck), get_product.New(rm), list_products.New(rm),
	)

	s := grpc.NewServer()
	pb.RegisterProductServiceServer(s, h)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Listening on :%s", port)
	if err := s.Serve(l); err != nil {
		log.Fatal(err)
	}
}
