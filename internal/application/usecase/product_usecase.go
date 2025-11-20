package usecase

import (
	"context"

	"github.com/tangyuweng/ecom/internal/application/dto"
	"github.com/tangyuweng/ecom/internal/domain/entity"
	"github.com/tangyuweng/ecom/internal/domain/repository"
)

type ProductUseCase struct {
	productRepo  repository.ProductRepository
	categoryRepo repository.CategoryRepository
	userRepo     repository.UserRepository
}

func NewProductUseCase(
	productRepo repository.ProductRepository,
	categoryRepo repository.CategoryRepository,
	userRepo repository.UserRepository,
) *ProductUseCase {
	return &ProductUseCase{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
		userRepo:     userRepo,
	}
}

func (uc *ProductUseCase) GetProducts(ctx context.Context, req dto.ProductListRequest) (*dto.ProductListResponse, error) {
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}
	if req.SortBy == "" {
		req.SortBy = "created_at"
	}
	if req.SortOrder == "" {
		req.SortOrder = "desc"
	}

	query := &entity.ProductQuery{
		Name:      req.Name,
		Category:  req.Category,
		MinPrice:  req.MinPrice,
		MaxPrice:  req.MaxPrice,
		IsActive:  req.IsActive,
		SortBy:    req.SortBy,
		SortOrder: req.SortOrder,
		Page:      req.Page,
		PageSize:  req.PageSize,
	}

	products, total, err := uc.productRepo.FindByQuery(ctx, query)
	if err != nil {
		return nil, err
	}

	productResponses := make([]*dto.ProductResponse, len(products))
	for i, product := range products {
		productResponses[i] = uc.toProductResponse(product)
	}

	return &dto.ProductListResponse{
		Total:    total,
		Products: productResponses,
	}, nil
}

func (uc *ProductUseCase) GetProduct(ctx context.Context, id string) (*dto.ProductResponse, error) {
	product, err := uc.productRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return uc.toProductResponse(product), nil
}

func (uc *ProductUseCase) GetProductsByCategory(ctx context.Context, categoryID string) (*dto.ProductListResponse, error) {
	_, err := uc.categoryRepo.FindByID(ctx, categoryID)
	if err != nil {
		return nil, err
	}

	products, err := uc.productRepo.FindByCategoryID(ctx, categoryID)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.ProductResponse, len(products))
	for i, product := range products {
		responses[i] = uc.toProductResponse(product)
	}

	return &dto.ProductListResponse{
		Total:    len(responses),
		Products: responses,
	}, nil
}

func (uc *ProductUseCase) CreateProduct(ctx context.Context, userID string, req dto.CreateProductRequest) (*dto.ProductResponse, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if !user.IsAdmin() {
		return nil, entity.ErrProductUnauthorized
	}

	product, err := entity.NewProduct(
		req.CategoryID,
		req.Name,
		req.Description,
		req.Price,
		req.StockQuantity,
		req.IsActive,
	)
	if err != nil {
		return nil, err
	}

	if err := uc.productRepo.Create(ctx, product); err != nil {
		return nil, err
	}

	createdProduct, err := uc.productRepo.FindByID(ctx, product.ID)
	if err != nil {
		return nil, err
	}

	return uc.toProductResponse(createdProduct), nil
}

func (uc *ProductUseCase) UpdateProduct(ctx context.Context, userID, productID string, req dto.UpdateProductRequest) (*dto.ProductResponse, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if !user.IsAdmin() {
		return nil, entity.ErrProductUnauthorized
	}

	product, err := uc.productRepo.FindByID(ctx, productID)
	if err != nil {
		return nil, err
	}

	if err := product.Update(
		req.CategoryID,
		req.Name,
		req.Description,
		req.Price,
		req.StockQuantity,
		req.IsActive,
	); err != nil {
		return nil, err
	}

	if err := uc.productRepo.Update(ctx, product); err != nil {
		return nil, err
	}

	updatedProduct, err := uc.productRepo.FindByID(ctx, productID)
	if err != nil {
		return nil, err
	}

	return uc.toProductResponse(updatedProduct), nil
}

func (uc *ProductUseCase) DeleteProduct(ctx context.Context, userID, productID string) error {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	if !user.IsAdmin() {
		return entity.ErrProductUnauthorized
	}

	_, err = uc.productRepo.FindByID(ctx, productID)
	if err != nil {
		return err
	}

	return uc.productRepo.Delete(ctx, productID)
}

func (uc *ProductUseCase) UpdateProductStock(ctx context.Context, userID, productID string, req dto.UpdateProductStockRequest) (*dto.ProductResponse, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if !user.IsAdmin() {
		return nil, entity.ErrProductUnauthorized
	}

	product, err := uc.productRepo.FindByID(ctx, productID)
	if err != nil {
		return nil, err
	}

	if err := product.UpdateStockQuantity(req.StockQuantity); err != nil {
		return nil, err
	}

	if err := uc.productRepo.Update(ctx, product); err != nil {
		return nil, err
	}

	updatedProduct, err := uc.productRepo.FindByID(ctx, productID)
	if err != nil {
		return nil, err
	}

	return uc.toProductResponse(updatedProduct), nil
}

func (uc *ProductUseCase) toProductResponse(product *entity.Product) *dto.ProductResponse {
	return &dto.ProductResponse{
		ID:            product.ID,
		Name:          product.Name,
		Description:   product.Description,
		Price:         product.Price,
		StockQuantity: product.StockQuantity,
		IsActive:      product.IsActive,
		CategoryID:    product.CategoryID,
		CategoryName:  product.CategoryName,
		CategorySlug:  product.CategorySlug,
		CreatedAt:     product.CreatedAt,
		UpdatedAt:     product.UpdatedAt,
	}
}
