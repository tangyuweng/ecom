package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tangyuweng/ecom/internal/application/dto"
	"github.com/tangyuweng/ecom/internal/application/usecase"
	"github.com/tangyuweng/ecom/internal/domain/entity"
	"github.com/tangyuweng/ecom/internal/presentation/http/middleware"
)

type CartHandler struct {
	cartUseCase *usecase.CartUseCase
}

func NewCartHandler(cartUseCase *usecase.CartUseCase) *CartHandler {
	return &CartHandler{cartUseCase: cartUseCase}
}

// GetCart godoc
// @Summary      獲取購物車
// @Description  獲取當前使用者的購物車資訊
// @Tags         cart
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} dto.StandardResponse{data=dto.CartResponse} "獲取成功"
// @Failure      401 {object} dto.StandardResponse "未授權"
// @Failure      500 {object} dto.StandardResponse "伺服器錯誤"
// @Router       /cart [get]
func (h *CartHandler) GetCart(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("unauthorized"))
		return
	}

	response, err := h.cartUseCase.GetCart(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(response))
}

// AddToCart godoc
// @Summary      添加商品到購物車
// @Description  將商品添加到購物車，如果商品已存在則增加數量
// @Tags         cart
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.AddToCartRequest true "商品資訊"
// @Success      200 {object} dto.StandardResponse{data=dto.CartResponse} "添加成功"
// @Failure      400 {object} dto.StandardResponse "請求參數錯誤"
// @Failure      401 {object} dto.StandardResponse "未授權"
// @Failure      404 {object} dto.StandardResponse "商品不存在"
// @Failure      409 {object} dto.StandardResponse "庫存不足"
// @Failure      500 {object} dto.StandardResponse "伺服器錯誤"
// @Router       /cart/items [post]
func (h *CartHandler) AddToCart(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("unauthorized"))
		return
	}

	var req dto.AddToCartRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	response, err := h.cartUseCase.AddToCart(c.Request.Context(), userID, req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch err {
		case entity.ErrProductNotFound:
			statusCode = http.StatusNotFound
		case entity.ErrProductInsufficientStock:
			statusCode = http.StatusConflict
		case entity.ErrInvalidQuantity:
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(response))
}

// UpdateCartItem godoc
// @Summary      更新購物車項目
// @Description  更新購物車中商品的數量
// @Tags         cart
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "購物車項目 ID"
// @Param        request body dto.UpdateCartItemRequest true "數量資訊"
// @Success      200 {object} dto.StandardResponse{data=dto.CartResponse} "更新成功"
// @Failure      400 {object} dto.StandardResponse "請求參數錯誤"
// @Failure      401 {object} dto.StandardResponse "未授權"
// @Failure      403 {object} dto.StandardResponse "無權限"
// @Failure      404 {object} dto.StandardResponse "購物車項目不存在"
// @Failure      409 {object} dto.StandardResponse "庫存不足"
// @Failure      500 {object} dto.StandardResponse "伺服器錯誤"
// @Router       /cart/items/{id} [put]
func (h *CartHandler) UpdateCartItem(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("unauthorized"))
		return
	}

	cartItemID := c.Param("id")

	var req dto.UpdateCartItemRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	response, err := h.cartUseCase.UpdateCartItem(c.Request.Context(), userID, cartItemID, req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch err {
		case entity.ErrCartItemNotFound:
			statusCode = http.StatusNotFound
		case entity.ErrCartUnauthorized:
			statusCode = http.StatusForbidden
		case entity.ErrProductInsufficientStock:
			statusCode = http.StatusConflict
		case entity.ErrInvalidQuantity:
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(response))
}

// RemoveFromCart godoc
// @Summary      移除購物車項目
// @Description  從購物車中移除指定商品
// @Tags         cart
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "購物車項目 ID"
// @Success      204 "移除成功"
// @Failure      401 {object} dto.StandardResponse "未授權"
// @Failure      403 {object} dto.StandardResponse "無權限"
// @Failure      404 {object} dto.StandardResponse "購物車項目不存在"
// @Failure      500 {object} dto.StandardResponse "伺服器錯誤"
// @Router       /cart/items/{id} [delete]
func (h *CartHandler) RemoveFromCart(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("unauthorized"))
		return
	}

	cartItemID := c.Param("id")

	err := h.cartUseCase.RemoveFromCart(c.Request.Context(), userID, cartItemID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch err {
		case entity.ErrCartItemNotFound:
			statusCode = http.StatusNotFound
		case entity.ErrCartUnauthorized:
			statusCode = http.StatusForbidden
		}

		c.JSON(statusCode, dto.ErrorResponse(err.Error()))
		return
	}

	c.Status(http.StatusNoContent)
}

// ClearCart godoc
// @Summary      清空購物車
// @Description  清空當前使用者的購物車
// @Tags         cart
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      204 "清空成功"
// @Failure      401 {object} dto.StandardResponse "未授權"
// @Failure      404 {object} dto.StandardResponse "購物車不存在"
// @Failure      500 {object} dto.StandardResponse "伺服器錯誤"
// @Router       /cart [delete]
func (h *CartHandler) ClearCart(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("unauthorized"))
		return
	}

	err := h.cartUseCase.ClearCart(c.Request.Context(), userID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch err {
		case entity.ErrCartNotFound:
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, dto.ErrorResponse(err.Error()))
		return
	}

	c.Status(http.StatusNoContent)
}
