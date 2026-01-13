package productpb

import (
	context "context"

	grpc "google.golang.org/grpc"
)

type CreateProductRequest struct {
	Name                 string
	Description          string
	Category             string
	BasePriceNumerator   int64
	BasePriceDenominator int64
}

type CreateProductReply struct {
	ProductId string
}

type UpdateProductRequest struct {
	ProductId   string
	Name        string
	Description string
	Category    string
}

type UpdateProductReply struct{}

type ActivateProductRequest struct {
	ProductId string
}

type ActivateProductReply struct{}

type DeactivateProductRequest struct {
	ProductId string
}

type DeactivateProductReply struct{}

type ApplyDiscountRequest struct {
	ProductId          string
	PercentNumerator   int64
	PercentDenominator int64
	StartTimestamp     int64
	EndTimestamp       int64
}

type ApplyDiscountReply struct{}

type RemoveDiscountRequest struct {
	ProductId string
}

type RemoveDiscountReply struct{}

type GetProductRequest struct {
	ProductId string
}

type GetProductReply struct {
	ProductId            string
	Name                 string
	Description          string
	Category             string
	BasePriceNumerator   int64
	BasePriceDenominator int64
	Status               string
	Discount             *Discount
}

type Discount struct {
	PercentNumerator   int64
	PercentDenominator int64
	StartTimestamp     int64
	EndTimestamp       int64
}

type ListProductsRequest struct {
	Category  string
	PageSize  int32
	PageToken string
}

type ListProductsReply struct {
	Products      []*ProductInfo
	NextPageToken string
}

type ProductInfo struct {
	ProductId string
	Name      string
	Category  string
	Status    string
}

type ProductServiceServer interface {
	CreateProduct(context.Context, *CreateProductRequest) (*CreateProductReply, error)
	UpdateProduct(context.Context, *UpdateProductRequest) (*UpdateProductReply, error)
	ActivateProduct(context.Context, *ActivateProductRequest) (*ActivateProductReply, error)
	DeactivateProduct(context.Context, *DeactivateProductRequest) (*DeactivateProductReply, error)
	ApplyDiscount(context.Context, *ApplyDiscountRequest) (*ApplyDiscountReply, error)
	RemoveDiscount(context.Context, *RemoveDiscountRequest) (*RemoveDiscountReply, error)
	GetProduct(context.Context, *GetProductRequest) (*GetProductReply, error)
	ListProducts(context.Context, *ListProductsRequest) (*ListProductsReply, error)
}

type UnimplementedProductServiceServer struct{}

func (UnimplementedProductServiceServer) CreateProduct(context.Context, *CreateProductRequest) (*CreateProductReply, error) {
	return nil, nil
}

func (UnimplementedProductServiceServer) UpdateProduct(context.Context, *UpdateProductRequest) (*UpdateProductReply, error) {
	return nil, nil
}

func (UnimplementedProductServiceServer) ActivateProduct(context.Context, *ActivateProductRequest) (*ActivateProductReply, error) {
	return nil, nil
}

func (UnimplementedProductServiceServer) DeactivateProduct(context.Context, *DeactivateProductRequest) (*DeactivateProductReply, error) {
	return nil, nil
}

func (UnimplementedProductServiceServer) ApplyDiscount(context.Context, *ApplyDiscountRequest) (*ApplyDiscountReply, error) {
	return nil, nil
}

func (UnimplementedProductServiceServer) RemoveDiscount(context.Context, *RemoveDiscountRequest) (*RemoveDiscountReply, error) {
	return nil, nil
}

func (UnimplementedProductServiceServer) GetProduct(context.Context, *GetProductRequest) (*GetProductReply, error) {
	return nil, nil
}

func (UnimplementedProductServiceServer) ListProducts(context.Context, *ListProductsRequest) (*ListProductsReply, error) {
	return nil, nil
}

func RegisterProductServiceServer(s *grpc.Server, srv ProductServiceServer) {
	s.RegisterService(&_ProductService_serviceDesc, srv)
}

var _ProductService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "product.v1.ProductService",
	HandlerType: (*ProductServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams:     []grpc.StreamDesc{},
	Metadata:    "proto/product/v1/product_service.proto",
}
