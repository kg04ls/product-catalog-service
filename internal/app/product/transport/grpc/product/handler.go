package product

import (
	"context"
	"math/big"
	"time"

	"github.com/google/uuid"

	"product-catalog-service/internal/app/product/contracts"
	"product-catalog-service/internal/app/product/domain"
	"product-catalog-service/internal/app/product/queries/get_product"
	"product-catalog-service/internal/app/product/queries/list_products"
	"product-catalog-service/internal/app/product/usecases/activate_product"
	"product-catalog-service/internal/app/product/usecases/apply_discount"
	"product-catalog-service/internal/app/product/usecases/archive_product"
	"product-catalog-service/internal/app/product/usecases/create_product"
	"product-catalog-service/internal/app/product/usecases/deactivate_product"
	"product-catalog-service/internal/app/product/usecases/remove_discount"
	"product-catalog-service/internal/app/product/usecases/update_product"
	pb "product-catalog-service/proto/product/v1"
)

type Handler struct {
	pb.UnimplementedProductServiceServer
	cUC  *create_product.Interactor
	uUC  *update_product.Interactor
	aUC  *activate_product.Interactor
	dUC  *deactivate_product.Interactor
	rUC  *archive_product.Interactor
	adUC *apply_discount.Interactor
	rdUC *remove_discount.Interactor
	gpQ  *get_product.Query
	lpQ  *list_products.Query
}

func NewHandler(c *create_product.Interactor, u *update_product.Interactor, a *activate_product.Interactor, d *deactivate_product.Interactor, r *archive_product.Interactor, ad *apply_discount.Interactor, rd *remove_discount.Interactor, gp *get_product.Query, lp *list_products.Query) *Handler {
	return &Handler{cUC: c, uUC: u, aUC: a, dUC: d, rUC: r, adUC: ad, rdUC: rd, gpQ: gp, lpQ: lp}
}

func (h *Handler) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.CreateProductReply, error) {
	bp, err := domain.NewMoneyFromFraction(req.BasePriceNumerator, req.BasePriceDenominator)
	if err != nil {
		return nil, err
	}
	id, err := h.cUC.Execute(ctx, create_product.Request{ID: uuid.NewString(), Name: req.Name, Description: req.Description, Category: req.Category, BasePrice: bp})
	if err != nil {
		return nil, err
	}
	return &pb.CreateProductReply{ProductId: id}, nil
}

func (h *Handler) UpdateProduct(ctx context.Context, req *pb.UpdateProductRequest) (*pb.UpdateProductReply, error) {
	return &pb.UpdateProductReply{}, h.uUC.Execute(ctx, update_product.Request{ProductID: req.ProductId, Name: req.Name, Description: req.Description, Category: req.Category})
}

func (h *Handler) ActivateProduct(ctx context.Context, req *pb.ActivateProductRequest) (*pb.ActivateProductReply, error) {
	return &pb.ActivateProductReply{}, h.aUC.Execute(ctx, activate_product.Request{ProductID: req.ProductId})
}

func (h *Handler) DeactivateProduct(ctx context.Context, req *pb.DeactivateProductRequest) (*pb.DeactivateProductReply, error) {
	return &pb.DeactivateProductReply{}, h.dUC.Execute(ctx, deactivate_product.Request{ProductID: req.ProductId})
}

func (h *Handler) ApplyDiscount(ctx context.Context, req *pb.ApplyDiscountRequest) (*pb.ApplyDiscountReply, error) {
	return &pb.ApplyDiscountReply{}, h.adUC.Execute(ctx, apply_discount.Request{ProductID: req.ProductId, Percent: big.NewRat(req.PercentNumerator, req.PercentDenominator), Start: time.Unix(req.StartTimestamp, 0), End: time.Unix(req.EndTimestamp, 0)})
}

func (h *Handler) RemoveDiscount(ctx context.Context, req *pb.RemoveDiscountRequest) (*pb.RemoveDiscountReply, error) {
	return &pb.RemoveDiscountReply{}, h.rdUC.Execute(ctx, remove_discount.Request{ProductID: req.ProductId})
}

func (h *Handler) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.GetProductReply, error) {
	d, err := h.gpQ.Execute(ctx, req.ProductId)
	if err != nil {
		return nil, err
	}
	var disc *pb.Discount
	if d.DiscountPct != "" {
		p, _ := new(big.Rat).SetString(d.DiscountPct)
		s, _ := time.Parse(time.RFC3339, d.DiscountStart)
		e, _ := time.Parse(time.RFC3339, d.DiscountEnd)
		disc = &pb.Discount{PercentNumerator: p.Num().Int64(), PercentDenominator: p.Denom().Int64(), StartTimestamp: s.Unix(), EndTimestamp: e.Unix()}
	}
	return &pb.GetProductReply{ProductId: d.ID, Name: d.Name, Description: d.Description, Category: d.Category, BasePriceNumerator: d.BasePriceNum, BasePriceDenominator: d.BasePriceDen, Status: d.Status, Discount: disc}, nil
}

func (h *Handler) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsReply, error) {
	r, err := h.lpQ.Execute(ctx, contracts.ListProductsFilter{Category: req.Category, OnlyActive: true, Limit: req.PageSize, PageToken: req.PageToken})
	if err != nil {
		return nil, err
	}
	var ps []*pb.ProductInfo
	for _, i := range r.Items {
		ps = append(ps, &pb.ProductInfo{ProductId: i.ID, Name: i.Name, Category: i.Category, Status: i.Status})
	}
	return &pb.ListProductsReply{Products: ps, NextPageToken: r.NextPageToken}, nil
}
