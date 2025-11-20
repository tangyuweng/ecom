package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tangyuweng/ecom/internal/application/dto"
	"github.com/tangyuweng/ecom/internal/application/usecase"
	"github.com/tangyuweng/ecom/internal/domain/entity"
	"github.com/tangyuweng/ecom/internal/presentation/http/middleware"
)

type OrderHandler struct {
	orderUseCase *usecase.OrderUseCase
}

func NewOrderHandler(orderUseCase *usecase.OrderUseCase) *OrderHandler {
	return &OrderHandler{orderUseCase: orderUseCase}
}

// CreateOrderFromCart godoc
// @Summary      從購物車建立訂單
// @Description  將購物車中的商品轉換為訂單（結帳）
// @Tags         orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.CreateOrderRequest true "訂單資訊"
// @Success      201 {object} dto.StandardResponse{data=dto.OrderResponse} "訂單建立成功"
// @Failure      400 {object} dto.StandardResponse "請求參數錯誤"
// @Failure      401 {object} dto.StandardResponse "未授權"
// @Failure      404 {object} dto.StandardResponse "購物車為空或商品不存在"
// @Failure      409 {object} dto.StandardResponse "庫存不足"
// @Failure      500 {object} dto.StandardResponse "伺服器錯誤"
// @Router       /orders [post]
func (h *OrderHandler) CreateOrderFromCart(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("unauthorized"))
		return
	}

	var req dto.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	response, err := h.orderUseCase.CreateOrderFromCart(c.Request.Context(), userID, req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch err {
		case entity.ErrEmptyCart:
			statusCode = http.StatusBadRequest
		case entity.ErrProductNotFound:
			statusCode = http.StatusNotFound
		case entity.ErrProductInsufficientStock:
			statusCode = http.StatusConflict
		case entity.ErrOrderShippingAddressRequired, entity.ErrOrderRecipientNameRequired:
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(response))
}

// GetOrder godoc
// @Summary      取得訂單詳情
// @Description  取得指定訂單的詳細資訊
// @Tags         orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "訂單 ID"
// @Success      200 {object} dto.StandardResponse{data=dto.OrderResponse} "取得成功"
// @Failure      401 {object} dto.StandardResponse "未授權"
// @Failure      403 {object} dto.StandardResponse "無權限訪問此訂單"
// @Failure      404 {object} dto.StandardResponse "訂單不存在"
// @Failure      500 {object} dto.StandardResponse "伺服器錯誤"
// @Router       /orders/{id} [get]
func (h *OrderHandler) GetOrder(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("unauthorized"))
		return
	}

	orderID := c.Param("id")

	response, err := h.orderUseCase.GetOrder(c.Request.Context(), userID, orderID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch err {
		case entity.ErrOrderNotFound:
			statusCode = http.StatusNotFound
		case entity.ErrOrderUnauthorized:
			statusCode = http.StatusForbidden
		}

		c.JSON(statusCode, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(response))
}

// GetUserOrders godoc
// @Summary      取得用戶所有訂單
// @Description  取得當前使用者的所有訂單列表
// @Tags         orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} dto.StandardResponse{data=[]dto.OrderResponse} "取得成功"
// @Failure      401 {object} dto.StandardResponse "未授權"
// @Failure      500 {object} dto.StandardResponse "伺服器錯誤"
// @Router       /orders [get]
func (h *OrderHandler) GetUserOrders(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("unauthorized"))
		return
	}

	response, err := h.orderUseCase.GetUserOrders(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(response))
}

// CancelOrder godoc
// @Summary      取消訂單
// @Description  取消指定的訂單（僅限 Pending 或 Processing 狀態）
// @Tags         orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "訂單 ID"
// @Success      200 {object} dto.StandardResponse{data=dto.OrderResponse} "取消成功"
// @Failure      400 {object} dto.StandardResponse "訂單無法取消"
// @Failure      401 {object} dto.StandardResponse "未授權"
// @Failure      403 {object} dto.StandardResponse "無權限"
// @Failure      404 {object} dto.StandardResponse "訂單不存在"
// @Failure      500 {object} dto.StandardResponse "伺服器錯誤"
// @Router       /orders/{id}/cancel [post]
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("unauthorized"))
		return
	}

	orderID := c.Param("id")

	response, err := h.orderUseCase.CancelOrder(c.Request.Context(), userID, orderID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch err {
		case entity.ErrOrderNotFound:
			statusCode = http.StatusNotFound
		case entity.ErrOrderUnauthorized:
			statusCode = http.StatusForbidden
		case entity.ErrOrderCannotCancel:
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(response))
}

// UpdateOrderStatus godoc
// @Summary      更新訂單狀態
// @Description  管理員更新訂單狀態，已完成或已取消的訂單不可更新(僅管理員)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "訂單 ID"
// @Param        request body dto.UpdateOrderStatusRequest true "新狀態"
// @Success      200 {object} dto.StandardResponse{data=dto.OrderResponse} "更新成功"
// @Failure      400 {object} dto.StandardResponse "請求參數錯誤或訂單狀態不允許更新"
// @Failure      401 {object} dto.StandardResponse "未授權"
// @Failure      403 {object} dto.StandardResponse "無管理員權限"
// @Failure      404 {object} dto.StandardResponse "訂單不存在"
// @Failure      500 {object} dto.StandardResponse "伺服器錯誤"
// @Router       /admin/orders/{id}/status [patch]
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	adminID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("unauthorized"))
		return
	}

	orderID := c.Param("id")

	var req dto.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	response, err := h.orderUseCase.UpdateOrderStatus(c.Request.Context(), adminID, orderID, req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch err {
		case entity.ErrOrderNotFound:
			statusCode = http.StatusNotFound
		case entity.ErrUserUnauthorized:
			statusCode = http.StatusForbidden
		case entity.ErrOrderInvalidStatusTransition, entity.ErrOrderInvalidStatus:
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(response))
}
