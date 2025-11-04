package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tangyuweng/ecom/internal/application/dto"
	"github.com/tangyuweng/ecom/internal/application/usecase"
	"github.com/tangyuweng/ecom/internal/domain/entity"
	"github.com/tangyuweng/ecom/internal/presentation/http/middleware"
)

type ProductHandler struct {
	productUseCase *usecase.ProductUseCase
}

func NewProductHandler(productUseCase *usecase.ProductUseCase) *ProductHandler {
	return &ProductHandler{productUseCase: productUseCase}
}

// GetProducts godoc
// @Summary      獲取商品列表
// @Description  根據查詢條件獲取商品列表,支援分頁、排序和篩選
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        name query string false "商品名稱"
// @Param        category query string false "類別 slug"
// @Param        min_price query number false "最低價格"
// @Param        max_price query number false "最高價格"
// @Param        is_active query boolean false "是否啟用"
// @Param        sort_by query string false "排序欄位" Enums(price, created_at, name)
// @Param        sort_order query string false "排序方向" Enums(asc, desc)
// @Param        page query int false "頁碼" default(1)
// @Param        page_size query int false "每頁數量" default(20)
// @Success      200 {object} dto.StandardResponse{data=dto.ProductListResponse} "獲取成功"
// @Failure      400 {object} dto.StandardResponse "請求參數錯誤"
// @Failure      500 {object} dto.StandardResponse "伺服器錯誤"
// @Router       /products [get]
func (h *ProductHandler) GetProducts(c *gin.Context) {
	var req dto.ProductListRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	response, err := h.productUseCase.GetProducts(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(response))
}

// GetProduct godoc
// @Summary      獲取單一商品
// @Description  根據商品 ID 獲取商品詳細資訊
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        id path string true "商品 ID"
// @Success      200 {object} dto.StandardResponse{data=dto.ProductResponse} "獲取成功"
// @Failure      404 {object} dto.StandardResponse "商品不存在"
// @Failure      500 {object} dto.StandardResponse "伺服器錯誤"
// @Router       /products/{id} [get]
func (h *ProductHandler) GetProduct(c *gin.Context) {
	productID := c.Param("id")

	response, err := h.productUseCase.GetProduct(c.Request.Context(), productID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch err {
		case entity.ErrProductNotFound:
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(response))
}

// CreateProduct godoc
// @Summary      創建商品
// @Description  創建新商品 (需要管理員權限)
// @Tags         products
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.CreateProductRequest true "商品資訊"
// @Success      201 {object} dto.StandardResponse{data=dto.ProductResponse} "創建成功"
// @Failure      400 {object} dto.StandardResponse "請求參數錯誤"
// @Failure      401 {object} dto.StandardResponse "未授權"
// @Failure      403 {object} dto.StandardResponse "無權限"
// @Failure      500 {object} dto.StandardResponse "伺服器錯誤"
// @Router       /products [post]
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("unauthorized"))
		return
	}

	var req dto.CreateProductRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	response, err := h.productUseCase.CreateProduct(c.Request.Context(), userID, req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch err {
		case entity.ErrProductUnauthorized:
			statusCode = http.StatusForbidden
		case entity.ErrCategoryNotFound:
			statusCode = http.StatusBadRequest
		case entity.ErrProductInvalidPrice, entity.ErrProductInvalidStock, entity.ErrProductNameRequired:
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(response))
}

// UpdateProduct godoc
// @Summary      更新商品
// @Description  更新商品資訊 (需要管理員權限)
// @Tags         products
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "商品 ID"
// @Param        request body dto.UpdateProductRequest true "商品資訊"
// @Success      200 {object} dto.StandardResponse{data=dto.ProductResponse} "更新成功"
// @Failure      400 {object} dto.StandardResponse "請求參數錯誤"
// @Failure      401 {object} dto.StandardResponse "未授權"
// @Failure      403 {object} dto.StandardResponse "無權限"
// @Failure      404 {object} dto.StandardResponse "商品不存在"
// @Failure      500 {object} dto.StandardResponse "伺服器錯誤"
// @Router       /products/{id} [put]
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("unauthorized"))
		return
	}

	productID := c.Param("id")

	var req dto.UpdateProductRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	response, err := h.productUseCase.UpdateProduct(c.Request.Context(), userID, productID, req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch err {
		case entity.ErrProductUnauthorized:
			statusCode = http.StatusForbidden
		case entity.ErrProductNotFound:
			statusCode = http.StatusNotFound
		case entity.ErrProductInvalidPrice, entity.ErrProductInvalidStock, entity.ErrProductNameRequired:
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(response))
}

// DeleteProduct godoc
// @Summary      刪除商品
// @Description  刪除商品 (需要管理員權限)
// @Tags         products
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "商品 ID"
// @Success      204 "刪除成功"
// @Failure      401 {object} dto.StandardResponse "未授權"
// @Failure      403 {object} dto.StandardResponse "無權限"
// @Failure      404 {object} dto.StandardResponse "商品不存在"
// @Failure      500 {object} dto.StandardResponse "伺服器錯誤"
// @Router       /products/{id} [delete]
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("unauthorized"))
		return
	}

	productID := c.Param("id")

	err := h.productUseCase.DeleteProduct(c.Request.Context(), userID, productID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch err {
		case entity.ErrProductUnauthorized:
			statusCode = http.StatusForbidden
		case entity.ErrProductNotFound:
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, dto.ErrorResponse(err.Error()))
		return
	}

	c.Status(http.StatusNoContent)
}

// UpdateProductStock godoc
// @Summary      更新商品庫存
// @Description  更新商品庫存數量 (需要管理員權限)
// @Tags         products
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "商品 ID"
// @Param        request body dto.UpdateProductStockRequest true "庫存資訊"
// @Success      200 {object} dto.StandardResponse{data=dto.ProductResponse} "更新成功"
// @Failure      400 {object} dto.StandardResponse "請求參數錯誤"
// @Failure      401 {object} dto.StandardResponse "未授權"
// @Failure      403 {object} dto.StandardResponse "無權限"
// @Failure      404 {object} dto.StandardResponse "商品不存在"
// @Failure      500 {object} dto.StandardResponse "伺服器錯誤"
// @Router       /products/{id}/stock [patch]
func (h *ProductHandler) UpdateProductStock(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("unauthorized"))
		return
	}

	productID := c.Param("id")

	var req dto.UpdateProductStockRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	response, err := h.productUseCase.UpdateProductStock(c.Request.Context(), userID, productID, req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch err {
		case entity.ErrProductUnauthorized:
			statusCode = http.StatusForbidden
		case entity.ErrProductNotFound:
			statusCode = http.StatusNotFound
		case entity.ErrProductInvalidStock:
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(response))
}

// GetProductsByCategory godoc
// @Summary      根據類別獲取商品列表
// @Description  根據類別 ID 獲取該類別下的所有商品
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        id path string true "類別 ID"
// @Success      200 {object} dto.StandardResponse{data=[]dto.ProductResponse} "獲取成功"
// @Failure      404 {object} dto.StandardResponse "類別不存在"
// @Failure      500 {object} dto.StandardResponse "伺服器錯誤"
// @Router       /categories/{id}/products [get]
func (h *ProductHandler) GetProductsByCategory(c *gin.Context) {
	categoryID := c.Param("id")

	response, err := h.productUseCase.GetProductsByCategory(c.Request.Context(), categoryID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch err {
		case entity.ErrCategoryNotFound:
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(response))
}
