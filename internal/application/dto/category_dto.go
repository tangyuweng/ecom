package dto

import "time"

type CreateCategoryRequest struct {
	Name string `json:"name" binding:"required"`
	Slug string `json:"slug" binding:"required"`
}

type UpdateCategoryRequest struct {
	ID   string `json:"id" binding:"required"`
	Name string `json:"name" binding:"required"`
	Slug string `json:"slug" binding:"required"`
}

type CategoryResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CategoryListResponse struct {
	Total      int                 `json:"total"`
	Categories []*CategoryResponse `json:"categories"`
}
