package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tangyuweng/ecom/internal/application/dto"
	"github.com/tangyuweng/ecom/internal/application/usecase"
	"github.com/tangyuweng/ecom/internal/domain/entity"
	"github.com/tangyuweng/ecom/internal/presentation/http/middleware"
)

type CategoryHandler struct {
	categoryUseCase *usecase.CategoryUseCase
}

func NewCategoryHandler(categoryUseCase *usecase.CategoryUseCase) *CategoryHandler {
	return &CategoryHandler{categoryUseCase: categoryUseCase}
}

// GetCategoryList godoc
// @Summary      獲取所有類別
// @Description  獲取所有類別列表
// @Tags         categories
// @Accept       json
// @Produce      json
// @Success      200 {object} dto.StandardResponse{data=dto.CategoryListResponse} "獲取成功"
// @Failure      500 {object} dto.StandardResponse "伺服器錯誤"
// @Router       /categories [get]
func (h *CategoryHandler) GetCategoryList(c *gin.Context) {
	response, err := h.categoryUseCase.FindAllCategory(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(response))
}

// GetCategory godoc
// @Summary      獲得單個類別
// @Description  根據 ID 獲得類別詳情
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        id path string true "類別 ID"
// @Success      200 {object} dto.StandardResponse{data=dto.CategoryResponse} "獲取成功"
// @Failure      404 {object} dto.StandardResponse "找不到"
// @Failure      500 {object} dto.StandardResponse "伺服器錯誤"
// @Router       /categories/{id} [get]
func (h *CategoryHandler) GetCategory(c *gin.Context) {
	categoryID := c.Param("id")

	response, err := h.categoryUseCase.FindByID(c.Request.Context(), categoryID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(response))
}

// CreateCategory godoc
// @Summary      創建新類別
// @Description  創建新的商品類別(僅管理員)
// @Tags         categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.CreateCategoryRequest true "類別資訊"
// @Success      201 {object} dto.StandardResponse{data=dto.CategoryResponse} "創建成功"
// @Failure      400 {object} dto.StandardResponse "請求參數錯誤"
// @Failure      401 {object} dto.StandardResponse "未授權"
// @Failure      403 {object} dto.StandardResponse "權限不足"
// @Failure      500 {object} dto.StandardResponse "伺服器錯誤"
// @Router       /categories [post]
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("unauthorized"))
		return
	}

	var req dto.CreateCategoryRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	response, err := h.categoryUseCase.CreateCategory(c.Request.Context(), userID, req)
	if err != nil {
		statusCode := http.StatusBadRequest
		if err == entity.ErrCategoryUnauthorized {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(response))
}

// CreateCategory godoc
// @Summary      更新指定類別
// @Description  根據 id 更新商品類別(僅管理員)
// @Tags         categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.UpdateCategoryRequest true "類別資訊"
// @Success      200 {object} dto.StandardResponse{data=dto.CategoryResponse} "創建成功"
// @Failure      400 {object} dto.StandardResponse "請求參數錯誤"
// @Failure      401 {object} dto.StandardResponse "未授權"
// @Failure      403 {object} dto.StandardResponse "權限不足"
// @Failure      500 {object} dto.StandardResponse "伺服器錯誤"
// @Router       /categories/:id [put]
func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("unauthorized"))
		return
	}

	var req dto.UpdateCategoryRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	response, err := h.categoryUseCase.UpdateCategory(c.Request.Context(), userID, req.ID, req)
	if err != nil {
		statusCode := http.StatusBadRequest
		if err == entity.ErrCategoryUnauthorized {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(response))
}

// DeleteCategory godoc
// @Summary      刪除指定類別
// @Description  根據 ID 刪除商品類別(僅管理員)
// @Tags         categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "類別 ID"
// @Success      204 "刪除成功"
// @Failure      400 {object} dto.StandardResponse "請求參數錯誤"
// @Failure      401 {object} dto.StandardResponse "未授權"
// @Failure      403 {object} dto.StandardResponse "權限不足"
// @Failure      500 {object} dto.StandardResponse "伺服器錯誤"
// @Router       /categories/{id} [delete]
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("unauthorized"))
		return
	}

	categoryID := c.Param("id")

	err := h.categoryUseCase.DeleteCategory(c.Request.Context(), userID, categoryID)
	if err != nil {
		statusCode := http.StatusBadRequest
		if err == entity.ErrCategoryUnauthorized {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, dto.ErrorResponse(err.Error()))
		return
	}

	c.Status(http.StatusNoContent)
}
