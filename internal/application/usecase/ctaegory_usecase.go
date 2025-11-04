package usecase

import (
	"context"

	"github.com/tangyuweng/ecom/internal/application/dto"
	"github.com/tangyuweng/ecom/internal/domain/entity"
	"github.com/tangyuweng/ecom/internal/domain/repository"
)

type CategoryUseCase struct {
	categoryRepo repository.CategoryRepository
	userRepo     repository.UserRepository
	productRepo  repository.ProductRepository
}

func NewCategoryUseCase(
	categoryRepo repository.CategoryRepository,
	userRepo repository.UserRepository,
	productRepo repository.ProductRepository,
) *CategoryUseCase {
	return &CategoryUseCase{
		categoryRepo: categoryRepo,
		userRepo:     userRepo,
		productRepo:  productRepo,
	}
}

func (uc *CategoryUseCase) FindAllCategory(ctx context.Context) (*dto.CategoryListResponse, error) {
	categories, err := uc.categoryRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	categoryResponses := make([]*dto.CategoryResponse, len(categories))
	for i, category := range categories {
		categoryResponses[i] = uc.toCategoryResponse(category)
	}

	return &dto.CategoryListResponse{
		Total:      len(categoryResponses),
		Categories: categoryResponses,
	}, nil
}

func (uc *CategoryUseCase) FindByID(ctx context.Context, id string) (*dto.CategoryResponse, error) {
	category, err := uc.categoryRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return uc.toCategoryResponse(category), nil
}

func (uc *CategoryUseCase) CreateCategory(ctx context.Context, userID string, req dto.CreateCategoryRequest) (*dto.CategoryResponse, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if !user.IsAdmin() {
		return nil, entity.ErrCategoryUnauthorized
	}

	exists, err := uc.categoryRepo.ExistsByName(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, entity.ErrCategoryExists
	}

	category, err := entity.NewCategory(req.Name, req.Slug)
	if err != nil {
		return nil, err
	}

	if err := uc.categoryRepo.Create(ctx, category); err != nil {
		return nil, err
	}

	return uc.toCategoryResponse(category), nil
}

func (uc *CategoryUseCase) UpdateCategory(ctx context.Context, userID, categoryID string, req dto.UpdateCategoryRequest) (*dto.CategoryResponse, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if !user.IsAdmin() {
		return nil, entity.ErrCategoryUnauthorized
	}

	category, err := uc.categoryRepo.FindByID(ctx, categoryID)
	if err != nil {
		return nil, err
	}

	if err := category.Update(req.Name, req.Slug); err != nil {
		return nil, err
	}

	if err := uc.categoryRepo.Update(ctx, category); err != nil {
		return nil, err
	}

	return uc.toCategoryResponse(category), nil
}

func (uc *CategoryUseCase) DeleteCategory(ctx context.Context, userID, categoryID string) error {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	if !user.IsAdmin() {
		return entity.ErrCategoryUnauthorized
	}

	_, err = uc.categoryRepo.FindByID(ctx, categoryID)
	if err != nil {
		return err
	}

	hashProducts, err := uc.productRepo.ExistsByCategoryID(ctx, categoryID)
	if err != nil {
		return err
	}

	if hashProducts {
		return entity.ErrCategoryHasProducts
	}

	return uc.categoryRepo.Delete(ctx, categoryID)
}

func (uc *CategoryUseCase) toCategoryResponse(category *entity.Category) *dto.CategoryResponse {
	return &dto.CategoryResponse{
		ID:        category.ID,
		Name:      category.Name,
		Slug:      category.Slug,
		CreatedAt: category.CreatedAt,
		UpdatedAt: category.UpdatedAt,
	}
}
